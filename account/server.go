package account

import (
	"context"
	"fmt"
	"go-graphql-grpc-microservice/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterAccountServiceServer(srv, &grpcServer{pb.UnimplementedAccountServiceServer{}, s})
	reflection.Register(srv)
	return srv.Serve(lis)
}

func (s *grpcServer) PostAccount(ctx context.Context, r *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	a, err := s.service.PostAccount(ctx, r.GetName())

	if err != nil {
		return nil, err
	}
	return &pb.PostAccountResponse{Account: &pb.Account{
		Id:   a.ID,
		Name: a.Name,
	}}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	a, err := s.service.GetAccount(ctx, r.GetId())

	if err != nil {
		return nil, err
	}
	return &pb.GetAccountResponse{Account: &pb.Account{
		Id:   a.ID,
		Name: a.Name,
	}}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	b, err := s.service.GetAccounts(ctx, r.GetSkip(), r.GetTake())

	if err != nil {
		return nil, err
	}
	arr := []*pb.Account{}
	for _, a := range b {
		arr = append(arr, &pb.Account{
			Id:   a.ID,
			Name: a.Name,
		})
	}
	return &pb.GetAccountsResponse{Accounts: arr}, nil
}
