package main

import (
    "log"
    "net/http"
    "context"
    "sync"
    "os/signal"
    "syscall"
    "github.com/docker/docker/client"
)

var (
    nodes = []*Node{
        {Name: "node1", TotalCPU: 2000, TotalMem: 2_048_000_000}, // 2 cpu and 2gb
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

    go func() {
        log.Println("API server listening on :8080")
        if err := http.ListenAndServe(":8080", nil); err != nil && err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    <-ctx.Done() // Wait for signal
    log.Println("Interrupt received, cleaning up containers...")

    if err := cleanupContainers(cli); err != nil {
        log.Printf("Error during cleanup: %v", err)
    }

    log.Println("Cleanup complete, exiting")
}

