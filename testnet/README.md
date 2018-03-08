# Running a testnet

This directory contains operational tools for running the alpha testnet.

This instructions specifically explain how to deploy a miner to the Google Cloud Platform.

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

