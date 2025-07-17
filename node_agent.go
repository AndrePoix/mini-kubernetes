package main

import (
    "context"
    "log"
    "time"

    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

func nodeAgent(node *Node) {
    cli, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        log.Fatal(err)
    }

    for {
        mu.Lock()
        for _, pod := range node.Pods {
            containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
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
                resp, err := cli.ContainerCreate(context.Background(), &container.Config{
                    Image: pod.Image,
                    Tty:   true,
                }, nil, nil, nil, pod.Name)
                if err != nil {
                    log.Println("Error creating container:", err)
                    continue
                }

                if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
                    log.Println("Error starting container:", err)
                    continue
                }
            }
        }
        mu.Unlock()
        time.Sleep(10 * time.Second)
    }
}