package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"sync"

	"github.com/RTradeLtd/Pay/signer"
	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/pay"
	"github.com/RTradeLtd/grpc/pay/request"
	"github.com/RTradeLtd/grpc/pay/response"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server defines our server interface
type Server struct {
	PS *signer.PaymentSigner
}

// RunServer is used to initialize and run our grpc payment server
func RunServer(ctx context.Context, wg *sync.WaitGroup, cfg config.TemporalConfig, logger *zap.SugaredLogger) error {
	url := cfg.Pay.Address + ":" + cfg.Pay.Port
	lis, err := net.Listen(cfg.Protocol, url)
	if err != nil {
		return err
	}
	logger = logger.Named("grpc").Named("server")
	serverOpts, err := options(
		cfg.Pay.TLS.CertPath,
		cfg.Pay.TLS.KeyPath,
		cfg.Pay.AuthKey,
		logger)
	if err != nil {
		return err
	}
	serverService := &Server{}
	gServer := grpc.NewServer(serverOpts...)
	pb.RegisterSignerServer(gServer, serverService)
	// allow for graceful closure if context is cancelled
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("shutting server down")
				gServer.GracefulStop()
				wg.Done()
				return
			}
		}
	}()
	// start the server
	logger.Infow("spinning up server", "address", url)
	return gServer.Serve(lis)
}

/*
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
}*/

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
