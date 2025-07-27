package master

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

type Master struct {
	NodeManager *NodeManager
	GRPCServer  *GRPCServer
	interval    time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewMaster() *Master {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	nodeManager := NewNodeManager()
	grpcServer := NewGRPCServer(nodeManager, "8080")

	return &Master{
		NodeManager: nodeManager,
		GRPCServer:  grpcServer,
		interval:    time.Second * 4,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (m *Master) Start() {

	if err := m.GRPCServer.Start(); err != nil {
		log.Fatalf("failed to start gRPC server: %v", err)
	}

	go m.Scheduler()

	log.Println("master server started")

	// Wait for shutdown signal
	<-m.ctx.Done()
	log.Println("shutting down master server")

	m.GRPCServer.Stop()
}

func (m *Master) Scheduler() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Println("Starting scheduler...")

	for {
		select {
		case <-ticker.C:
			m.NodeManager.SchedulePods()
		case <-m.ctx.Done():
			log.Println("Scheduler stopping...")
			return
		}
	}
}

func (m *Master) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}
