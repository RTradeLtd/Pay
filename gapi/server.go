package gapi

import (
	"log"
	"net"

	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	response "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/response"
	pb "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/service"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func generateServerAndList(listenAddr, protocol string) {
	lis, err := net.Listen(protocol, listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()
	gServer := grpc.NewServer()
	server := &Service{}
	pb.RegisterSignerServer(gServer, server)
	gServer.Serve(lis)
}

type Service struct{}

func (s *Service) GetSignedMessage(ctx context.Context, req *request.SignRequest) (*response.SignResponse, error) {
	res := &response.SignResponse{}
	return res, nil
}
