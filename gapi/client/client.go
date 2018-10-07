package client

import (
	"context"
	"log"

	"github.com/RTradeLtd/Temporal_Payment-ETH/gapi/response"

	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	pb "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/service"
	"google.golang.org/grpc"
)

// GetSignedPaymentMessage is used to get a signed message
func GetSignedPaymentMessage(grcServerAddress string, insecure bool, req *request.SignRequest) (*response.SignResponse, error) {
	conn, err := grpc.Dial(grcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewSignerClient(conn)
	return client.GetSignedMessage(context.Background(), req)
}
