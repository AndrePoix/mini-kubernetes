package main

import (
    "log"
    "net/http"
)

var (
    nodes = []*Node{
        {Name: "node1", TotalCPU: 2000, TotalMem: 4096},
    }
    pods []*PodSpec
)

func main() {
    setupRoutes()

    go schedulePods()
    go nodeAgent(nodes[0])

    log.Println("API server listening on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
