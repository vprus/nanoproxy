# nanoproxy
Minimal, but useful, reverse proxy

Nanoproxy proxies HTTP to another address, adding logging, and (soon) monitoring, authentication and
tracing on top. In many cases, it is much easier that read documentation of dozens of other
projects which all have different ways to configure the same features.

# Running

With docker, run 
```shell
docker -p 7070:7070 ghcr.io/vprus/nanoproxy:main
```
and then connect to `http://localhost:7070`

If you want to build yourself:
```shell
go build .
./nanoproxy
```
