package main

import (
	"log"
	"os"
	"strconv"

	"coinkit/data"
	"coinkit/network"
)

// cserver runs a coinkit server.

func usage() {
	log.Fatal("Usage: cserver <i> where i is in [0, 1, 2, 3]")
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	arg, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if arg < 0 || arg > 3 {
		usage()
	}

	config, kps := network.NewLocalNetwork()
	db := data.NewLocalDatabase(arg)
	s := network.NewServer(kps[arg], config, db)
	s.ServeForever()
}
