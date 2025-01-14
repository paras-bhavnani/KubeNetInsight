#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

struct ip_key {
    __u32 src_ip;
    __u32 dst_ip;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct ip_key);
    __type(value, __u64);
    __uint(max_entries, 1024);
} packet_count SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct ip_key);
    __type(value, __u64);
    __uint(max_entries, 1024);
} latency_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 16);
} drop_map SEC(".maps");

static __always_inline int process_packet(struct xdp_md *ctx, __u64 *ts) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    struct ethhdr *eth = data;

    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;

    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return XDP_PASS;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;

    struct ip_key key = {
        .src_ip = ip->saddr,
        .dst_ip = ip->daddr
    };

    // Update packet count
    __u64 *count = bpf_map_lookup_elem(&packet_count, &key);
    if (count) {
        __sync_fetch_and_add(count, 1);
        bpf_printk("Packet captured: src_ip=%u, dst_ip=%u\n", key.src_ip, key.dst_ip);
    } else {
        __u64 initial = 1;
        bpf_map_update_elem(&packet_count, &key, &initial, BPF_ANY);
        bpf_printk("Packet captured: src_ip=%u, dst_ip=%u\n", key.src_ip, key.dst_ip);
    }

    // Update latency
    __u64 *latency = bpf_map_lookup_elem(&latency_map, &key);
    if (latency)
        *latency = *ts - *latency;
    else
        bpf_map_update_elem(&latency_map, &key, ts, BPF_ANY);

    return XDP_PASS;
}

SEC("xdp")
int monitor_packets(struct xdp_md *ctx) {
    __u64 ts = bpf_ktime_get_ns();
    int ret = process_packet(ctx, &ts);

    if (ret == XDP_DROP) {
        __u32 reason = 1; // Generic drop reason
        __u64 *drops = bpf_map_lookup_elem(&drop_map, &reason);
        if (drops)
            __sync_fetch_and_add(drops, 1);
        else {
            __u64 initial = 1;
            bpf_map_update_elem(&drop_map, &reason, &initial, BPF_ANY);
        }
    }

    return ret;
}

char _license[] SEC("license") = "GPL";