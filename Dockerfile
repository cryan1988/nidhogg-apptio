# Build the manager binary
FROM golang:1.21 as builder

# Copy in the go src
WORKDIR /go/src/github.com/uswitch/nidhogg
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
ENV GO111MODULE=auto
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/uswitch/nidhogg/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:23.10
WORKDIR /
COPY --from=builder /go/src/github.com/uswitch/nidhogg/manager .
ENTRYPOINT ["/manager"]
