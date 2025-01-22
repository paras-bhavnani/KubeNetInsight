package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paras-bhavnani/KubeNetInsight/pkg/ebpf"
	"github.com/paras-bhavnani/KubeNetInsight/pkg/kubernetes"
	"github.com/paras-bhavnani/KubeNetInsight/pkg/metrics"
)

func main() {
	log.Println("Starting KubeNetInsight...")

	// Initialize eBPF collector
	collector, err := ebpf.NewCollector()
	if err != nil {
		log.Fatalf("Failed to initialize eBPF collector: %v", err)
	}

	// Initialize Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Initialize metrics exporter
	exporter, err := metrics.NewExporter()
	if err != nil {
		log.Fatalf("Failed to initialize metrics exporter: %v", err)
	}

	// Start the metrics server
	go exporter.StartServer("8080")

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start monitoring
	go func() {
		if err := startMonitoring(ctx, collector, kubeClient, exporter); err != nil {
			log.Printf("Monitoring stopped: %v", err)
			cancel()
		}
	}()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigCh
	log.Println("Shutting down KubeNetInsight...")
	cancel()
	time.Sleep(2 * time.Second) // Give some time for goroutines to clean up
}

func startMonitoring(ctx context.Context, collector *ebpf.Collector, kubeClient *kubernetes.Client, exporter *metrics.Exporter) error {
	// Start the eBPF collector
	if err := collector.Start(); err != nil {
		return err
	}
	defer collector.Stop()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Periodically fetch Kubernetes resources and update metrics
			if err := updateKubernetesMetrics(kubeClient, exporter); err != nil {
				log.Printf("Failed to update Kubernetes metrics: %v", err)
			}

			// Read and process eBPF data
			if err := processEBPFData(collector, kubeClient, exporter); err != nil {
				log.Printf("Failed to process eBPF data: %v", err)
			}
		}
	}
}

func updateKubernetesMetrics(kubeClient *kubernetes.Client, exporter *metrics.Exporter) error {
	namespaces, err := kubeClient.GetNamespaces()
	if err != nil {
		return fmt.Errorf("failed to get namespaces: %v", err)
	}

	for _, ns := range namespaces {
		pods, err := kubeClient.GetPods(ns)
		if err != nil {
			log.Printf("Failed to get pods for namespace %s: %v", ns, err)
			continue
		}
		exporter.UpdatePodCount(ns, len(pods))

		services, err := kubeClient.GetServices(ns)
		if err != nil {
			log.Printf("Failed to get services for namespace %s: %v", ns, err)
			continue
		}
		exporter.UpdateServiceCount(ns, len(services))
	}

	return nil
}

