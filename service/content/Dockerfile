FROM golang:1.26 AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY . /src
WORKDIR /src/service/content

RUN go mod download && make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/service/content/bin /app

WORKDIR /app

EXPOSE 8101
EXPOSE 9101
VOLUME /data/conf

CMD ["./content", "-conf", "/data/conf"]
