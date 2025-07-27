package worker

import (
	"context"
	"mini-kubernetes/pkg"
	"sync"
	"time"
)

type Node struct {
	pkg.NodeInfo
	mu  sync.Mutex
	cli *Client
	ctx context.Context
}

func NewNode(ctx context.Context, nodeInfo pkg.NodeInfo) (*Node, error) {
	// Initialize Docker client
	cli := &Client{}
	cli.initClient(ctx)

	return &Node{
		NodeInfo: nodeInfo,
		cli:      cli,
		ctx:      ctx,
	}, nil
}

func (n *Node) GetNodeInfo() pkg.NodeInfo {
	n.mu.Lock()
	defer n.mu.Unlock()

	usedCPU := 0
	usedMem := 0

	for _, pod := range n.Pods {
		if pod.Phase == pkg.Running {
			usedCPU += pod.CPURequest
			usedMem += pod.MemRequest
		}
	}

	n.UsedCPU = usedCPU
	n.UsedMem = usedMem

	return n.NodeInfo
}

func (n *Node) nodeAgent() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.processPods()
		}
	}
}

func (n *Node) processPods() {
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
		case pkg.Stopped:
			n.removePod(pod)
		}
	}
}

func (n *Node) cleanupContainers() {
	n.mu.Lock()
	podsCopy := make([]*pkg.Pod, len(n.Pods))
	copy(podsCopy, n.Pods)
	n.mu.Unlock()

	for _, pod := range podsCopy {
		if pod.ContainerID != "" {
			n.cli.deleteContainer(pod)
		}
	}
}

func (n *Node) addPods(pods []*pkg.Pod) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, newPod := range pods {
		// Check if pod already exist to avoid duplicates
		exists := false
		for _, existingPod := range n.Pods {
			if existingPod.Name == newPod.Name {
				exists = true
				break
			}
		}

		if !exists {
			n.Pods = append(n.Pods, newPod)
		}
	}
}

func (n *Node) removePod(podToRemove *pkg.Pod) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, pod := range n.Pods {
		if pod.Name == podToRemove.Name {
			// Remove pod from slice
			n.Pods = append(n.Pods[:i], n.Pods[i+1:]...)
			break
		}
	}
}
