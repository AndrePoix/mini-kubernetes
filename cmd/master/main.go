package main

import (
	"mini-kubernetes/internal/master"
)

func main() {
	m := master.NewMaster()
	m.Start()
}
