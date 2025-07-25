package master

import (
	"context"
	"os/signal"
	"syscall"
	"time"
)

type Master struct {
	NodeManager *NodeManager
	Api         *Api
	interval    time.Duration
	ctx         context.Context
}

func (m *Master) Start() {
	m.interval = time.Second * 4
	m.ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGINT)
	m.NodeManager = NewNodeManager()
	m.Api.NodeManager = m.NodeManager
	m.Api.initListeningPort("8080")
	m.Api.setupRoutes()
	go m.Scheduler()

	<-m.ctx.Done()
}

func (m *Master) Scheduler() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.NodeManager.SchedulePods()
		case <-m.ctx.Done():
			return
		}
	}
}
