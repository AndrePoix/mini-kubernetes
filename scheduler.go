package main

import (
    "log"
    "time"
    "context"
)

func schedulePods(parentContext context.Context) {
    for {
        mu.Lock()
        newPods := make([]*PodSpec, 0, len(pods))
        for _, pod := range pods {
            log.Println(pod)
            if pod.NodeName == "" {
                for _, node := range nodes { // Find first node with enough capacity
                    if (node.TotalCPU - node.UsedCPU) >= pod.CPURequest &&
                        (node.TotalMem - node.UsedMem) >= pod.MemRequest {
                        pod.NodeName = node.Name
                        node.UsedCPU += pod.CPURequest
                        node.UsedMem += pod.MemRequest
                        node.Pods = append(node.Pods, pod)
                        log.Printf("Scheduled pod %s on node %s\n", pod.Name, node.Name)
                        break
                    }
                }
            }
            if pod.Phase == Terminating {
                for _, node := range nodes {
                    if node.Name == pod.NodeName {
                        newNodePods := make([]*PodSpec, 0, len(node.Pods))
                        for _, p := range node.Pods {
                            if p.ContainerID != pod.ContainerID {
                                newNodePods = append(newNodePods, p)
                            }
                        }
                        node.Pods = newNodePods

                        // Free node resources
                        node.UsedCPU -= pod.CPURequest
                        node.UsedMem -= pod.MemRequest

                        log.Printf("Removed pod %s from node %s\n", pod.Name, node.Name)
                        break
                    }
                }
                
                continue
            }
            newPods = append(newPods, pod)
        }
        pods = newPods
        mu.Unlock()
        time.Sleep(5 * time.Second)
    }
}
