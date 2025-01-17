package ebpf

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/vishvananda/netlink"
)

type Collector struct {
	program           *ebpf.Program
	packetCountMap    *ebpf.Map
	latencyMap        *ebpf.Map
	dropMap           *ebpf.Map
	packetSizeMap     *ebpf.Map    
	connectionMap     *ebpf.Map    
	protocolCountMap  *ebpf.Map
	link              link.Link
}

type Connection struct {
	SourceIP string
	DestIP   string
}

type ConnectionInfo struct {
    SourceIP   string
    DestIP     string
    SourcePort uint16
    DestPort   uint16
    Protocol   uint8
}

func NewCollector() (*Collector, error) {
	// Load pre-compiled eBPF program
	spec, err := ebpf.LoadCollectionSpec("ebpf/monitor.o")
	if err != nil {
		return nil, fmt.Errorf("failed to load eBPF program: %v", err)
	}

	var objs struct {
		MonitorPackets *ebpf.Program `ebpf:"monitor_packets"`
		PacketCount    *ebpf.Map     `ebpf:"packet_count"`
		LatencyMap     *ebpf.Map     `ebpf:"latency_map"`
		DropMap        *ebpf.Map     `ebpf:"drop_map"`
		PacketSize       *ebpf.Map     `ebpf:"packet_size"`      
		ConnectionMap    *ebpf.Map     `ebpf:"connection_map"`   
		ProtocolCount    *ebpf.Map     `ebpf:"protocol_count"`   
	}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load eBPF objects: %v", err)
	}

	return &Collector{
		program:        objs.MonitorPackets,
		packetCountMap: objs.PacketCount,
		latencyMap:     objs.LatencyMap,
		dropMap:        objs.DropMap,
		packetSizeMap:     objs.PacketSize,      
		connectionMap:     objs.ConnectionMap,   
		protocolCountMap:  objs.ProtocolCount,
	}, nil
}

func (c *Collector) GetPacketCounts() (map[string]map[string]uint64, error) {
	counts := make(map[string]map[string]uint64)
	var key struct{ SrcIP, DstIP uint32 }
	var value uint64

	entries := c.packetCountMap.Iterate()
	for entries.Next(&key, &value) {
		srcIP := int2ip(key.SrcIP).String()
		dstIP := int2ip(key.DstIP).String()
		if _, ok := counts[srcIP]; !ok {
			counts[srcIP] = make(map[string]uint64)
		}
		counts[srcIP][dstIP] = value
	}

	return counts, nil
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func (c *Collector) GetPacketSizes() (map[string]map[string]uint64, error) {
    sizes := make(map[string]map[string]uint64)
    var key struct{ SrcIP, DstIP uint32 }
    var value uint64

    entries := c.packetSizeMap.Iterate()
    for entries.Next(&key, &value) {
        srcIP := int2ip(key.SrcIP).String()
        dstIP := int2ip(key.DstIP).String()
        if _, ok := sizes[srcIP]; !ok {
            sizes[srcIP] = make(map[string]uint64)
        }
        sizes[srcIP][dstIP] = value
    }

    return sizes, nil
}

func (c *Collector) GetConnections() (map[ConnectionInfo]uint64, error) {
    connections := make(map[ConnectionInfo]uint64)
    var key struct {
        SrcIP, DstIP uint32
        SrcPort, DstPort uint16
        Protocol uint8
    }
    var value uint64

    entries := c.connectionMap.Iterate()
    for entries.Next(&key, &value) {
        connInfo := ConnectionInfo{
            SourceIP:   int2ip(key.SrcIP).String(),
            DestIP:     int2ip(key.DstIP).String(),
            SourcePort: ntohs(key.SrcPort),
            DestPort:   ntohs(key.DstPort),
            Protocol:   key.Protocol,
        }
        connections[connInfo] = value
    }

    return connections, nil
}

func ntohs(n uint16) uint16 {
    return (n<<8)&0xff00 | (n>>8)&0x00ff
}

func (c *Collector) GetProtocolCounts() (map[string]uint64, error) {
    counts := make(map[string]uint64)
    var value uint64

    if err := c.protocolCountMap.Lookup(uint32(0), &value); err == nil {
        counts["TCP"] = value
    }
    if err := c.protocolCountMap.Lookup(uint32(1), &value); err == nil {
        counts["UDP"] = value
    }

    return counts, nil
}

func (c *Collector) GetLatencies() (map[string]map[string]float64, error) {
    latencies := make(map[string]map[string]float64)
    var key struct{ SrcIP, DstIP uint32 }
    var value uint64

    entries := c.latencyMap.Iterate()
    for entries.Next(&key, &value) {
        srcIP := int2ip(key.SrcIP).String()
        dstIP := int2ip(key.DstIP).String()
        if _, ok := latencies[srcIP]; !ok {
            latencies[srcIP] = make(map[string]float64)
        }
        latencies[srcIP][dstIP] = float64(value) / 1000000 // Convert ns to ms
    }

    return latencies, nil
}

func (c *Collector) GetPacketDrops() (map[string]uint64, error) {
    drops := make(map[string]uint64)
    var key uint32
    var value uint64

    entries := c.dropMap.Iterate()
    for entries.Next(&key, &value) {
        reason := getDropReason(key)
        drops[reason] = value
    }

    return drops, nil
}

func getDropReason(code uint32) string {
    reasons := map[uint32]string{
        1: "Generic drop",
        2: "Invalid IP header",
        3: "TCP checksum error",
        4: "UDP checksum error",
    }
    if reason, ok := reasons[code]; ok {
        return reason
    }
    return fmt.Sprintf("Unknown (%d)", code)
}

// Stop detaches the eBPF program from the network interface
func (c *Collector) Stop() error {
	if c.link != nil {
		return c.link.Close()
	}
	return nil
}

// Start attaches the eBPF program to the network interface
func (c *Collector) Start() error {
	iface, err := netlink.LinkByName("eth0") // Change to your interface name
	if err != nil {
		return fmt.Errorf("failed to get interface: %v", err)
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   c.program,
		Interface: iface.Attrs().Index,
	})
	if err != nil {
		return fmt.Errorf("failed to attach XDP program: %v", err)
	}

	c.link = l
	log.Println("eBPF program attached successfully")
	return nil
}
