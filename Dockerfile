FROM golang:1.18 AS builder

ENV GOPROXY=https://goproxy.cn,direct \
    GO111MODULE=on

WORKDIR /tmp/app

COPY . .

RUN go mod download all && go build -tags netgo -o /LarkGPT

FROM alpine:3.13.1

COPY config.yaml /etc/LarkGPT/config.yaml

COPY --from=builder /LarkGPT /LarkGPT

RUN mkdir -p /log

ENTRYPOINT ["/LarkGPT"]