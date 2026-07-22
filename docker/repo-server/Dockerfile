FROM golang:1.24-alpine3.20 AS builder

WORKDIR /usr/src
COPY go.mod /usr/src/go.mod
COPY go.sum /usr/src/go.sum
RUN go mod download

COPY cmd cmd
COPY internal internal

WORKDIR /usr/src/cmd/repo-server
RUN go build -o server -ldflags="-s -w" .

FROM alpine:3.20
RUN apk add --no-cache git openssh-client \
    && adduser -D -u 1000 -h /home/gitops gitops
WORKDIR /usr/bin
COPY --from=builder /usr/src/cmd/repo-server/server .
USER gitops
ENV HOME=/home/gitops
ENTRYPOINT ["/usr/bin/server"]
EXPOSE 8080