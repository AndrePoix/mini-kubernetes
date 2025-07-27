package main

import (
	"mini-kubernetes/internal/worker"
	"mini-kubernetes/pkg"
)

func main() {
	nodeInfo := &pkg.NodeInfo{
		Name:     "worker-1",
		TotalCPU: 2_000,
		TotalMem: 2_000,
		UsedCPU:  0,
		UsedMem:  0,
	}
	w, _ := worker.NewWorker("localhost:8080", *nodeInfo)

	w.Start()
}
