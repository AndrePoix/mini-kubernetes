package main

import (
    "context"
    "log"
    "time"

    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
    "github.com/docker/go-connections/nat"
)

var startedContainers []string

func nodeAgent(parentContext context.Context, node *Node, cli *client.Client) {
    for {
        // We copy to not lock the mutex for too long
        mu.Lock()
        podsCopy := make([]*PodSpec, len(node.Pods))
        copy(podsCopy, node.Pods)
        mu.Unlock()
        
        for _, pod := range podsCopy {

            if !pod.Running {
                pod.Running = true
                log.Printf("Starting container for pod %s\n", pod.Name)
                containerConfig := &container.Config{
                    Image: pod.Image,
                    Tty:   true,
                }

                hostConfig := &container.HostConfig{}
                if pod.CPURequest > 0 {
                    hostConfig.Resources.NanoCPUs = int64(pod.CPURequest) * 1_000_000  // milliCPU to nanoCPU
                }
                if pod.MemRequest > 0 {
                    hostConfig.Resources.Memory = int64(pod.MemRequest)  // in bytes
                }
                if (pod.ExposePort != "" && pod.HostPort != ""){
                    hostConfig.PortBindings = nat.PortMap{
                         nat.Port(pod.ExposePort + "/tcp"): []nat.PortBinding{
                            {
                                HostIP:   "127.0.0.1",
                                HostPort: pod.HostPort,
                            },
                        },
                    }
                    
                    containerConfig.ExposedPorts = nat.PortSet{
                        nat.Port(pod.ExposePort + "/tcp"): struct{}{},
                    }
                }
            
                resp, err := cli.ContainerCreate(parentContext,containerConfig ,hostConfig, nil, nil, pod.Name)
                //assign ContainerID
                pod.ContainerID = resp.ID
                if err != nil {
                    log.Println("Error creating container:", err)
                    continue
                }

                if err := cli.ContainerStart(parentContext, resp.ID, container.StartOptions{}); err != nil {
                    log.Println("Error starting container:", err)
                    continue
                }
            }
        }
        time.Sleep(10 * time.Second)
    }
}

func deleteContainer(cli *client.Client, containerID string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    timeoutSecs := 5 
    log.Printf("Stopping container %s", containerID)
    err := cli.ContainerStop(ctx, containerID, container.StopOptions{
        Timeout: &timeoutSecs,
    })

    if err != nil {
        log.Printf("Failed to stop container %s: %v. Will try force remove", containerID, err)

        // Try force remove
        removeErr := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
            Force: true,
        })
        if removeErr != nil {
            log.Printf("Force remove failed for %s: %v", containerID, removeErr)
            return removeErr
        } else {
            log.Printf("Force removed container %s", containerID)
        }
    }
    err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})

    if err != nil {
        return err
    }

    mu.Lock()
    // Filter node.Pods to remove this pod
    newPods := make([]*PodSpec, 0, len(node.Pods))
    for _, p := range node.Pods {
        if p.ContainerID != pod.ContainerID {
            newPods = append(newPods, p)
        }
    }
    node.Pods = newPods
    pod = nil // free for the garbage collector
    mu.Unlock()

    return nil 
}

func cleanupContainers(cli *client.Client) {
    mu.Lock()
    podsCopy := make([]*PodSpec, len(node.Pods))
    copy(podsCopy, node.Pods)
    mu.Unlock()

    for _, pod := range podsCopy {
        deleteContainer(cli, pod.ContainerID)
    }
}
