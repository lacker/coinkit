FROM golang:1.9

RUN go get -v github.com/lacker/coinkit/...

WORKDIR /go/src/github.com/lacker/coinkit

RUN go install ./...

ENTRYPOINT [ "cserver", \
             "--keypair=./testnet/keypair0.json", \
	     "--network=./testnet/network.json", \
	     "--healthz=8000" ]

# Healthz
EXPOSE 8000

# The miner protocol
EXPOSE 9000
