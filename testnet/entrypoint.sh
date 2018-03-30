#!/bin/bash

# This script is the entry point for the Docker container, designed to be run on
# the Google cloud platform from the coinkit directory.

echo "contents of /secrets/keypair:"
ls /secrets/keypair
KEYPAIR=`ls /secrets/keypair | grep json | head -1`

cserver \
    --keypair=$KEYPAIR \
    --network=./testnet/network.json \
    --logtostdout \
    --http=8000

