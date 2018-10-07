package main

import (
	"errors"
	"log"
	"os"

	"github.com/RTradeLtd/config"
)

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
		generateServerAndList("127.0.0.1:9090", "tcp", cfg)
	case "client":
		generateClient("127.0.0.1:9090", false)
	default:
		err := errors.New("argument nto supported")
		log.Fatal(err)
	}

}
