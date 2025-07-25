package worker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
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
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go p.sendHeartbeat()
		case <-ctx.Done():
			return
		}
	} //TODO adaptive heartbeat based of the workload
}

func (p *Pulse) sendHeartbeat() {
	resp, err := http.Post(p.masterURL+"/heartbeat", "application/json", nil)
	if err != nil {
		log.Println("Erreur heartbeat:", err)
		return
	}
	defer resp.Body.Close()

	var pods []*Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		log.Println("Erreur décodage pods:", err)
		return
	}

	p.node.addPods(pods)
}
