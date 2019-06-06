package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/RTradeLtd/Pay/log"
	"github.com/RTradeLtd/Pay/queue"
	"github.com/RTradeLtd/Pay/server"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/cmd/v2"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2"
	pbSigner "github.com/RTradeLtd/grpc/pay"
)

// Version denotes the tag of this build
var Version string

const (
	closeMessage   = "press CTRL+C to stop processing and close queue resources"
	defaultLogPath = "/var/log/temporal/"
)

// globals
var (
	ctx    context.Context
	cancel context.CancelFunc
	signer pbSigner.SignerClient
)

// command-line flags
var (
	devMode    *bool
	configPath *string
	dbNoSSL    *bool
	dbMigrate  *bool
	grpcNoSSL  *bool
	apiPort    *string
)

func baseFlagSet() *flag.FlagSet {
	var f = flag.NewFlagSet("", flag.ExitOnError)

	// basic flags
	devMode = f.Bool("dev", false,
		"toggle dev mode")
	configPath = f.String("config", os.Getenv("CONFIG_DAG"),
		"path to Temporal configuration")

	// db configuration
	dbNoSSL = f.Bool("db.no_ssl", false,
		"toggle SSL connection with database")
	dbMigrate = f.Bool("db.migrate", false,
		"toggle whether a database migration should occur")

	// grpc configuration
	grpcNoSSL = f.Bool("grpc.no_ssl", false,
		"toggle SSL connection with GRPC services")

	// api configuration
	apiPort = f.String("api.port", "6767",
		"set port to expose API on")

	return f
}

func logPath(base, file string) (logPath string) {
	if base == "" {
		logPath = filepath.Join(base, file)
	} else {
		logPath = filepath.Join(base, file)
	}
	return
}

func newDB(cfg config.TemporalConfig, noSSL bool) (*gorm.DB, error) {
	dbm, err := database.New(&cfg, database.Options{LogMode: true, SSLModeDisable: noSSL})
	if err != nil {
		return nil, err
	}
	return dbm.DB, nil
}

var commands = map[string]cmd.Cmd{
	"queue": {
		Blurb:         "execute commands for various queues",
		Description:   "Interact with Temporal's various queue APIs",
		ChildRequired: true,
		Children: map[string]cmd.Cmd{
			"payment": cmd.Cmd{
				Blurb:         "payment queue sub commands",
				Description:   "Used to launch various payment queue processors",
				ChildRequired: true,
				Children: map[string]cmd.Cmd{
					"ethereum": cmd.Cmd{
						Blurb:       "Ethereum payment confirmation queue",
						Description: "Used to process and confirm ethereum/rtc based payments",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							logger, err := log.NewLogger(logPath(cfg.LogDir, "eth_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger", err)
								os.Exit(1)
							}
							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.EthPaymentConfirmationQueue, &cfg, logger, *devMode)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
					"dash": cmd.Cmd{
						Blurb:       "Dash payment confirmation queue",
						Description: "Used to process and confirm dash based payments",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							logger, err := log.NewLogger(logPath(cfg.LogDir, "dash_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger", err)
								os.Exit(1)
							}
							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.DashPaymentConfirmationQueue, &cfg, logger, true)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
					"bch": cmd.Cmd{
						Blurb:       "Bitcoin Cash payment confirmation queue",
						Description: "Used to process and confirm BCH payments",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							logger, err := log.NewLogger(logPath(cfg.LogDir, "bch_consumer.log"), *devMode)
							if err != nil {
								fmt.Println("failed to start logger", err)
								os.Exit(1)
							}
							db, err := newDB(cfg, *dbNoSSL)
							if err != nil {
								fmt.Println("failed to start db", err)
								os.Exit(1)
							}
							quitChannel := make(chan os.Signal)
							signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
							waitGroup := &sync.WaitGroup{}
							go func() {
								fmt.Println(closeMessage)
								<-quitChannel
								cancel()
							}()
							for {
								qm, err := queue.New(queue.BitcoinCashPaymentConfirmationQueue, &cfg, logger, *devMode)
								if err != nil {
									fmt.Println("failed to start queue", err)
									os.Exit(1)
								}
								waitGroup.Add(1)
								err = qm.ConsumeMessages(ctx, waitGroup, db, &cfg)
								if err != nil && err.Error() != queue.ErrReconnect {
									fmt.Println("failed to consume messages", err)
									os.Exit(1)
								} else if err != nil && err.Error() == queue.ErrReconnect {
									continue
								}
								// this will only be true if we had a graceful exit to the queue process, aka CTRL+C
								if err == nil {
									break
								}
							}
							waitGroup.Wait()
						},
					},
				},
			},
		},
	},
	"grpc": cmd.Cmd{
		Blurb:         "run gRPC API related commands",
		Description:   "allows running gRPC server and client",
		ChildRequired: true,
		Children: map[string]cmd.Cmd{
			"server": cmd.Cmd{
				Blurb:       "run the grpc server",
				Description: "runs our gRPC API server to generate signed messages",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					logger, err := log.NewLogger(logPath(cfg.LogDir, "pay_grpc_server.log"), *devMode)
					if err != nil {
						fmt.Println("failed to start logger", err.Error())
						os.Exit(1)
					}
					quitChannel := make(chan os.Signal)
					signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
					waitGroup := &sync.WaitGroup{}
					go func() {
						fmt.Println(closeMessage)
						<-quitChannel
						cancel()
					}()
					if err := server.RunServer(ctx, waitGroup, cfg, logger); err != nil {
						fmt.Println("an error occurred while running grpc server", err.Error())
						os.Exit(1)
					}
					waitGroup.Wait()
				},
			},
		},
	},
}

func main() {
	if Version == "" {
		Version = "latest"
	}

	// initialize global context
	ctx, cancel = context.WithCancel(context.Background())

	// create app
	temporal := cmd.New(commands, cmd.Config{
		Name:     "Temporal",
		ExecName: "temporal",
		Version:  Version,
		Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
		Options:  baseFlagSet(),
	})

	// run no-config commands, exit if command was run
	if exit := temporal.PreRun(nil, os.Args[1:]); exit == cmd.CodeOK {
		os.Exit(0)
	}

	// load config
	tCfg, err := config.LoadConfig(*configPath)
	if err != nil {
		println("failed to load config at", *configPath)
		os.Exit(1)
	}

	// load arguments
	flags := map[string]string{
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,
		"dbPass":        tCfg.Database.Password,
		"dbURL":         tCfg.Database.URL,
		"dbUser":        tCfg.Database.Username,
		"version":       Version,
	}

	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
