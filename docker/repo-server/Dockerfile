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
RUN apk add --no-cache git
WORKDIR /usr/bin
COPY --from=builder /usr/src/cmd/repo-server/server .
ENTRYPOINT ["/usr/bin/server"]
EXPOSE 8080