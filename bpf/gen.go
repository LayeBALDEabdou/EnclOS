package bpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpfel -cc clang Bpf execve_tracker.c -- -I../bpf -g -O2
