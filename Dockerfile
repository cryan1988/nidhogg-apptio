# Build the manager binary
FROM golang:1.22.12 as builder

# Copy in the go src
WORKDIR /go/src/github.com/uswitch/nidhogg
COPY . .

# Build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -tags netgo -ldflags="-extldflags=-static" -o manager github.com/uswitch/nidhogg/cmd/manager

# Copy the controller-manager into a thin image
FROM --platform=linux/amd64 ubuntu:22.04
RUN apt-get update && apt-get upgrade -y && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /go/src/github.com/uswitch/nidhogg/manager .
ENTRYPOINT ["/manager"]
