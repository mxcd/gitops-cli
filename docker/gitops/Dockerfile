FROM golang:1.24-alpine3.20 AS builder
RUN apk add --no-cache git

ARG GITOPS_VERSION="development"

WORKDIR /usr/src
COPY go.mod /usr/src/go.mod
COPY go.sum /usr/src/go.sum
RUN go mod download

COPY cmd cmd
COPY internal internal

WORKDIR /usr/src/cmd/gitops
RUN go build -o gitops -ldflags="-s -w 'main.version=${GITOPS_VERSION}'" .

FROM alpine:3.20
WORKDIR /usr/bin
COPY --from=builder /usr/src/cmd/gitops/gitops .
ENTRYPOINT ["/usr/bin/gitops"]
CMD ["--help"]