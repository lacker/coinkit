#!/bin/bash

# This script is the entry point for a Docker container that runs the node hosting server.
# It is designed to be run on the Google cloud platform from the coinkit/js directory.

echo ------------------------------ host-entrypoint.sh ------------------------------

KEYPAIR=`find /secrets/keypair | grep json | head -1`
echo loading keypair: $KEYPAIR

# TODO: make stuff work

npm run host
