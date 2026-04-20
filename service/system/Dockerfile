FROM golang:1.26 AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY . /src
WORKDIR /src/service/system

RUN go mod download && make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/service/system/bin /app

WORKDIR /app

EXPOSE 8106
EXPOSE 9106
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
