package main

import (
	"context"
	"fmt"
	"log"

	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	pb "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/service"
	"google.golang.org/grpc"
)

type Client struct{}

func generateClient(grcServerAddress string, insecure bool) {
	conn, err := grpc.Dial(grcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewSignerClient(conn)
	req := request.SignRequest{
		Address:      "adress",
		Method:       "method",
		Number:       "number",
		ChargeAmount: "chargeAmount",
	}
	resp, err := client.GetSignedMessage(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
