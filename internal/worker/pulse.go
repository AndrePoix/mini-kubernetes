package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"mini-kubernetes/pkg"
)

type Pulse struct {
	interval  time.Duration
	node      *Node
	masterURL string
}

func (p *Pulse) newPulse(inter time.Duration, n *Node, mURL string) {
	p.interval = inter
	p.node = n
	p.masterURL = mURL
}

func (p *Pulse) startHeartbeat(ctx context.Context) {
	log.Println("starting heartbeat")
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.sendHeartbeat()
		case <-ctx.Done():
			return
		}
	} //TODO adaptive heartbeat based of the workload
}

func (p *Pulse) sendHeartbeat() {

	data, err := json.Marshal(p.node.NodeInfo)
	if err != nil {
		//TODO handle error
	}
	resp, err := http.Post(p.masterURL+"/heartbeat", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Println("erreur heartbeat:", err)
		return
	}
	defer resp.Body.Close()

	var pods []*pkg.Pod
	log.Println(resp.Body)
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		log.Println("erreur décodage pods:", err)
		return
	}
	log.Println(pods)

	p.node.addPods(pods)
}
