#!/bin/bash

if [ ! -f ./deployment.yaml ]; then
    echo "please run deploy.sh from the testnet directory"
    exit 1
fi

if [[ ! "$1" =~ [0-3] ]]; then
    echo "usage: ./deploy.sh n where n is in 0..3"
    exit 1
fi

APP=cserver$1
DB=db$1

CONNECTION_NAME=`gcloud sql instances describe $DB | grep connectionName | sed 's/connectionName: //'`

echo sql connection name: $CONNECTION_NAME

exit 0

sed s/PROJECT_ID/$PROJECT_ID/g ./deployment.yaml \
    | sed "s/cserverX/$APP/g" \
    | sed "s/dbX/$DB/g" \
    | sed "s/DEPLOY_TIME/`date`/" \
    | sed "s/CONNECTION_NAME/$CONNECTION_NAME/" \
    | kubectl apply -f -
