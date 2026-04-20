FROM golang:1.26 AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY . /src
WORKDIR /src/service/order

RUN go mod download && make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/service/order/bin /app

WORKDIR /app

EXPOSE 8103
EXPOSE 9103
VOLUME /data/conf

CMD ["./order", "-conf", "/data/conf"]
