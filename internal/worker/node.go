package worker

import (
	"context"
	"sync"
	"time"

	"mini-kubernetes/pkg"
)

type Node struct {
	pkg.NodeInfo

	mu  sync.Mutex
	cli *Client
	ctx context.Context
}

func (n *Node) initNodeClient() {
	if n.cli == nil {
		n.cli = &Client{}
	}
	n.cli.initClient(n.ctx)
}

func (n *Node) nodeAgent() {
	for {
		// We copy to not lock the mutex for too long
		n.mu.Lock()
		podsCopy := make([]*pkg.Pod, len(n.Pods))
		copy(podsCopy, n.Pods)
		n.mu.Unlock()

		for _, pod := range podsCopy {

			switch pod.Phase {
			case pkg.Pending:
				n.cli.startContainer(pod)
			case pkg.Stopped:
				n.cli.deleteContainer(pod)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (n *Node) cleanupContainers() {
	n.mu.Lock()
	podsCopy := make([]*pkg.Pod, len(n.Pods))
	copy(podsCopy, n.Pods)
	n.mu.Unlock()

	for _, pod := range podsCopy {
		go n.cli.deleteContainer(pod)
	}
}

func (n *Node) addPods(pods []*pkg.Pod) {
	n.mu.Lock()
	n.Pods = append(n.Pods, pods...)
	n.mu.Unlock()
}
