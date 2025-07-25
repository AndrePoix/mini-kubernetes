package master

import (
	"log"
	"sync"

	"mini-kubernetes/pkg"
)

type Node struct {
	pkg.NodeInfo
	PodsAssigned []*pkg.Pod
}

type NodeManager struct {
	nodes            map[string]*Node
	mu               sync.RWMutex
	PodsToBeAssigned []*pkg.Pod
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes: make(map[string]*Node),
	}
}

func (nm *NodeManager) SchedulePods() {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	log.Println("Scheduling : ", nm.PodsToBeAssigned)
	newPending := nm.PodsToBeAssigned[:0]

	for _, pod := range nm.PodsToBeAssigned {
		if pod.Phase != pkg.Pending {
			continue
		}
		for _, node := range nm.nodes {
			if node.TotalCPU-node.UsedCPU >= pod.CPURequest &&
				node.TotalMem-node.UsedMem >= pod.MemRequest {

				pod.NodeName = node.Name
				pod.Phase = pkg.Send

				node.UsedCPU += pod.CPURequest
				node.UsedMem += pod.MemRequest

				node.PodsAssigned = append(node.PodsAssigned, pod)
				break
			}
		}
		if pod.Phase == pkg.Pending {
			newPending = append(newPending, pod)
		}
	}
	nm.PodsToBeAssigned = newPending
}

func (nm *NodeManager) RegisterNode(info pkg.NodeInfo) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.nodes[info.Name]; !exists {
		nm.nodes[info.Name] = &Node{
			NodeInfo:     info,
			PodsAssigned: []*pkg.Pod{},
		}
	} else {
		n := nm.nodes[info.Name]
		n.TotalCPU = info.TotalCPU
		n.TotalMem = info.TotalMem
		n.UsedCPU = info.UsedCPU
		n.UsedMem = info.UsedMem
	}
}

func (nm *NodeManager) GetAssignedPods(nodeName string) []pkg.Pod {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	node, exists := nm.nodes[nodeName]
	if !exists {
		return nil
	}
	pods := make([]pkg.Pod, 0, len(node.PodsAssigned))
	for _, p := range node.PodsAssigned {
		pods = append(pods, *p)
	}
	return pods
}

func (nm *NodeManager) ClearAssignedPods(nodeName string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	node, exists := nm.nodes[nodeName]
	if !exists {
		return
	}
	node.PodsAssigned = node.PodsAssigned[:0]
}

func (nm *NodeManager) AssignPodToBeScheduled(pod *pkg.Pod) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	log.Printf("Assigned list: %+v", nm.PodsToBeAssigned)
	nm.PodsToBeAssigned = append(nm.PodsToBeAssigned, pod)
}
