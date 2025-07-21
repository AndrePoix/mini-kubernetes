package main

type Node struct {
    Name     string
    TotalCPU int // milliCPU (1000 = 1 CPU)
    TotalMem int // MB
    UsedCPU  int
    UsedMem  int
    Pods     []*PodSpec
}

type PodSpecInput struct {
    Name       string `json:"name"`
    Image      string `json:"image"`
    CPURequest int    `json:"cpu_request"` // milliCPU
    MemRequest int    `json:"mem_request"` // MB

    //for networking
    ExposePort  string `json:"expose_port,omitempty"`   
    HostPort    string `json:"host_port,omitempty"`
}

type PodSpec struct {
    PodSpecInput
    Running bool `json:"running"`
    NodeName   string `json:"node,omitempty"` // assigned node
    ContainerID string `json:"container_id"`
    ToDelete bool `json:-`
}
