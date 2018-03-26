package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/network"
	"github.com/lacker/coinkit/util"
)

// cserver runs a coinkit server.

func main() {
	var databaseFilename string
	var keyPairFilename string
	var networkFilename string
	var httpPort int
	var logToStdOut bool

	flag.StringVar(&databaseFilename,
		"database", "", "optional. the file to load database config from")
	flag.StringVar(&keyPairFilename,
		"keypair", "", "the file to load keypair config from")
	flag.StringVar(&networkFilename,
		"network", "", "the file to load network config from")
	flag.IntVar(&httpPort, "http", 0, "the port to serve /healthz etc on")
	flag.BoolVar(&logToStdOut, "logtostdout", false, "whether to log to stdout")

	flag.Parse()

	if keyPairFilename == "" {
		util.Logger.Fatal("the --keypair flag must be set")
	}

	if networkFilename == "" {
		util.Logger.Fatal("the --network flag must be set")
	}

	if logToStdOut {
		util.Logger = log.New(os.Stdout, "", log.LstdFlags)
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
	if httpPort != 0 {
		s.ServeHttpInBackground(httpPort)
	}
	s.ServeForever()
}
