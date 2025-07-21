package main

import (
    "log"
    "encoding/json"
    "net/http"
    "github.com/julienschmidt/httprouter"
)


func createPodHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    var pod PodSpec
    if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    mu.Lock()
    pods = append(pods, &pod)
    mu.Unlock()

    response := map[string]interface{}{
        "message": "Created pod",
        "code":    http.StatusCreated,
        "pod":     pod,
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

func listPodsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    mu.Lock()
    defer mu.Unlock()
    json.NewEncoder(w).Encode(pods)
}


func deletePodHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    name := ps.ByName("name")

    mu.Lock()
    defer mu.Unlock()

    for i, pod := range pods {
        if pod.Name == name {
            pods = append(pods[:i], pods[i+1:]...)
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]string{
                "message": "Pod deleted",
                "name":    name,
            })
            return
        }
    }

    http.Error(w, "Pod not found", http.StatusNotFound)
}


func setupRoutes() {
    router := httprouter.New()

    router.POST("/pods", createPodHandler)
    router.GET("/pods", listPodsHandler)
    router.DELETE("/pods/:name", deletePodHandler)



    log.Println("API server listening on :8080")
    go func() {
      if err := http.ListenAndServe(":8080", router); err != nil && err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()
}
