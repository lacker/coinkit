#!/bin/bash

# This script is the entry point for the Docker container, designed to be run on
# the Google cloud platform from the coinkit directory.

cserver \
    --keypair=./testnet/keypair0.json \
    --network=./testnet/network.json \
    --http=8000

