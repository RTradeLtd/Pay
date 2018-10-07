package main

import (
	"errors"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 2 || len(os.Args) < 2 {
		err := errors.New("invalid invocation, ./gapi <server>")
		log.Fatal(err)
	}
	switch os.Args[1] {
	case "server":
		generateServerAndList("127.0.0.1:9090", "tcp")
	case "client":
		generateClient("127.0.0.1:9090", false)
	default:
		err := errors.New("argument nto supported")
		log.Fatal(err)
	}

}
