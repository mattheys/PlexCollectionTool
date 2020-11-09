# Multi-stage build setup (https://docs.docker.com/develop/develop-images/multistage-build/)

# Stage 1 (to create a "build" image, ~850MB)
FROM golang:1.15 AS builder
RUN go version
RUN git clone https://github.com/mattheys/PlexCollectionTool /go/src/github.com/mattheys/PlexCollectionTool/

WORKDIR /go/src/github.com/mattheys/PlexCollectionTool/
RUN set -x && \
    go get github.com/golang/dep/cmd/dep && \
    dep ensure -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -i -v -ldflags="-X main.version=$(git describe --always --long --dirty)" -a -o app .
#go build -a -o app .

# Stage 2 (to create a downsized "container executable", ~7MB)

# If you need SSL certificates for HTTPS, replace `FROM SCRATCH` with:
#
#   FROM alpine:3.7
#   RUN apk --no-cache add ca-certificates
#
FROM scratch
WORKDIR /root/

COPY --from=builder /go/src/github.com/mattheys/gwlg.link/app .

ENTRYPOINT ["./app"]
