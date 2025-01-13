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
	program        *ebpf.Program
	packetCountMap *ebpf.Map
	latencyMap     *ebpf.Map
	dropMap        *ebpf.Map
	link           link.Link
}

type Connection struct {
	SourceIP string
	DestIP   string
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
	}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load eBPF objects: %v", err)
	}

	return &Collector{
		program:        objs.MonitorPackets,
		packetCountMap: objs.PacketCount,
		latencyMap:     objs.LatencyMap,
		dropMap:        objs.DropMap,
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

func (c *Collector) GetConnectionLatencies() (map[Connection]uint64, error) {
	latencies := make(map[Connection]uint64)
	var key struct{ SrcIP, DstIP uint32 }
	var value uint64

	entries := c.latencyMap.Iterate()
	for entries.Next(&key, &value) {
		conn := Connection{
			SourceIP: int2ip(key.SrcIP).String(),
			DestIP:   int2ip(key.DstIP).String(),
		}
		latencies[conn] = value
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
	// Map drop codes to reasons
	reasons := map[uint32]string{
		1: "Invalid IP header",
		2: "TCP checksum error",
		3: "UDP checksum error",
		// Add more reasons as needed
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
