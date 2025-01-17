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
	packetCounts, err := collector.GetPacketCounts()
    if err != nil {
        return fmt.Errorf("failed to get packet counts: %v", err)
    }

    latencies, err := collector.GetLatencies()
    if err != nil {
        return fmt.Errorf("failed to get latencies: %v", err)
    }

    drops, err := collector.GetPacketDrops()
    if err != nil {
        return fmt.Errorf("failed to get packet drops: %v", err)
    }

    packetSizes, err := collector.GetPacketSizes()
    if err != nil {
        return fmt.Errorf("failed to get packet sizes: %v", err)
    }

    protocolCounts, err := collector.GetProtocolCounts()
    if err != nil {
        return fmt.Errorf("failed to get protocol counts: %v", err)
    }

	connections, err := collector.GetConnections()
    if err != nil {
        return fmt.Errorf("failed to get connections: %v", err)
    }

    fmt.Println("Network Traffic Summary:")
    for srcIP, dests := range packetCounts {
        srcResource, _ := correlateWithKubernetes(kubeClient, srcIP)
        for dstIP, count := range dests {
            dstResource, _ := correlateWithKubernetes(kubeClient, dstIP)
            latency := latencies[srcIP][dstIP]
            bytes := packetSizes[srcIP][dstIP]
            fmt.Printf("  %s -> %s: %d packets, %d bytes, %.2f ms avg latency\n", 
                       srcResource, dstResource, count, bytes, latency)
            exporter.AddNetworkTraffic(srcIP, dstIP, float64(count))
            exporter.ObserveConnectionLatency(srcIP, dstIP, latency)
        }
    }

	fmt.Println("Detailed Connections:")
    for connInfo, count := range connections {
        srcResource, _ := correlateWithKubernetes(kubeClient, connInfo.SourceIP)
        dstResource, _ := correlateWithKubernetes(kubeClient, connInfo.DestIP)
        fmt.Printf("  %s:%d -> %s:%d (%s): %d packets\n",
            srcResource, connInfo.SourcePort,
            dstResource, connInfo.DestPort,
            protocolToString(connInfo.Protocol), count)
    }

    if len(drops) > 0 {
        fmt.Println("Packet Drops:")
        for reason, count := range drops {
            fmt.Printf("  %s: %d\n", reason, count)
            exporter.IncrementPacketDrops(reason)
        }
    }

    printSummaryStats(packetCounts, packetSizes, protocolCounts)

	return nil
}

func protocolToString(protocol uint8) string {
	switch protocol {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return fmt.Sprintf("Unknown (%d)", protocol)
	}
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


func correlateWithKubernetes(kubeClient *kubernetes.Client, ip string) (string, error) {
    pod, err := kubeClient.GetPodByIP(ip)
    if err == nil {
        return fmt.Sprintf("%s (Pod: %s)", ip, pod), nil
    }

    service, err := kubeClient.GetServiceByIP(ip)
    if err == nil {
        return fmt.Sprintf("%s (Service: %s)", ip, service), nil
    }

    return ip, nil
}