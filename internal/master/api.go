package master

import (
	"encoding/json"
	"log"
	"net/http"

	"mini-kubernetes/pkg"

	"github.com/julienschmidt/httprouter"
)

type Api struct {
	listeningPort string
	NodeManager   *NodeManager
}

func (api *Api) initListeningPort(listeningPort string) {
	api.listeningPort = listeningPort
}

func (api *Api) createPodHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var inputPod pkg.PodSpecInput
	if err := json.NewDecoder(r.Body).Decode(&inputPod); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println(inputPod)

	pod := &pkg.Pod{
		PodSpecInput: inputPod,
		Phase:        pkg.Pending,
	}
	log.Println(pod)
	api.NodeManager.AssignPodToBeScheduled(pod)

	response := map[string]interface{}{
		"message": "Created pod",
		"code":    http.StatusCreated,
		"pod":     pod,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (api *Api) listPodsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//json.NewEncoder(w).Encode(pods)
}

func (api *Api) deletePodHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//name := ps.ByName("name")

	//http.Error(w, "Pod not found", http.StatusNotFound)
}

func (api *Api) heartbeatHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var nodeInfo pkg.NodeInfo
	if err := json.NewDecoder(r.Body).Decode(&nodeInfo); err != nil {
		log.Println("Invalid : ")
		http.Error(w, "Invalid node heartbeat", http.StatusBadRequest)
		return
	}
	log.Println("Registering : ", nodeInfo)
	api.NodeManager.RegisterNode(nodeInfo)
	assigned := api.NodeManager.GetAssignedPods(nodeInfo.Name)
	api.NodeManager.ClearAssignedPods(nodeInfo.Name)

	log.Println("Sendin : ", assigned)

	json.NewEncoder(w).Encode(assigned)
}

func (api *Api) setupRoutes() {
	router := httprouter.New()

	router.POST("/pods", api.createPodHandler)
	router.GET("/pods", api.listPodsHandler)
	router.DELETE("/pods/:name", api.deletePodHandler)
	router.POST("/heartbeat", api.heartbeatHandler)

	log.Println("API server listening on :" + api.listeningPort)
	go func() {
		if err := http.ListenAndServe(":"+api.listeningPort, router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}
