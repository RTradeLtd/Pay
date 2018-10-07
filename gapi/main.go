package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/RTradeLtd/Temporal_Payment-ETH/gapi/client"
	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	"github.com/RTradeLtd/Temporal_Payment-ETH/gapi/server"
	"github.com/RTradeLtd/config"
	"github.com/ethereum/go-ethereum/common"
)

// This is intended to "demo" the GRPC api
func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		err := errors.New("invalid invocation, ./gapi <server>")
		log.Fatal(err)
	}
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		err := errors.New("CONFIG_PATH env var is empty")
		log.Fatal(err)
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	switch os.Args[1] {
	case "server":
		server.RunServer("127.0.0.1:9090", "tcp", cfg)
	case "client":
		req := &request.SignRequest{
			Address:      common.HexToAddress("0").String(),
			Method:       "0",
			Number:       "0",
			ChargeAmount: "1",
		}
		resp, err := client.GetSignedPaymentMessage("127.0.0.1:9090", false, req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(resp)
	default:
		err := errors.New("argument nto supported")
		log.Fatal(err)
	}

}
