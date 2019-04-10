#!/bin/bash

if [[ $(git diff) ]]; then
    echo "build-cserver.sh creates a container from master. check in your changes before building."
    exit 1
fi

if [[ $(git status | grep "branch is ahead") ]]; then
    echo "build-cserver.sh creates a container from master. push your changes before building."
    exit 1
fi

if [ ! -f ./deployment.yaml ]; then
    echo "please run build-cserver.sh from the testnet directory"
    exit 1
fi

# The `--no-cache` is needed because the build process grabs fresh code from GitHub, and
# if you enable the cache it'll keep using your old code.
echo building Docker image...
docker build \
       --no-cache \
       -t gcr.io/${PROJECT_ID}/cserver \
       --file ./cserver-Dockerfile \
       .

# Upload it to Google's container registry
echo uploading Docker image to Google\'s container registry...
docker push gcr.io/${PROJECT_ID}/cserver
