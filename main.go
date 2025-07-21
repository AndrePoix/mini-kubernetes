package main

import (
    "log"
    "context"
    "sync"
    "os/signal"
    "syscall"
    "github.com/docker/docker/client"
)

var (
    nodes = []*Node{
        {Name: "node1", TotalCPU: 2_000, TotalMem: 2_048_000_000}, // 2 cpu and 2gb
    }
    pods []*PodSpec
    mu sync.Mutex // shared mutex for pods and nodes
)


func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
    defer stop()

    cli, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        log.Fatalf("Failed to create Docker client: %v", err)
    }
    defer cli.Close()

    setupRoutes()

    go schedulePods(ctx )
    go nodeAgent(ctx, nodes[0], cli)


    <-ctx.Done() // Wait for signal
    log.Println("Interrupt received, cleaning up containers...")

    cleanupContainers(cli, nodes[0])

    log.Println("Cleanup complete, exiting")
}

