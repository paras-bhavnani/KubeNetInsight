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

	ticker := time.NewTicker(30 * time.Second)
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
			if err := processEBPFData(collector, exporter); err != nil {
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

func processEBPFData(collector *ebpf.Collector, exporter *metrics.Exporter) error {
	// Read packet count map
	packetCounts, err := collector.GetPacketCounts()
	if err != nil {
		return fmt.Errorf("failed to get packet counts: %v", err)
	}

	// Add this section to display metrics
    for sourceIP, destinations := range packetCounts {
        for destIP, count := range destinations {
            // Just use %s for string formatting of IPs
            log.Printf("Packets from %s to %s: %d", sourceIP, destIP, count)
            exporter.AddNetworkTraffic(sourceIP, destIP, float64(count)*1500)
        }
    }

	for sourceIP, destinations := range packetCounts {
		for destIP, count := range destinations {
			exporter.AddNetworkTraffic(sourceIP, destIP, float64(count)*1500)
		}
	}

	// Read connection latency map
	latencies, err := collector.GetConnectionLatencies()
	if err != nil {
		return fmt.Errorf("failed to get connection latencies: %v", err)
	}

	for conn, latency := range latencies {
		exporter.ObserveConnectionLatency(conn.SourceIP, conn.DestIP, float64(latency)/1000000000) // Convert ns to seconds
	}

	// Read packet drop map
	drops, err := collector.GetPacketDrops()
	if err != nil {
		return fmt.Errorf("failed to get packet drops: %v", err)
	}

	for reason, count := range drops {
		for i := 0; i < int(count); i++ {
			exporter.IncrementPacketDrops(reason)
		}
	}

	return nil
}
