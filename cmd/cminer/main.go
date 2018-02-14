package main

import (
	"coinkit/network"
)

func main() {
	config := network.NewMinerConfig()
	s := network.NewServer(config)
	s.InitMint()
	fs := network.NewFileServer(s.DataStore())
	go fs.ServeForever(7777)
	s.ServeForever()
}
