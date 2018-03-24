FROM golang:1.9

ARG PROJECT_ID

RUN go get -v github.com/lacker/coinkit/...

WORKDIR /go/src/github.com/lacker/coinkit

RUN go install ./...

ENTRYPOINT [ "cserver", \
             "--keypair=./testnet/keypair0.json", \
	     "--network=./testnet/network.json", \
	     "--http=8000", \
	     "--project=$PROJECT_ID" ]

# Healthz
EXPOSE 8000

# The miner protocol
EXPOSE 9000
