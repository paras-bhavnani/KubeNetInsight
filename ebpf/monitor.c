#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
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

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct ip_key);
    __type(value, __u64);
    __uint(max_entries, 1024);
} packet_size SEC(".maps");

struct conn_info {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 protocol;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct conn_info);
    __type(value, __u64);
    __uint(max_entries, 1024);
} connection_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 2);
} protocol_count SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct ip_key);
    __type(value, __u64);
    __uint(max_entries, 1024);
} packet_start_time SEC(".maps");

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

    // Add packet size tracking
    __u64 packet_len = (__u64)(data_end - data);
    __u64 *size = bpf_map_lookup_elem(&packet_size, &key);
    if (size) {
        __sync_fetch_and_add(size, packet_len);
    } else {
        bpf_map_update_elem(&packet_size, &key, &packet_len, BPF_ANY);
    }

    struct conn_info conn = {
        .src_ip = ip->saddr,
        .dst_ip = ip->daddr,
        .protocol = ip->protocol
    };

    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)(ip + 1);
        if ((void *)(tcp + 1) > data_end)
            return XDP_PASS;
        conn.src_port = tcp->source;
        conn.dst_port = tcp->dest;
    } else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)(ip + 1);
        if ((void *)(udp + 1) > data_end)
            return XDP_PASS;
        conn.src_port = udp->source;
        conn.dst_port = udp->dest;
    }

    __u64 *conn_count = bpf_map_lookup_elem(&connection_map, &conn);
    if (conn_count) {
        __sync_fetch_and_add(conn_count, 1);
    } else {
        __u64 initial = 1;
        bpf_map_update_elem(&connection_map, &conn, &initial, BPF_ANY);
    }

    __u32 proto_index;
    if (ip->protocol == IPPROTO_TCP) {
        proto_index = 0;
    } else if (ip->protocol == IPPROTO_UDP) {
        proto_index = 1;
    } else {
        return XDP_PASS;
    }

    __u64 *proto_count = bpf_map_lookup_elem(&protocol_count, &proto_index);
    if (proto_count) {
        __sync_fetch_and_add(proto_count, 1);
    }

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
    __u64 *start_time = bpf_map_lookup_elem(&packet_start_time, &key);
    if (start_time) {
        __u64 latency = *ts - *start_time;
        __u64 *total_latency = bpf_map_lookup_elem(&latency_map, &key);
        if (total_latency) {
            __sync_fetch_and_add(total_latency, latency);
        } else {
            bpf_map_update_elem(&latency_map, &key, &latency, BPF_ANY);
        }
        bpf_map_delete_elem(&packet_start_time, &key);
    } else {
        bpf_map_update_elem(&packet_start_time, &key, ts, BPF_ANY);
    }

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