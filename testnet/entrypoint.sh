#!/bin/bash

# This script is the entry point for the Docker container, designed to be run on
# the Google cloud platform from the coinkit directory.

cserver \
    --keypair=/secrets/keypair.json \
    --network=./testnet/network.json \
    --logtostdout \
    --http=8000

