FROM golang:1.9

ARG PROJECT_ID
ENV PROJECT_ID $PROJECT_ID

RUN go get -v github.com/lacker/coinkit/...

WORKDIR /go/src/github.com/lacker/coinkit

RUN go install ./...

ENTRYPOINT ./testnet/entrypoint.sh

# Healthz
EXPOSE 8000

# The miner protocol
EXPOSE 9000
