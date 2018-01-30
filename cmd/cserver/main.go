package main

import (
	"log"
	"os"
	//"runtime/pprof"
	"strconv"

	"coinkit/network"
)

// cserver runs a coinkit server.

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: cserver <i> where i is in [0, 1, 2, ..., NumPeers - 1]")
	}
	arg, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// CPU profiling
	/*
	if arg == 0 {
		f, err := os.Create("./profile")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()		
	}
    */
	
	config := network.NewLocalConfig(arg)
	
	s := network.NewServer(config)
	s.ServeForever()
}
