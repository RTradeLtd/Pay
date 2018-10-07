package gapi

import (
	"net"
	pb "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/service"
	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	"google.golang.org/grpc"
)

func generateServer(listenAddr, protocol string) error {
	lis, err := net.Listen(protocol, listenAddr)
	if err != nil {
		return err
	}
	defer lis.Close()
	gServer := grpc.NewServer()
	server := Server{}
	pb.RegisterSignerServer(server, server)
}


type Server struct {}

func (s *Server) GetSignedMessage(ctx context.Context, in *request.SignRequest, opts ...grpc.CallOption) (*response.SignResponse, error) {
	out := new(response.SignResponse)
	err := grpc.Invoke(ctx, "/models.Signer/GetSignedMessage", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}