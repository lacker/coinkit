package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/network"
	"github.com/lacker/coinkit/util"
)

// cserver runs a coinkit server.

func main() {
	var databaseFilename string
	var keyPairFilename string
	var networkFilename string
	var healthz int

	flag.StringVar(&databaseFilename,
		"database", "", "optional. the file to load database config from")
	flag.StringVar(&keyPairFilename,
		"keypair", "", "the file to load keypair config from")
	flag.StringVar(&networkFilename,
		"network", "", "the file to load network config from")
	flag.IntVar(&healthz, "healthz", 0, "the port to serve /healthz on")
	flag.Parse()

	if keyPairFilename == "" {
		log.Fatal("the --keypair flag must be set")
	}

	if networkFilename == "" {
		log.Fatal("the --network flag must be set")
	}

	var db *data.Database
	if databaseFilename != "" {
		bytes, err := ioutil.ReadFile(databaseFilename)
		if err != nil {
			panic(err)
		}
		dbConfig := data.NewConfigFromSerialized(bytes)
		db = data.NewDatabase(dbConfig)
	}

	bytes, err := ioutil.ReadFile(keyPairFilename)
	if err != nil {
		panic(err)
	}
	kp := util.NewKeyPairFromSerialized(bytes)

	bytes, err = ioutil.ReadFile(networkFilename)
	if err != nil {
		panic(err)
	}
	net := network.NewConfigFromSerialized(bytes)

	s := network.NewServer(kp, net, db)
	if healthz != 0 {
		s.ServeHttpInBackground(healthz)
	}
	s.ServeForever()
}
