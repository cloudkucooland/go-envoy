package main

import (
	"flag"
	"fmt"
	"github.com/cloudkucooland/go-envoy"
	// "strconv"
)

func main() {
	var command, host, value string
	// debug := flag.Bool("d", false, "debug")

	flag.Parse()
	args := flag.Args()
	argc := len(args)
	if argc == 0 {
		command = "unset"
	}
	if argc >= 1 {
		command = args[0]
	}
	if argc > 1 {
		host = args[1]
	}
	if argc > 2 {
		value = args[2]
	}

	var e *envoy.Envoy
	if host != "" {
		var err error
        e, err = envoy.New(host)
		if err != nil {
			panic(err)
		}
	}

	switch command {
	case "pull":
		s, err := e.Pull()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", s)
	default:
        fmt.Printf("Valid commands: pull (sent: %s %s %s)\n", command, host, value)
	}
}
