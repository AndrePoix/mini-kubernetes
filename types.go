package main

type PodPhase string
const (
    Pending    PodPhase = "Pending" // waiting for container creation
    Running    PodPhase = "Running"
    Stopped    PodPhase = "Stopped" 
    Terminating PodPhase = "Terminating" //waiting for delete
    Succeeded  PodPhase = "Succeeded"
    Failed     PodPhase = "Failed"
)

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
    Phase       PodPhase
    NodeName   string `json:"node,omitempty"` // assigned node
    ContainerID string `json:"container_id"`
}
