CLANG ?= clang
CFLAGS := -O2 -g -Wall -target bpf -I/usr/include

EBPF_DIR := ebpf
GO_DIR := cmd/kubenetinsight

.PHONY: all clean

all: $(EBPF_DIR)/monitor.o $(GO_DIR)/kubenetinsight

$(EBPF_DIR)/monitor.o: $(EBPF_DIR)/monitor.c
	$(CLANG) $(CFLAGS) -c $< -o $@

$(GO_DIR)/kubenetinsight: $(GO_DIR)/main.go
	go build -o $@ $<

clean:
	rm -f $(EBPF_DIR)/*.o $(GO_DIR)/kubenetinsight
