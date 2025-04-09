# Build the manager binary
FROM golang:1.22.12 as builder

# Copy in the go src
WORKDIR /go/src/github.com/uswitch/nidhogg
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
ENV GO111MODULE=auto
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -a -o manager github.com/uswitch/nidhogg/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:22.04
RUN apt-get update && apt-get upgrade -y && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /go/src/github.com/uswitch/nidhogg/manager .
ENTRYPOINT ["/manager"]
