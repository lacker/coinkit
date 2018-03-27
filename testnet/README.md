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

Choose a name for your gcloud coinkit project and create it:

```
gcloud projects create your-coinkit-project-name
gcloud config set project your-coinkit-project-name
```

It is handy to have `PROJECT_ID` set to the name of your GCP project in your shell,
so add this to your bash config and source it:

```
export PROJECT_ID="$(gcloud config get-value project -q)"
```

It seemed like Iowa "A" was the best place, so I set the `gcloud` defaults with:

```
gcloud config set compute/zone us-central1-a
```

Enable billing for your project: https://cloud.google.com/billing/docs/how-to/modify-project

Add `Kubernetes Engine` and `Container Registry` API access to your project:

```
gcloud services enable container.googleapis.com
gcloud services enable containerregistry.googleapis.com
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

The build process takes a snapshot of the latest code on `github.com/lacker/coinkit`,
creates a Docker image from that, and uses that to deploy. So get your changes into
master before trying to deploy them.

From the `testnet` directory, build a container and upload it to Google's container
registry with:

```
./build.sh
```

The container and its presence on the registry is specific to your project, so this
won't interfere with other peoples' builds.

### 4. Start your cluster

You will need to specifically enable some APIs.

Enable logging at: https://console.cloud.google.com/flows/enableapi?apiid=logging.googleapis.com

Enable SQL at: https://console.cloud.google.com/flows/enableapi?apiid=sqladmin

Then let's make the cluster, named "testnet". Once you run this, it'll
start charging you money. The setup here is pretty dinky, and it should be about $35 a month, but proceed at your own risk.

```
gcloud container clusters create testnet --num-nodes=1 --scopes https://www.googleapis.com/auth/logging.write
```

You will also need a firewall opening ports to the outside world:

```
gcloud compute firewall-rules create cfirewall --allow tcp:30800,tcp:30900
```

### 5. Start a database

Create a new database instance at https://console.cloud.google.com/projectselector/sql/instances . Pick postgres. Name it `db0` - that is your "instance name".

Generate a random password, but take note of it.

I edited the resources to be the minimum, 1 shared cpu and 0.6 GB memory.

You need a "service account" for this database. Create one at https://console.cloud.google.com/projectselector/iam-admin/serviceaccounts

Create a service account with the "Cloud SQL Client" role. Name it `sql-client` and select "Furnish a new private key" using `JSON` type. Hang on to the json file that your browser downloads.

Now you need to create a proxy user. For the database `db0` name the user `proxyuser0`.
Use that password you noted when you created the database instance.

```
gcloud sql users create proxyuser0 host --instance=db0 --password=[PASSWORD]
```

Now we need to create some Kubernetes secrets. Both the service account and the proxy user require secrets to use them. The service account can be shared among multiple databases, but the proxy user is tied to a specific database.

To create a secret for the service account, named `cloudsql-instance-credentials`:

```
kubectl create secret generic cloudsql-instance-credentials --from-file=credentials.json=that-json-file-you-downloaded.json
```

For the proxy user, create a secret named `cloudsql-db0-credentials` with:

```
kubectl create secret generic cloudsql-db0-credentials --from-literal=username=proxyuser0 --from-literal=password=[PASSWORD]
```

### 6. Deploy a cserver to your cluster

To deploy a `cserver` to your cluster, run:

```
./deploy.sh
```

This same command should also update the deployment, when a new
"latest" image exists or when the yaml file has been updated.

To expose the `cserver` to public internet ports, you need to create a kubernetes service
and a firewall:

```
kubectl apply -f ./service.yaml
gcloud compute firewall-rules create cfirewall --allow tcp:30800,tcp:30900
```

To find the external ip, run

```
gcloud compute instances list
```

Then go to `your.external.ip:30800/healthz` in the browser. You should see an `OK`.
Port `30800` is where status information is, port `30900` runs the peer-to-peer protocol.

### 7. Updating the server

When you've updated the code, just rebuild a container image and redeploy.

```
./build.sh
./deploy.sh
```

### 8. Running more servers

TODO: explain how to run more than a single node

### 9. Cleaning up

If you don't want to keep things running, you can shut down the deployment, the service,
and the cluster itself:

```
kubectl delete service cservice0
kubectl delete deployment cserver0-deployment
gcloud container clusters delete testnet
```

You can leave the firewall rules running for free.

You can delete the database from the UI, but be aware that you can't recreate one with the same name for a week or so.