func processEBPFData(collector *ebpf.Collector, kubeClient *kubernetes.Client, exporter *metrics.Exporter) error {
	// Get consolidated stats
	packetStats, err := collector.GetPacketStats()
	if err != nil {
		return fmt.Errorf("failed to get packet stats: %v", err)
	}

	connStats, err := collector.GetConnectionStats()
	if err != nil {
		return fmt.Errorf("failed to get connection stats: %v", err)
	}

	// Get packet drops
	drops, err := collector.GetPacketDrops()
	if err != nil {
		return fmt.Errorf("failed to get packet drops: %v", err)
	}

	// Create maps for summary statistics
	packetCounts := make(map[string]map[string]uint64)
	bytesCounts := make(map[string]map[string]uint64)
	// protocolCounts := make(map[string]uint64)

	// Process packet statistics
	fmt.Println("Network Traffic Summary:")
	for _, stat := range packetStats {
		srcResource, srcNamespace, _ := correlateWithKubernetes(kubeClient, stat.Source)
		dstResource, dstNamespace, _ := correlateWithKubernetes(kubeClient, stat.Destination)

		// Update summary statistics maps
		if _, ok := packetCounts[stat.Source]; !ok {
			packetCounts[stat.Source] = make(map[string]uint64)
			bytesCounts[stat.Source] = make(map[string]uint64)
		}
		packetCounts[stat.Source][stat.Destination] = stat.Count
		bytesCounts[stat.Source][stat.Destination] = stat.Bytes

		fmt.Printf("  %s/%s -> %s/%s: %d packets, %d bytes, %s avg latency\n",
			srcNamespace, srcResource, dstNamespace, dstResource,
			stat.Count, stat.Bytes, formatLatency(stat.Latency))

		exporter.AddNetworkTraffic(stat.Source, stat.Destination, float64(stat.Count))
		exporter.ObserveConnectionLatency(stat.Source, stat.Destination, float64(stat.Latency))

		protocol := "unknown"
		for _, conn := range connStats {
			if conn.Source == stat.Source && conn.Destination == stat.Destination {
				protocol = conn.Protocol
				break
			}
		}

		exporter.ObservePacketSize(stat.Source, stat.Destination, protocol, float64(stat.Bytes/stat.Count))
	}

	protocolCounts, err := collector.GetProtocolCounts()
	if err != nil {
		return fmt.Errorf("failed to get protocol counts: %v", err)
	}

	// Process packet drops
	if len(drops) > 0 {
		fmt.Println("Packet Drops:")
		for reason, count := range drops {
			fmt.Printf("  %s: %d\n", reason, count)
			exporter.IncrementPacketDrops(reason)
		}
	}

	// Print summary statistics
	printSummaryStats(packetCounts, bytesCounts, protocolCounts)

	return nil
}

func printSummaryStats(packetCounts map[string]map[string]uint64, bytesCounts map[string]map[string]uint64, protocolCounts map[string]uint64) {
	var totalPackets, totalBytes uint64
	var uniqueSources, uniqueDestinations int
	sourcesSet := make(map[string]bool)
	destinationsSet := make(map[string]bool)

	for src, dests := range packetCounts {
		sourcesSet[src] = true
		for dst, count := range dests {
			destinationsSet[dst] = true
			totalPackets += count
			totalBytes += bytesCounts[src][dst]
		}
	}

	uniqueSources = len(sourcesSet)
	uniqueDestinations = len(destinationsSet)

	fmt.Println("Summary Statistics:")
	fmt.Printf("- Total Packets: %d\n", totalPackets)
	fmt.Printf("- Total Bytes: %d\n", totalBytes)
	fmt.Printf("- Unique Sources: %d\n", uniqueSources)
	fmt.Printf("- Unique Destinations: %d\n", uniqueDestinations)
	fmt.Println("- Protocol Breakdown:")
	for proto, count := range protocolCounts {
		fmt.Printf("  - %s: %d packets\n", proto, count)
	}
	fmt.Println("--------------------")
}

func correlateWithKubernetes(kubeClient *kubernetes.Client, ip string) (string, string, error) {
	pod, namespace, err := kubeClient.GetPodByIP(ip)
	if err == nil {
		return fmt.Sprintf("%s (Pod)", pod), namespace, nil
	}

	service, namespace, err := kubeClient.GetServiceByIP(ip)
	if err == nil {
		return fmt.Sprintf("%s (Service)", service), namespace, nil
	}

	return ip, "", nil
}

type ConnectionSummary struct {
	Source      string
	Destination string
	Protocol    string
	PacketCount uint64
}

func formatLatency(latencyNs uint64) string {
	if latencyNs < 1000 {
		return fmt.Sprintf("%.2f ns", float64(latencyNs))
	} else if latencyNs < 1000000 {
		return fmt.Sprintf("%.2f μs", float64(latencyNs)/1000)
	} else if latencyNs < 1000000000 {
		return fmt.Sprintf("%.2f ms", float64(latencyNs)/1000000)
	} else {
		return fmt.Sprintf("%.2f s", float64(latencyNs)/1000000000)
	}
}
