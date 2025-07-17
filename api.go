package main

import (
    "encoding/json"
    "net/http"
    "sync"
)

var mu sync.Mutex // shared mutex for pods and nodes

func createPodHandler(w http.ResponseWriter, r *http.Request) {
    var pod PodSpec
    if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()
    pods = append(pods, &pod)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(pod)
}

func listPodsHandler(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()
    json.NewEncoder(w).Encode(pods)
}

func setupRoutes() {
    http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodPost:
            createPodHandler(w, r)
        case http.MethodGet:
            listPodsHandler(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
}
