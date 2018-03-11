# Running a testnet

This directory contains operational tools for running the alpha testnet.

This instructions specifically explain how to deploy a miner to the Google Cloud Platform.

Estimated cost for keeping one miner running using these instructions:
* n1-standard-1 for the app servers is $25 a month
* db-f1-micro for the database is $8 a month
* 100 GB of storage is another $9 a month

### 1. Set up a GCP account and install the Cloud Tools

https://cloud.google.com/sdk/docs/

```
$ gcloud version
Google Cloud SDK 192.0.0
bq 2.0.29
core 2018.03.02
gsutil 4.28
```

Also use `gcloud` to install Kubernetes:

```
gcloud components install kubectl
```

My GCP project was named "coinkitalpha" and it seemed like Iowa "A" was the best place, so
I set the `gcloud` defaults with:

```
gcloud config set project coinkitalpha
gcloud config set compute/zone us-central1-a
```

It is handy to have `PROJECT_ID` set to the name of your GCP project in your shell,
so add this to your bash config:

```
export PROJECT_ID="$(gcloud config get-value project -q)"
```

### 2. Install Docker

https://www.docker.com/community-edition

I went for "Docker CE for Mac (Stable)".

```
$ docker version
Client:
 Version:	17.12.0-ce
 API version:	1.35
 Go version:	go1.9.2
 Git commit:	c97c6d6
 Built:	Wed Dec 27 20:03:51 2017
 OS/Arch:	darwin/amd64

Server:
 Engine:
  Version:	17.12.0-ce
  API version:	1.35 (minimum version 1.12)
  Go version:	go1.9.2
  Git commit:	c97c6d6
  Built:	Wed Dec 27 20:12:29 2017
  OS/Arch:	linux/amd64
  Experimental:	true
```

### 3. Make a container image

From the `testnet` directory, first you need to build it:

```
docker build --no-cache -t gcr.io/${PROJECT_ID}/cserver .
```

The `--no-cache` is needed because the build process grabs fresh code from GitHub, and
if you enable the cache it'll keep using your old code.

TODO: right now this only builds a miner with one hardcoded set of credentials. I need
to find a way to pass these credentials in.
TODO: this also does not connect to a database, and it should

Then upload it to Google's container registry:

```
gcloud docker -- push gcr.io/${PROJECT_ID}/cserver
```

### 4. Start running stuff on your cluster

First, let's make a cluster named "testnet". Once you run this, it'll
start charging you money. A standard node is about $25 a month.

```
gcloud container clusters create testnet --num-nodes=1
```

To deploy a `cserver` to your cluster, run:

```
./deploy.sh
```

This same command should also update the deployment, when a new
"latest" image exists or when the yaml file has been updated.

TODO: change below here, the load balancer is needless. Try setting up
service.yaml somehow instead


XXX TRY:
kubectl apply -f ./service.yaml

XXX BUT:
i'm pretty sure the labels are wrong. I don't know what they are
supposed to be for, but I don't see how the service can attach to the deployment.

# XXX things dont work below here maybe

Now we expose our deployment to the internet with a
load balancer service named `loadbalancer`:

```
kubectl expose deployment cserver-deployment --type=LoadBalancer --name=loadbalancer
```

To see what its external IP is:

```
kubectl get services loadbalancer
```

It might take a couple minutes to bind to your external IP. Once it binds,
check `your.ip.address:8000/healthz` in your browser. If it says `OK`,
you are successfully running a cserver.

To make this a static IP address, the easiest way is to go to
https://console.cloud.google.com/projectselector/networking/addresses/list
and use the GUI to change it to "static". It's the one listed as a
"forwarding rule". I couldn't figure out how to
do this from the CLI, but I'm sure there is a way :P
