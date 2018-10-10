package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"

	request "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/request"
	response "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/response"
	pb "github.com/RTradeLtd/Temporal_Payment-ETH/gapi/service"
	"github.com/RTradeLtd/Temporal_Payment-ETH/signer"
	"github.com/RTradeLtd/config"
	"github.com/ethereum/go-ethereum/common"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// RunServer allows us to run our GRPC API Server
func RunServer(listenAddr, protocol string, cfg *config.TemporalConfig) error {
	lis, err := net.Listen(protocol, listenAddr)
	if err != nil {
		return err
	}
	defer lis.Close()
	gServer := grpc.NewServer()
	ps, err := signer.GeneratePaymentSigner(cfg)
	server := &Server{
		PS: ps,
	}
	pb.RegisterSignerServer(gServer, server)
	if err = gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

// Server defines our server interface
type Server struct {
	PS *signer.PaymentSigner
}

// GetSignedMessage allows the caller (client) to request a signed message
func (s *Server) GetSignedMessage(ctx context.Context, req *request.SignRequest) (*response.SignResponse, error) {
	fmt.Println("message received, processing...")
	addr := req.Address
	method := req.Method
	number := req.Number
	addrTyped := common.HexToAddress(addr)
	methodUint64, err := strconv.ParseUint(method, 10, 64)
	if err != nil {
		return nil, err
	}
	methodUint8 := uint8(methodUint64)
	numberBig, valid := new(big.Int).SetString(number, 10)
	if !valid {
		return nil, errors.New("failed to convert payment number to big int")
	}
	chargeAmountBig, valid := new(big.Int).SetString(req.ChargeAmount, 10)
	if !valid {
		return nil, errors.New("failed to convert charge amount from string to big int")
	}
	msg, err := s.PS.GenerateSignedPaymentMessagePrefixed(
		addrTyped, methodUint8, numberBig, chargeAmountBig,
	)
	if err != nil {
		fmt.Println("failed to generate signed payment message ", err.Error())
		return nil, err
	}
	fmt.Printf("%+v\n", msg)
	hEncoded := hex.EncodeToString(msg.H[:])
	rEncoded := hex.EncodeToString(msg.R[:])
	sEncoded := hex.EncodeToString(msg.S[:])
	addressString := msg.Address.String()
	hashEncoded := hex.EncodeToString(msg.Hash)
	sigEncoded := hex.EncodeToString(msg.S[:])
	res := &response.SignResponse{
		H:       hEncoded,
		R:       rEncoded,
		S:       sEncoded,
		V:       fmt.Sprintf("%v", msg.V),
		Address: addressString,
		Hash:    hashEncoded,
		Sig:     sigEncoded,
	}
	fmt.Println("processing finished")
	return res, nil
}
