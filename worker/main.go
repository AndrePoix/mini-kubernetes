package worker

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	node  = &Node{Name: "node1", TotalCPU: 2_000, TotalMem: 2_048_000_000} // 2 cpu and 2gb
	pods  []*Pod
	mu    sync.Mutex
	pulse *Pulse
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	node.initNodeClient()
	pulse.newPulse(time.Second*5, node, "http://localhost:8080/")
	pulse.startHeartbeat(context.TODO())
	go node.nodeAgent()

	<-ctx.Done() // Wait for signal
	log.Println("Interrupt received, cleaning up containers...")

	node.cleanupContainers()

	log.Println("Cleanup complete, exiting")
}
