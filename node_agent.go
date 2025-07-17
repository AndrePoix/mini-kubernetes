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
            ctx, cancel := context.WithTimeout(parentContext, 10*time.Second)
            defer cancel()
            containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
            if err != nil {
                log.Println("Error listing containers:", err)
                continue
            }

            running := false
            for _, c := range containers {
                for _, name := range c.Names {
                    if name == "/"+pod.Name {
                        running = true
                        break
                    }
                }
                if running {
                    break
                }
            }

            if !running {
                log.Printf("Starting container for pod %s\n", pod.Name)

                containerConfig := &container.Config{
                    Image: pod.Image,
                    Tty:   true,
                }

                hostConfig := &container.HostConfig{}
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
            
                resp, err := cli.ContainerCreate(ctx,containerConfig ,hostConfig, nil, nil, pod.Name)
                startedContainers = append(startedContainers, resp.ID)
                if err != nil {
                    log.Println("Error creating container:", err)
                    continue
                }

                if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
                    log.Println("Error starting container:", err)
                    continue
                }
            }
        }
        time.Sleep(10 * time.Second)
    }
}

func cleanupContainers(cli *client.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()


    for _, containerID := range startedContainers {
        timeoutSecs := 5 
        log.Printf("Stopping container %s", containerID)
        err := cli.ContainerStop(ctx, containerID, container.StopOptions{
            Timeout: &timeoutSecs,
        })

        if err != nil {
            log.Printf("Failed to stop container %s: %v. Will try force remove", containerID, err)

            // Try force remove
            removeErr := cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{
                Force: true,
            })
            if removeErr != nil {
                log.Printf("Force remove failed for %s: %v", containerID, removeErr)
            } else {
                log.Printf("Force removed container %s", containerID)
            }

            return nil
        }

        log.Printf("Removing container %s", containerID)
        err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
        if err != nil {
            log.Printf("Failed to remove container %s: %v", containerID, err)
        }
    }
    return nil
}
