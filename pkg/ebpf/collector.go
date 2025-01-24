package ebpf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/paras-bhavnani/KubeNetInsight/pkg/kubernetes"
	"github.com/vishvananda/netlink"
)

type Collector struct {
	program          *ebpf.Program
	packetCountMap   *ebpf.Map
	latencyMap       *ebpf.Map
	dropMap          *ebpf.Map
	packetSizeMap    *ebpf.Map
	connectionMap    *ebpf.Map
	protocolCountMap *ebpf.Map
	link             link.Link
	kubeClient       *kubernetes.Client
}

type ConnInfo struct {
	SrcIP    uint32
	DstIP    uint32
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
}

func (c *ConnInfo) UnmarshalBinary(data []byte) error {
	if len(data) < 13 {
		return errors.New("not enough data")
	}
	c.SrcIP = binary.LittleEndian.Uint32(data[0:4])
	c.DstIP = binary.LittleEndian.Uint32(data[4:8])
	c.SrcPort = binary.BigEndian.Uint16(data[8:10])
	c.DstPort = binary.BigEndian.Uint16(data[10:12])
	c.Protocol = data[12]
	return nil
}

type Connection struct {
	SourceIP string
	DestIP   string
}

type ConnectionInfo struct {
	SourceIP        string
	DestIP          string
	SourcePort      uint16
	DestPort        uint16
	Protocol        uint8
	SourceName      string
	SourceNamespace string
	DestName        string
	DestNamespace   string
}

type PacketStats struct {
	Source      string
	Destination string
	Protocol    string
	Size        uint64
	Count       uint64
	Latency     uint64
	Bytes       uint64
}

type ConnectionStats struct {
	Source      string
	Destination string
	State       string
	Count       uint64
	Protocol    string
	SourcePort  uint16
	DestPort    uint16
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
		PacketSize     *ebpf.Map     `ebpf:"packet_size"`
		ConnectionMap  *ebpf.Map     `ebpf:"connection_map"`
		ProtocolCount  *ebpf.Map     `ebpf:"protocol_count"`
	}

	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load eBPF objects: %v", err)
	}

	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return &Collector{
		program:          objs.MonitorPackets,
		packetCountMap:   objs.PacketCount,
		latencyMap:       objs.LatencyMap,
		dropMap:          objs.DropMap,
		packetSizeMap:    objs.PacketSize,
		connectionMap:    objs.ConnectionMap,
		protocolCountMap: objs.ProtocolCount,
		kubeClient:       kubeClient,
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
	binary.LittleEndian.PutUint32(ip, nn)
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
	var key ConnInfo
	var value uint64

	entries := c.connectionMap.Iterate()
	for entries.Next(&key, &value) {
		srcIP := int2ip(key.SrcIP).String()
		dstIP := int2ip(key.DstIP).String()

		// Look up pod or service for source IP
		srcName, srcNamespace, _ := c.kubeClient.GetPodByIP(srcIP)
		if srcName == "" {
			srcName, srcNamespace, _ = c.kubeClient.GetServiceByIP(srcIP)
		}

		// Look up pod or service for destination IP
		dstName, dstNamespace, _ := c.kubeClient.GetPodByIP(dstIP)
		if dstName == "" {
			dstName, dstNamespace, _ = c.kubeClient.GetServiceByIP(dstIP)
		}

		connInfo := ConnectionInfo{
			SourceIP:        srcIP,
			DestIP:          dstIP,
			SourcePort:      ntohs(key.SrcPort),
			DestPort:        ntohs(key.DstPort),
			Protocol:        key.Protocol,
			SourceName:      srcName,
			SourceNamespace: srcNamespace,
			DestName:        dstName,
			DestNamespace:   dstNamespace,
		}
		connections[connInfo] = value
	}
	if entries.Err() != nil {
		log.Printf("Error iterating connection_map: %v", entries.Err())
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

func (c *Collector) GetLatencies() (map[string]map[string]uint64, error) {
	latencies := make(map[string]map[string]uint64)
	var key struct{ SrcIP, DstIP uint32 }
	var value struct {
		TotalLatency uint64
		PacketCount  uint64
	}

	entries := c.latencyMap.Iterate()
	for entries.Next(&key, &value) {
		srcIP := int2ip(key.SrcIP).String()
		dstIP := int2ip(key.DstIP).String()
		if _, ok := latencies[srcIP]; !ok {
			latencies[srcIP] = make(map[string]uint64)
		}
		// latencies[srcIP][dstIP] = value // Store latency in nanoseconds / 1000000 // Convert ns to ms
		if value.PacketCount > 0 {
			avgLatency := value.TotalLatency / value.PacketCount
			latencies[srcIP][dstIP] = avgLatency
		}
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

func (c *Collector) GetPacketStats() ([]PacketStats, error) {
	var stats []PacketStats

	// Get all required metrics
	packetCounts, err := c.GetPacketCounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get packet counts: %v", err)
	}

	packetSizes, err := c.GetPacketSizes()
	if err != nil {
		return nil, fmt.Errorf("failed to get packet sizes: %v", err)
	}

	latencies, err := c.GetLatencies()
	if err != nil {
		return nil, fmt.Errorf("failed to get latencies: %v", err)
	}

	// Aggregate metrics
	for srcIP, dests := range packetCounts {
		for dstIP, count := range dests {
			stat := PacketStats{
				Source:      srcIP,
				Destination: dstIP,
				Count:       count,
				Size:        packetSizes[srcIP][dstIP],
				Latency:     latencies[srcIP][dstIP],
				Bytes:       packetSizes[srcIP][dstIP] * count,
			}
			stats = append(stats, stat)
		}
	}

	return stats, nil
}

func (c *Collector) GetConnectionStats() ([]ConnectionStats, error) {
	var stats []ConnectionStats

	connections, err := c.GetConnections()
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %v", err)
	}

	for connInfo, count := range connections {
		stat := ConnectionStats{
			Source:      connInfo.SourceIP,
			Destination: connInfo.DestIP,
			SourcePort:  connInfo.SourcePort,
			DestPort:    connInfo.DestPort,
			Protocol:    protocolToString(connInfo.Protocol),
			Count:       count,
			State:       determineConnectionState(connInfo),
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func determineConnectionState(connInfo ConnectionInfo) string {
	// Basic state determination based on protocol
	// You can enhance this based on your needs
	if connInfo.Protocol == 6 { // TCP
		return "ESTABLISHED" // Simplified - you might want to track actual TCP states
	}
	return "ACTIVE" // For UDP and others
}

func protocolToString(protocol uint8) string {
	switch protocol {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", protocol)
	}
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
