#!/bin/bash

echo rebuilding chost...
go install ./...

if [ $? -ne 0 ]
then
    echo "not running chost due to error"
    exit 1
fi

chost
