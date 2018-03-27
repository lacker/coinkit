#!/bin/bash

if [ ! -f ./deployment.yaml ]; then
    echo "please run deploy.sh from the testnet directory"
    exit 1
fi

CONNECTION_NAME=`gcloud sql instances describe db0 | grep connectionName | sed 's/connectionName: //'`

echo sql connection name: $CONNECTION_NAME

sed s/PROJECT_ID/$PROJECT_ID/g ./deployment.yaml \
    | sed "s/DEPLOY_TIME/`date`/" \
    | sed "s/CONNECTION_NAME/$CONNECTION_NAME/" \
    | kubectl apply -f -
