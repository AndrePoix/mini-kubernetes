package master

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mini-kubernetes/pkg"
	pb "mini-kubernetes/proto/master"
)

type GRPCServer struct {
	pb.UnimplementedMasterServiceServer
	NodeManager *NodeManager
	server      *grpc.Server
	port        string
}

func NewGRPCServer(nodeManager *NodeManager, port string) *GRPCServer {
	return &GRPCServer{
		NodeManager: nodeManager,
		port:        port,
	}
}

func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", s.port, err)
	}

	s.server = grpc.NewServer()
	pb.RegisterMasterServiceServer(s.server, s)

	log.Printf("gRPC server listening on port %s", s.port)

	go func() {
		if err := s.server.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	return nil
}

func (s *GRPCServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

func (s *GRPCServer) CreatePod(ctx context.Context, req *pb.CreatePodRequest) (*pb.CreatePodResponse, error) {
	if req.PodSpec == nil {
		return nil, status.Error(codes.InvalidArgument, "pod spec is required")
	}

	pod := &pkg.Pod{
		PodSpecInput: pkg.PodSpecInput{
			Name:       req.PodSpec.Name,
			Image:      req.PodSpec.Image,
			CPURequest: int(req.PodSpec.CpuRequest),
			MemRequest: int(req.PodSpec.MemRequest),
			ExposePort: req.PodSpec.ExposePort,
			HostPort:   req.PodSpec.HostPort,
		},
		Phase: pkg.Pending,
	}

	log.Printf("creating pod: %+v", pod)
	s.NodeManager.AssignPodToBeScheduled(pod)

	pbPod := &pb.Pod{
		Spec: &pb.PodSpecInput{
			Name:       pod.Name,
			Image:      pod.Image,
			CpuRequest: int32(pod.CPURequest),
			MemRequest: int32(pod.MemRequest),
			ExposePort: pod.ExposePort,
			HostPort:   pod.HostPort,
		},
		Phase:       pb.PodPhase_PENDING,
		NodeName:    pod.NodeName,
		ContainerId: pod.ContainerID,
		CreatedAt:   timestamppb.Now(),
	}

	return &pb.CreatePodResponse{
		Message: "Created pod",
		Code:    201,
		Pod:     pbPod,
	}, nil
}

func (s *GRPCServer) ListPods(ctx context.Context, req *pb.ListPodsRequest) (*pb.ListPodsResponse, error) {
	// TODO : list pods
	return &pb.ListPodsResponse{
		Pods: []*pb.Pod{},
	}, nil
}

func (s *GRPCServer) DeletePod(ctx context.Context, req *pb.DeletePodRequest) (*pb.DeletePodResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "pod name is required")
	}

	// TODO : pod deletion
	return &pb.DeletePodResponse{
		Message: "Pod deleted",
		Code:    200,
	}, nil
}

func (s *GRPCServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.NodeInfo == nil {
		return nil, status.Error(codes.InvalidArgument, "node info is required")
	}

	nodeInfo := pkg.NodeInfo{
		Name:     req.NodeInfo.Name,
		TotalCPU: int(req.NodeInfo.TotalCpu),
		TotalMem: int(req.NodeInfo.TotalMem),
		UsedCPU:  int(req.NodeInfo.UsedCpu),
		UsedMem:  int(req.NodeInfo.UsedMem),
	}

	log.Printf("registering node: %+v", nodeInfo)
	s.NodeManager.RegisterNode(nodeInfo)

	assigned := s.NodeManager.GetAssignedPods(nodeInfo.Name)
	s.NodeManager.ClearAssignedPods(nodeInfo.Name)

	pbPods := make([]*pb.Pod, len(assigned))
	for i, pod := range assigned {
		pbPods[i] = &pb.Pod{
			Spec: &pb.PodSpecInput{
				Name:       pod.Name,
				Image:      pod.Image,
				CpuRequest: int32(pod.CPURequest),
				MemRequest: int32(pod.MemRequest),
				ExposePort: pod.ExposePort,
				HostPort:   pod.HostPort,
			},
			Phase:       convertPodPhase(pod.Phase),
			NodeName:    pod.NodeName,
			ContainerId: pod.ContainerID,
		}
	}

	log.Printf("sending assigned pods: %d", len(pbPods))

	return &pb.HeartbeatResponse{
		AssignedPods: pbPods,
		Status:       "ok",
	}, nil
}

func (s *GRPCServer) WatchPods(req *pb.ListPodsRequest, stream pb.MasterService_WatchPodsServer) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			// TODO : Send pod updtae
		}
	}
}

func (s *GRPCServer) NodeHeartbeatStream(stream pb.MasterService_NodeHeartbeatStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		ctx := stream.Context()
		resp, err := s.Heartbeat(ctx, req)
		if err != nil {
			return err
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func convertPodPhase(phase pkg.PodPhase) pb.PodPhase {
	switch phase {
	case pkg.Pending:
		return pb.PodPhase_PENDING
	case pkg.Send:
		return pb.PodPhase_SEND
	case pkg.Running:
		return pb.PodPhase_RUNNING
	case pkg.Stopping:
		return pb.PodPhase_STOPPING
	case pkg.Stopped:
		return pb.PodPhase_STOPPED
	case pkg.Succeeded:
		return pb.PodPhase_SUCCEEDED
	case pkg.Failed:
		return pb.PodPhase_FAILED
	default:
		return pb.PodPhase_PENDING
	}
}
