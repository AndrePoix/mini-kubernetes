package worker

import (
	"context"
	"sync"
	"time"
)

type Node struct {
	Name     string
	TotalCPU int // milliCPU (1000 = 1 CPU)
	TotalMem int // MB
	UsedCPU  int
	UsedMem  int
	mu       sync.Mutex
	cli      *Client
	Pods     []*Pod
	ctx      context.Context
}

func (n *Node) initNodeClient() {
	n.cli.initClient(n.ctx)
}

func (n *Node) nodeAgent() {
	for {
		// We copy to not lock the mutex for too long
		n.mu.Lock()
		podsCopy := make([]*Pod, len(n.Pods))
		copy(podsCopy, n.Pods)
		n.mu.Unlock()

		for _, pod := range podsCopy {

			switch pod.Phase {
			case Pending:
				n.cli.startContainer(pod)
			case Stopped:
				n.cli.deleteContainer(pod)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (n *Node) cleanupContainers() {
	mu.Lock()
	podsCopy := make([]*Pod, len(n.Pods))
	copy(podsCopy, n.Pods)
	mu.Unlock()

	for _, pod := range podsCopy {
		go n.cli.deleteContainer(pod)
	}
}

func (n *Node) addPods(pods []*Pod) {
	n.mu.Lock()
	n.Pods = append(n.Pods, pods...)
	n.mu.Unlock()
}
