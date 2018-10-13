package client

import (
	"context"
	"fmt"
	"log"

	request "github.com/RTradeLtd/Pay/gapi/request"
	"github.com/RTradeLtd/Pay/gapi/response"
	pb "github.com/RTradeLtd/Pay/gapi/service"
	"google.golang.org/grpc"
)

// GetSignedPaymentMessage is used to get a signed message
func GetSignedPaymentMessage(grcServerAddress string, insecure bool, req *request.SignRequest) (*response.SignResponse, error) {
	fmt.Println(1)
	conn, err := grpc.Dial(grcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(2)
	defer conn.Close()
	client := pb.NewSignerClient(conn)
	return client.GetSignedMessage(context.Background(), req)
}
