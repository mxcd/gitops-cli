FROM golang:1.24-alpine AS builder

RUN apk add --no-cache \
  build-base \
  cmake \
  openssl-dev \
  libssh2-dev \
  zlib-dev \
  curl \
  git

WORKDIR /tmp
RUN curl -L https://github.com/libgit2/libgit2/archive/refs/tags/v1.5.0.tar.gz | tar xz && \
    cd libgit2-1.5.0 && \
    cmake . -DBUILD_SHARED_LIBS=OFF -DBUILD_TESTS=OFF -DCMAKE_INSTALL_PREFIX=/usr && \
    cmake --build . --target install

WORKDIR /go/src/github.com/libgit2
RUN git clone https://github.com/libgit2/git2go.git && \
    cd git2go && \
    git checkout v34.0.0 && \
    git submodule update --init

ENV CGO_CFLAGS="-I/usr/include"
ENV CGO_LDFLAGS="-L/usr/lib -lz -lssl -lcrypto"
ENV LD_LIBRARY_PATH="/usr/lib"

WORKDIR /usr/src
COPY go.mod /usr/src/go.mod
COPY go.sum /usr/src/go.sum
RUN go mod download

COPY cmd cmd
COPY internal internal

WORKDIR /usr/src/cmd/repo-server
RUN go build -o repo-server -ldflags="-s -w" .

FROM alpine:3.21
WORKDIR /usr/bin
COPY --from=builder /usr/src/cmd/repo-server/repo-server .
ENTRYPOINT ["/usr/bin/repo-server"]
EXPOSE 8080