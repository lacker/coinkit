#!/bin/bash

if [ ! -f ./deployment.yaml ]; then
    echo "please run build.sh from the testnet directory"
    exit 1
fi

# The `--no-cache` is needed because the build process grabs fresh code from GitHub, and
# if you enable the cache it'll keep using your old code.
echo building container...
docker build \
       --no-cache \
       -t gcr.io/${PROJECT_ID}/cserver \
       .

# Upload it to Google's container registry
echo uploading to Google\'s container registry...
gcloud docker --verbosity=error -- push gcr.io/${PROJECT_ID}/cserver
