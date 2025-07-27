package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"mini-kubernetes/pkg"
	pb "mini-kubernetes/proto/master"
)

type MasterClient struct {
	client     pb.MasterServiceClient
	conn       *grpc.ClientConn
	masterAddr string
	interval   time.Duration
	node       *Node
}

func NewMasterClient(masterAddr string, interval time.Duration, node *Node) (*MasterClient, error) {

	creds, err := credentials.NewClientTLSFromFile("certs/server.crt", "")
	if err != nil {
		log.Fatalf("failed to load server TLS certificate: %v", err)
	}

	conn, err := grpc.NewClient(masterAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %v", err)
	}

	client := pb.NewMasterServiceClient(conn)

	return &MasterClient{
		client:     client,
		conn:       conn,
		masterAddr: masterAddr,
		interval:   interval,
		node:       node,
	}, nil
}

func (mc *MasterClient) Close() error {
	return mc.conn.Close()
}

func (mc *MasterClient) StartHeartbeat(ctx context.Context) {
	log.Println("starting gRPC heartbeat")
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.sendHeartbeat(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (mc *MasterClient) sendHeartbeat(ctx context.Context) {
	nodeInfo := mc.node.GetNodeInfo()

	req := &pb.HeartbeatRequest{
		NodeInfo: &pb.NodeInfo{
			Name:     nodeInfo.Name,
			TotalCpu: int32(nodeInfo.TotalCPU),
			TotalMem: int32(nodeInfo.TotalMem),
			UsedCpu:  int32(nodeInfo.UsedCPU),
			UsedMem:  int32(nodeInfo.UsedMem),
		},
	}

	// Send heartbeat with timeout
	heartbeatCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := mc.client.Heartbeat(heartbeatCtx, req)
	if err != nil {
		log.Printf("heartbeat error: %v", err)
		return
	}

	if len(resp.AssignedPods) > 0 {
		log.Printf("received %d pods from master", len(resp.AssignedPods))

		pods := make([]*pkg.Pod, len(resp.AssignedPods))
		for i, pbPod := range resp.AssignedPods {
			pods[i] = &pkg.Pod{
				PodSpecInput: pkg.PodSpecInput{
					Name:       pbPod.Spec.Name,
					Image:      pbPod.Spec.Image,
					CPURequest: int(pbPod.Spec.CpuRequest),
					MemRequest: int(pbPod.Spec.MemRequest),
					ExposePort: pbPod.Spec.ExposePort,
					HostPort:   pbPod.Spec.HostPort,
				},
				Phase:       convertPbPodPhase(pbPod.Phase),
				NodeName:    pbPod.NodeName,
				ContainerID: pbPod.ContainerId,
			}
		}

		mc.node.addPods(pods)
	}
}

func convertPbPodPhase(phase pb.PodPhase) pkg.PodPhase {
	switch phase {
	case pb.PodPhase_PENDING:
		return pkg.Pending
	case pb.PodPhase_SEND:
		return pkg.Send
	case pb.PodPhase_RUNNING:
		return pkg.Running
	case pb.PodPhase_STOPPING:
		return pkg.Stopping
	case pb.PodPhase_STOPPED:
		return pkg.Stopped
	case pb.PodPhase_SUCCEEDED:
		return pkg.Succeeded
	case pb.PodPhase_FAILED:
		return pkg.Failed
	default:
		return pkg.Pending
	}
}
