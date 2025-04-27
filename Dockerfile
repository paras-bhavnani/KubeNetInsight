# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install eBPF dependencies
RUN apk add --no-cache clang llvm linux-headers libbpf-dev

# Build eBPF program
COPY ebpf/monitor.c .
RUN clang -O2 -g -Wall -target bpf -I/usr/include/ -c monitor.c -o monitor.o

# Build Go binary
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o /kubenetinsight ./cmd/kubenetinsight

# Final stage
FROM alpine:3.19
WORKDIR /app

# Install required tools for eBPF
RUN apk add --no-cache \
    libbpf-dev \
    iproute2 \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D -u 10001 kubenet

# Copy files with correct ownership
COPY --from=builder --chown=kubenet:kubenet /kubenetinsight /app/kubenetinsight
COPY --from=builder --chown=kubenet:kubenet /app/monitor.o /app/ebpf/

# Set permissions (now works because kubenet owns the files)
USER kubenet
RUN chmod 550 /app/kubenetinsight && \
    chmod 440 /app/ebpf/monitor.o

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD curl --fail http://localhost:8080/healthz || exit 1

CMD ["/app/kubenetinsight"]