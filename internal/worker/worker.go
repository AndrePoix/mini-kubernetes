package worker

import (
	"context"
	"fmt"
	"log"
	"mini-kubernetes/pkg"
	"os/signal"
	"syscall"
	"time"
)

type Worker struct {
	node         *Node
	masterClient *MasterClient
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewWorker(masterAddr string, nodeInfo pkg.NodeInfo) (*Worker, error) {
	ctx, cancel := context.WithCancel(context.Background())

	node, err := NewNode(ctx, nodeInfo)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create node: %v", err)
	}

	masterClient, err := NewMasterClient(masterAddr, time.Second*5, node)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create master client: %v", err)
	}

	return &Worker{
		node:         node,
		masterClient: masterClient,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (w *Worker) Start() error {

	sigCtx, sigStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigStop()

	log.Printf("starting worker node: %s", w.node.Name)

	go w.node.nodeAgent()

	go w.masterClient.StartHeartbeat(w.ctx)

	<-sigCtx.Done()

	return w.shutdown()
}

func (w *Worker) shutdown() error {

	// Cancel context to stop all goroutines
	w.cancel()

	log.Println("cleaning up containers")
	w.node.cleanupContainers()

	if err := w.masterClient.Close(); err != nil {
		log.Printf("error closing master client: %v", err)
	}

	// Give some time for cleanup
	time.Sleep(2 * time.Second)

	log.Println("Cleanup complete, exiting")
	return nil
}

func (w *Worker) Stop() error {
	return w.shutdown()
}
