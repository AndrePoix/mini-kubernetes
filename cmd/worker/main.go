package main

import (
	"mini-kubernetes/internal/worker"
)

func main() {

	w := &worker.Worker{
		Pulse: &worker.Pulse{},
	}

	w.Start()
}
