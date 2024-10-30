FROM golang:1.23.2 AS build

WORKDIR /src

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /proxy .

FROM ubuntu AS bin

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

COPY --from=build /proxy /proxy
ENTRYPOINT ["/proxy"]
