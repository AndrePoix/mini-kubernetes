package main

import (
    "log"
    "time"
    "context"
)

func schedulePods(parentContext context.Context, cli *client.Client) {
    for {
        mu.Lock()
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
            if pod.ToDelete {
                log.Printf("Tryn to delete pod %s on node %s\n", pod.Name, pod.NodeName)
                deleteContainer(cli, pod.ContainerID)
            }
        }
        mu.Unlock()
        time.Sleep(5 * time.Second)
    }
}
