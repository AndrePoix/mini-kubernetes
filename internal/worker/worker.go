package worker

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"mini-kubernetes/pkg"
)

type Worker struct {
	node  *Node
	Pulse *Pulse
}

func (w *Worker) Start() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()
	if w.node == nil {
		w.node = &Node{}
	}
	if w.Pulse == nil {
		w.Pulse = &Pulse{}
	}
	w.node.initNodeClient()
	w.node.NodeInfo = pkg.NodeInfo{Name: "node1", TotalCPU: 2_000, TotalMem: 2_048_000_000} // 2 cpu and 2gb

	w.Pulse.newPulse(time.Second*5, w.node, "http://localhost:8080/")

	go w.Pulse.startHeartbeat(context.TODO())

	go w.node.nodeAgent()

	<-ctx.Done() // Wait for signal
	log.Println("Interrupt received, cleaning up containers...")

	w.node.cleanupContainers()

	log.Println("Cleanup complete, exiting")
}
