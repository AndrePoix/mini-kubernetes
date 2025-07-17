package main

import (
    "log"
    "time"
)

func schedulePods() {
    for {
        mu.Lock()
        for _, pod := range pods {
            if pod.NodeName == "" {
                for _, node := range nodes {
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
        }
        mu.Unlock()
        time.Sleep(5 * time.Second)
    }
}
