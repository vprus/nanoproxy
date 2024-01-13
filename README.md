# nanoproxy
Minimal, but useful, reverse proxy

Nanoproxy proxies HTTP to another address, adding logging, and (soon) monitoring, authentication and
tracing on top. In many cases, it is much easier that read documentation of dozens of other
projects which all have different ways to configure the same features.

# Running Locally

With docker, run 
```shell
docker -p 7070:7070 ghcr.io/vprus/nanoproxy:main
```
and then connect to `http://localhost:7070`, which will proxy your request to `https://example.com`.


You can change the proxy target with the `-target` option:
```shell
docker run -p 7070:7070 ghcr.io/vprus/nanoproxy:main -target=https://example.net
```


Finally, you can proxy to another port on your local machine using a magic Docker URL:

```shell
docker run -p 7070:7070 ghcr.io/vprus/nanoproxy:main -target=http://host.docker.internal:8080
```


If you want to build yourself:
```shell
go build .
./nanoproxy
```

# Running on Google Cloud with Kubernetes

Kubernetes manifest that work for GCP GKE are provided.

If you already have a cluster, please make sure that you have the `gcloud` tool installed, that
you have `kubectl` installed, and that you switched to your cluster, using a command such as
```shell
gcloud container clusters get-credentials <cluster> --location <location>
```

If you don't have a cluster yet, you can [create one](https://cloud.google.com/kubernetes-engine/docs/how-to/creating-an-autopilot-cluster).

With prerequisites in place, just run
```shell
kubectl apply -f k8s-gke.yaml
```

It will create a namespace `nanoproxy`, a deployment there, and an ingress. After a few minutes, you can
find the IP address of the ingress:
```shell
kubectl -n nanoproxy get ingress
```
and connect to that address, port 80.

After you're done, don't forget to remove all resources to avoid further charges:

```shell
kubectl delete -f k8s-gke.yaml
```

# Running on AWS with Kubernetes

Kubernetes manifest that work for AWS EKS are provided.

If you already have a cluster, please make sure that you have the AWS CLI installed, that
you have `kubectl` installed, and that you switched to your cluster, using a command such as
```shell
aws eks update-kubeconfig --name <cluster>
```

If you don't have a cluster yet, you can [create one](https://docs.aws.amazon.com/eks/latest/userguide/getting-started.html).

With prerequisites in place, just run
```shell
kubectl apply -f k8s-eks.yaml
```

It will create a namespace `nanoproxy`, a deployment there, and an ingress. After a few minutes, you can
find the domain name of the  of the ingress:
```shell
kubectl -n nanoproxy get ingress
```
and connect to that domain name, port 80.

After you're done, don't forget to remove all resources to avoid further charges:

```shell
kubectl delete -f k8s-eks.yaml
```
