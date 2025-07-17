package main

import (
    "log"
    "net/http"
    "context"
    "sync"
    "os/signal"
    "syscall"
)

var (
    nodes = []*Node{
        {Name: "node1", TotalCPU: 2000, TotalMem: 4096},
    }
    pods []*PodSpec
    mu sync.Mutex // shared mutex for pods and nodes
)


func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
    defer stop()

    setupRoutes()

    go schedulePods()
    go nodeAgent(nodes[0])

    go func() {
        log.Println("API server listening on :8080")
        if err := http.ListenAndServe(":8080", nil); err != nil && err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    <-ctx.Done() // Wait for signal
    log.Println("Interrupt received, cleaning up containers...")

    if err := cleanupContainers(); err != nil {
        log.Printf("Error during cleanup: %v", err)
    }

    log.Println("Cleanup complete, exiting")
}

