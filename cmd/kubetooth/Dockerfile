FROM golang:1.16 AS builder

LABEL maintainer="Raphael Santo Domingo <raphael.santodomingo@target.com>"

ARG BIN
ARG VERSION
ARG GIT_COMMIT

ENV BIN=${BIN}
ENV VERSION=${VERSION}
ENV GIT_COMMIT=${GIT_COMMIT}

ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Create a location in the container for the source code.
RUN mkdir -p /app

# Copy the module files first and then download the dependencies. If this
# doesn't change, we won't need to do this again in future builds.
COPY go.* /app/

WORKDIR /app

# Uitlize non-breaking Kubernetes go-client latest
RUN go get k8s.io/client-go@latest

RUN go mod download
RUN go mod verify

# Copy the source code into the container.
COPY pkg pkg
COPY models models
COPY cmd/"$BIN" cmd/"$BIN"
# # Copy required proto files
# COPY cmd/auth/proto cmd/auth/proto

RUN go build \
    -mod=readonly \
    -ldflags "-X pa3ng/kubetooth/pkg/buildinfo.Version=$VERSION -X pa3ng/kubetooth/pkg/buildinfo.GitCommit=$GIT_COMMIT -X 'pa3ng/kubetooth/pkg/buildinfo.BuildTime=$(date -u '+%Y-%m-%d %H:%M:%S')'" \
    -a -o /go/bin/app ./cmd/"$BIN"

FROM scratch
COPY --from=builder /go/bin/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]