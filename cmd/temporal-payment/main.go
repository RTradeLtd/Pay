package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	//_ "./docs"

	"github.com/RTradeLtd/Temporal_Payment-ETH/cmd/temporal-payment/app"
	"github.com/RTradeLtd/Temporal_Payment-ETH/gapi/server"
	"github.com/RTradeLtd/Temporal_Payment-ETH/queue"
	"github.com/RTradeLtd/config"
)

var (
	// Version denotes the tag of this build
	Version string

	certFile = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile  = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg     config.TemporalConfig
)

var commands = map[string]app.Cmd{
	"queue": app.Cmd{
		Blurb:         "execute commands for various queues",
		Description:   "Interact with Temporal's various queue APIs",
		ChildRequired: true,
		Children: map[string]app.Cmd{
			"payment": app.Cmd{
				Blurb:         "Payment queue sub commands",
				Description:   "Used to launch the various queues that interact with our payment backend",
				ChildRequired: true,
				Children: map[string]app.Cmd{
					"payment-confirmation": app.Cmd{
						Blurb:       "Payment confirmation queue",
						Description: "Listens to, and process payment confirmation messages",
						Action: func(cfg config.TemporalConfig, args map[string]string) {
							mqConnectionURL := cfg.RabbitMQ.URL
							fmt.Println("initializing queue")
							qm, err := queue.Initialize(queue.PaymentConfirmationQueue, mqConnectionURL, false, true)
							if err != nil {
								log.Fatal(err)
							}
							fmt.Println("consuming messages")
							err = qm.ConsumeMessage("", args["dbPass"], args["dbURL"], args["dbUser"], &cfg)
							if err != nil {
								log.Fatal(err)
							}
						},
					},
				},
			},
		},
	},
	"gapi": app.Cmd{
		Blurb:         "run gRPC API related commands",
		Description:   "allows running gRPC server and client",
		ChildRequired: true,
		Children: map[string]app.Cmd{
			"server": app.Cmd{
				Blurb:       "run the gapi server",
				Description: "runs our gRPC API server to generate signed messages",
				Action: func(cfg config.TemporalConfig, args map[string]string) {
					if err := server.RunServer("0.0.0.0:9090", "tcp", &cfg); err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	},
}

func main() {
	// create app
	temporal := app.New(commands, app.Config{
		Name:     "Temporal",
		ExecName: "temporal",
		Version:  Version,
		Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
	})

	// run no-config commands
	exit := temporal.PreRun(os.Args[1:])
	if exit == app.CodeOK {
		os.Exit(0)
	}

	// load config
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		log.Fatal("CONFIG_DAG is not set")
	}
	tCfg, err := config.LoadConfig(configDag)
	if err != nil {
		log.Fatal(err)
	}

	// load arguments
	flags := map[string]string{
		"configDag":     configDag,
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,

		"dbPass": tCfg.Database.Password,
		"dbURL":  tCfg.Database.URL,
		"dbUser": tCfg.Database.Username,
	}

	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
