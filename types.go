package main

type Node struct {
    Name     string
    TotalCPU int // milliCPU (1000 = 1 CPU)
    TotalMem int // MB
    UsedCPU  int
    UsedMem  int
    Pods     []*PodSpec
}

type PodSpec struct {
    Name       string `json:"name"`
    Image      string `json:"image"`
    CPURequest int    `json:"cpu_request"` // milliCPU
    MemRequest int    `json:"mem_request"` // MB
    NodeName   string `json:"node,omitempty"` // assigned node
}
