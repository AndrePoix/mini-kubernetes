package pkg

type PodPhase string

const (
	Pending   PodPhase = "Pending" // waiting for to be send
	Send      PodPhase = "Send"
	Running   PodPhase = "Running"
	Stopping  PodPhase = "Stopping" //waiting for delete
	Stopped   PodPhase = "Stopped"
	Succeeded PodPhase = "Succeeded"
	Failed    PodPhase = "Failed"
)

type PodSpecInput struct {
	Name       string `json:"name"`
	Image      string `json:"image"`
	CPURequest int    `json:"cpu_request"` // milliCPU
	MemRequest int    `json:"mem_request"` // MB

	//for networking
	ExposePort string `json:"expose_port,omitempty"`
	HostPort   string `json:"host_port,omitempty"`
}

type Pod struct {
	PodSpecInput
	Phase       PodPhase
	NodeName    string `json:"node,omitempty"` // assigned node
	ContainerID string `json:"container_id"`
}

type NodeInfo struct {
	Name     string
	TotalCPU int // milliCPU (1000 = 1 CPU)
	TotalMem int // MB
	UsedCPU  int
	UsedMem  int
	Pods     []*Pod
}
