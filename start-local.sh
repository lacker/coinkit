#!/bin/bash

LOGS="$HOME/logs"

if [ ! -d "$LOGS" ]; then
    echo "please create a logs directory in ~/logs"
    exit 1
fi

RUNNING=`ps aux | grep ^cserver`
if [ -n "$RUNNING" ]
then
    echo "there are already cservers running:"
    echo $RUNNING
    exit 1
fi

go install ./...

for i in `seq 0 3`;
do
    nohup cserver $i &> $LOGS/cserver$i.log &
done 
