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
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-context.Background().Done(): //TODO Context cancel
			return
		case <-ticker.C:
			n.mu.Lock()
			pods := make([]*pkg.Pod, len(n.Pods))
			copy(pods, n.Pods)
			n.mu.Unlock()

			for _, pod := range pods {
				switch pod.Phase {
				case pkg.Send:
					n.cli.startContainer(pod)
				case pkg.Stopping:
					n.cli.deleteContainer(pod)
				}
			}
		}
	}
}

func (n *Node) cleanupContainers() {
	n.mu.Lock()
	podsCopy := make([]*pkg.Pod, len(n.Pods))
	copy(podsCopy, n.Pods)
	n.mu.Unlock()

	for _, pod := range podsCopy {
		n.cli.deleteContainer(pod) //maybe do go ?
	}
}

func (n *Node) addPods(pods []*pkg.Pod) {
	n.mu.Lock()
	n.Pods = append(n.Pods, pods...)
	n.mu.Unlock()
}
