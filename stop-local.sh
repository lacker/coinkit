#!/bin/bash

pgrep '[^a-z]cserver' | xargs kill -9
LEFT=`ps aux | grep '[^a-z]cserver' | grep -v grep`
if [ -n "$LEFT" ]
then
    echo "could not kill:"
    echo $LEFT
fi
