# nanoproxy
Minimal, but useful, reverse proxy

Nanoproxy proxies HTTP to another address, adding logging, and (soon) monitoring, authentication and
tracing on top. In many cases, it is much easier that read documentation of dozens of other
projects which all have different ways to configure the same features.

# Running Locally

With docker, run 
```shell
docker run -p 7070:7070 ghcr.io/vprus/nanoproxy:main
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

# Monitoring

The proxy can export Prometheus-compatible metrics on port 9090. For example, if you run
```shell
docker run -p 7070:7070 -p 9090:9090 ghcr.io/vprus/nanoproxy:main
```

You can first connect to `http://localhost:7070` and then review metrics at `http://localhost:9090/metrics`

# Tracining

The proxy can also export traces to a tracing service that supports the Open Telemetry protocol. We'll use our
old friend Jaeger. First, start it using the official recipe:

```shell
docker run --rm --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.62.0
```

Then, run nanoproxy, passing the path to jaeger:

```
docker run -p 7070:7070 -e OTEL_EXPORTER_OTLP_ENDPOINT=http://host.docker.internal:4318  ghcr.io/vprus/nanoproxy:main
```

Then, make a request to `http://localhost:7070`, go the Jaeger UI at `http://localhost:16686/`, select `nanoproxy` in
the 'service' dropdown, and view the traces.

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
