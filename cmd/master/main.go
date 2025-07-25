package main

import (
	"mini-kubernetes/internal/master"
)

func main() {
	nodeManager := master.NewNodeManager()
	api := &master.Api{
		NodeManager: nodeManager,
	}

	m := &master.Master{
		NodeManager: nodeManager,
		Api:         api,
	}
	m.Start()
}
