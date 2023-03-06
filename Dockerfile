FROM golang:1.20-alpine3.17 AS builder
RUN apk add --no-cache git

WORKDIR /usr/src
COPY go.mod /usr/src/go.mod
COPY go.sum /usr/src/go.sum
RUN go mod download

COPY . .
ENV SOPS_AGE_KEY_FILE=/usr/src/test/keys.txt
RUN go test ./...

RUN go build -o gitops -ldflags="-s -w" cmd/gitops/main.go

FROM alpine:3.17
WORKDIR /usr/bin
COPY --from=builder /usr/src/gitops .
ENTRYPOINT ["/usr/bin/gitops"]
CMD ["--help"]