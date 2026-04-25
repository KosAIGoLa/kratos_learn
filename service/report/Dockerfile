FROM golang:1.24 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn,direct go build -o ./bin/report ./cmd/report/

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin/report /app/
COPY --from=builder /src/configs /data/conf

WORKDIR /app

EXPOSE 8007
EXPOSE 9007

CMD ["./report", "-conf", "/data/conf"]
