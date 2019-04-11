#!/bin/bash

if [ ! -f ./deployment.yaml ]; then
    echo "please run deploy.sh from the testnet directory"
    exit 1
fi

if (( "$1" < 0 )) || (( "$1" > 3 )); then
    echo "usage: ./deploy.sh n where n is in 0..3"
    exit 1
fi

CSERVER=cserver$1
HSERVER=hserver$1
DB=db$1
KEYPAIR=keypair$1

CONNECTION_NAME=`gcloud sql instances describe $DB | grep connectionName | sed 's/connectionName: //'`

echo sql connection name: $CONNECTION_NAME

sed s/PROJECT_ID/$PROJECT_ID/g ./deployment.yaml \
    | sed "s/cserverX/$CSERVER/g" \
    | sed "s/hserverX/$HSERVER/g" \
    | sed "s/dbX/$DB/g" \
    | sed "s/keypairX/$KEYPAIR/g" \
    | sed "s/DEPLOY_TIME/`date`/" \
    | sed "s/CONNECTION_NAME/$CONNECTION_NAME/" \
    | kubectl apply -f -
