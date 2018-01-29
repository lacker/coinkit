#!/bin/bash

LOGS="$HOME/logs"

if [ ! -d "$LOGS" ]; then
    echo "please create a logs directory in ~/logs"
    exit 1
fi

RUNNING=`pgrep ^cserver`
if [ -n "$RUNNING" ]
then
    echo "there are already cservers running:"
    ps aux | grep [^a-z]cserver | grep -v grep
    exit 1
fi

go install ./...

for i in `seq 0 3`;
do
    nohup cserver $i &> $LOGS/cserver$i.log &
done 

sleep 0.1
ps aux | grep [^a-z]cserver | grep -v grep
