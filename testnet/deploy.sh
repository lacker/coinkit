#!/bin/bash

if [ ! -f ./deployment.yaml ]; then
    echo "please run deploy.sh from the testnet directory"
    exit 1
fi

sed s/PROJECT_ID/$PROJECT_ID/g ./deployment.yaml \
    | sed "s/THE_DEPLOY_TIME/`date`/" \
    | kubectl apply -f -
