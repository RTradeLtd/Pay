package main

import (
	"fmt"
	"log"
	"net"

	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	response "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/response"
	pb "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/service"
	"github.com/RTradeLtd/Temporal_Payment-ETH/signer"
	"github.com/RTradeLtd/config"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func generateServerAndList(listenAddr, protocol string, cfg *config.TemporalConfig) {
	lis, err := net.Listen(protocol, listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()
	gServer := grpc.NewServer()
	ps, err := signer.GeneratePaymentSigner(cfg)
	server := &Server{
		PS: ps,
	}
	pb.RegisterSignerServer(gServer, server)
	gServer.Serve(lis)
}

type Server struct {
	PS *signer.PaymentSigner
}

func (s *Server) GetSignedMessage(ctx context.Context, req *request.SignRequest) (*response.SignResponse, error) {
	res := &response.SignResponse{}
	fmt.Println("new message received")
	return res, nil
}
