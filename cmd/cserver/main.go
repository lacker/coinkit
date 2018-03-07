package main

import (
	"flag"
	"io/ioutil"
	"log"

	"coinkit/data"
	"coinkit/network"
	"coinkit/util"
)

// cserver runs a coinkit server.

func main() {
	var databaseFilename string
	var keyPairFilename string
	var networkFilename string

	flag.StringVar(&databaseFilename,
		"database", "", "the file to load database config from")
	flag.StringVar(&keyPairFilename,
		"keypair", "", "the file to load keypair config from")
	flag.StringVar(&networkFilename,
		"network", "", "the file to load network config from")
	flag.Parse()

	if databaseFilename == "" {
		log.Fatal("the --database flag must be set")
	}

	if keyPairFilename == "" {
		log.Fatal("the --keypair flag must be set")
	}

	if networkFilename == "" {
		log.Fatal("the --network flag must be set")
	}

	bytes, err := ioutil.ReadFile(databaseFilename)
	if err != nil {
		panic(err)
	}
	dbConfig := data.NewConfigFromSerialized(bytes)

	bytes, err = ioutil.ReadFile(keyPairFilename)
	if err != nil {
		panic(err)
	}
	kp := util.NewKeyPairFromSerialized(bytes)

	bytes, err = ioutil.ReadFile(networkFilename)
	if err != nil {
		panic(err)
	}
	net := network.NewConfigFromSerialized(bytes)

	db := data.NewDatabase(dbConfig)
	s := network.NewServer(kp, net, db)
	s.ServeForever()
}
