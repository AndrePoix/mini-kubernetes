package main

import (
    "context"
    "log"
    "time"

    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

var startedContainers []string

func nodeAgent(parentContext context.Context, node *Node, cli *client.Client) {
    for {
        mu.Lock()
        for _, pod := range node.Pods {
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
                resp, err := cli.ContainerCreate(ctx, &container.Config{
                    Image: pod.Image,
                    Tty:   true,
                }, nil, nil, nil, pod.Name)
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
        mu.Unlock()
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
