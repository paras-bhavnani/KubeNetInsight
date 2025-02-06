from prometheus_api_client import PrometheusConnect
from datetime import datetime, timedelta

PROMETHEUS_URL = "http://prometheus.monitoring.svc.cluster.local:9090"
# PROMETHEUS_URL = "http://localhost:9090" # For testing purpose
prom = PrometheusConnect(url=PROMETHEUS_URL, disable_ssl=True)

def query_prometheus(promql_query):
    try:
        result = prom.custom_query(query=promql_query)
        return result
    except Exception as e:
        raise Exception(f"Prometheus Query Failed: {str(e)}")

def get_metric_range(metric_name, start_time, end_time):
    try:
        result = prom.get_metric_range_data(
            metric_name=metric_name,
            start_time=start_time,
            end_time=end_time
        )
        return result
    except Exception as e:
        raise Exception(f"Metric Range Query Failed: {str(e)}")

def get_average_latency():
    # Define the time window (e.g., 1 minute)
    time_window = "[1m]"

    # Query for the sum of latencies
    sum_query = f'rate(kubenetinsight_connection_latency_nano_seconds_sum{time_window})'
    sum_result = query_prometheus(sum_query)
    print("\nLatency Sum Rate:", sum_result)

    # Query for the count of observations
    count_query = f'rate(kubenetinsight_connection_latency_nano_seconds_count{time_window})'
    count_result = query_prometheus(count_query)
    print("\nLatency Count Rate:", count_result)

    # Calculate average latency
    query = f"""
        rate(kubenetinsight_connection_latency_nano_seconds_sum{time_window}) 
        / 
        rate(kubenetinsight_connection_latency_nano_seconds_count{time_window})
    """
    return query_prometheus(query)

def get_latency_percentiles():
    # Define the percentiles to calculate
    percentiles = [0.10, 0.50, 0.90, 0.99]  # 1st, 50th, 90th, and 99th percentiles

    # Query for each percentile
    results = {}
    for p in percentiles:
        query = f"""
            histogram_quantile({p}, 
                rate(kubenetinsight_connection_latency_nano_seconds_bucket[1m])
            )
        """
        results[p] = query_prometheus(query)

    return results

def format_latency(latency_ns):    
    if latency_ns < 1000:
        return f"{latency_ns:.2f} ns"
    elif latency_ns < 1000000:
        return f"{latency_ns/1000:.2f} μs"
    elif latency_ns < 1000000000:
        return f"{latency_ns/1000000:.2f} ms"
    else:
        return f"{latency_ns/1000000000:.2f} s"

def format_latency_stats(avg_latency, percentiles):
    print("\nLatency Statistics:")
    if avg_latency:
        for stat in avg_latency:
            source_ip = stat['metric']['source_ip']
            dest_ip = stat['metric']['destination_ip']
            latency = float(stat['value'][1])
            print(f"  {source_ip} -> {dest_ip}: {latency} {format_latency(latency)} (avg)")

    if percentiles:
        for p, result in percentiles.items():
            for stat in result:
                source_ip = stat['metric'].get('source_ip', 'all')
                dest_ip = stat['metric'].get('destination_ip', 'all')
                latency = float(stat['value'][1])
                print(f"  {source_ip} -> {dest_ip}: {format_latency(latency)} ({int(p*100)}th percentile)")

if __name__ == '__main__':
    # Test instant query
    pod_count = query_prometheus("kubenetinsight_pod_count")
    print("Pod Count:", pod_count)

    # Test range query for network traffic
    end_time = datetime.now()
    start_time = end_time - timedelta(hours=1)
    network_traffic = get_metric_range(
        "kubenetinsight_network_traffic_bytes",
        start_time,
        end_time
    )
    print("\nNetwork Traffic:", network_traffic)

    # Get latency metrics using rate and histogram_quantile
    avg_latency = get_average_latency()
    latency_percentiles = get_latency_percentiles()
    format_latency_stats(avg_latency, latency_percentiles)

    packet_drops = query_prometheus("kubenetinsight_packet_drops")
    print("\nPacket Drops:", packet_drops)

    # Query connection states
    conn_states = query_prometheus("kubenetinsight_connection_states")
    print("\nConnection States:", conn_states)

    # Query protocol traffic
    protocol_traffic = query_prometheus("kubenetinsight_protocol_traffic_total")
    print("\nProtocol Traffic:", protocol_traffic)