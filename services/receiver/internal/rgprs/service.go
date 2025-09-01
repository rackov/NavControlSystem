package rgprs

import (
	"context"

	"github.com/rackov/NavControlSystem/proto"
	"github.com/rackov/NavControlSystem/services/receiver/internal/portmanager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Service реализует gRPC-сервис ReceiverService
type Service struct {
	proto.UnimplementedReceiverServiceServer
	portManager *portmanager.PortManager
}

func NewService(pm *portmanager.PortManager) *Service {
	return &Service{portManager: pm}
}

func (s *Service) CreatePort(ctx context.Context, req *proto.CreatePortRequest) (*proto.PortResponse, error) {
	return s.portManager.CreatePort(req)
}

func (s *Service) DeletePort(ctx context.Context, req *proto.DeletePortRequest) (*emptypb.Empty, error) {
	err := s.portManager.DeletePort(req.PortNumber)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) OpenPort(ctx context.Context, req *proto.OpenPortRequest) (*proto.PortResponse, error) {
	return s.portManager.OpenPort(req.PortNumber)
}

func (s *Service) ClosePort(ctx context.Context, req *proto.ClosePortRequest) (*proto.PortResponse, error) {
	return s.portManager.ClosePort(req.PortNumber)
}

func (s *Service) ListPorts(ctx context.Context, _ *emptypb.Empty) (*proto.ListPortsResponse, error) {
	return s.portManager.ListPorts(), nil
}
