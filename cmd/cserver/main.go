package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/logging"

	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/network"
	"github.com/lacker/coinkit/util"
)

// cserver runs a coinkit server.

func main() {
	var databaseFilename string
	var keyPairFilename string
	var networkFilename string
	var logProject string
	var logName string
	var httpPort int

	flag.StringVar(&databaseFilename,
		"database", "", "optional. the file to load database config from")
	flag.StringVar(&keyPairFilename,
		"keypair", "", "the file to load keypair config from")
	flag.StringVar(&networkFilename,
		"network", "", "the file to load network config from")
	flag.IntVar(&httpPort, "http", 0, "the port to serve /healthz etc on")
	flag.StringVar(&logProject, "logproject", "", "the Google Cloud project to log to")
	flag.StringVar(&logName, "logname", "cserver-log",
		"the Google Cloud log name to log to")
	flag.Parse()

	if logProject != "" {
		client, err := logging.NewClient(context.Background(), logProject)
		if err != nil {
			util.Logger.Fatal("Failed to create logging client: %+v", err)
		}
		defer client.Close()
		util.Logger = client.Logger(logName).StandardLogger(logging.Info)
		util.LogType = fmt.Sprintf("projects/%s/logs/%s", logProject, logName)
	}

	if keyPairFilename == "" {
		util.Logger.Fatal("the --keypair flag must be set")
	}

	if networkFilename == "" {
		util.Logger.Fatal("the --network flag must be set")
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
