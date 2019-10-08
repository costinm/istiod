###############################################################################
#### Run the build on alpine - istiod doesn't need more.
# Main docker images for istiod will be distroless and alpine.
FROM golang:1.13-alpine AS build-base

ENV GO111MODULE=on
RUN apk add --no-cache git && pwd && ls && mkdir /ws
WORKDIR /ws
# With caching should avoid repeated downloads as long as the sum/mod don't change
COPY go.mod go.sum  ./
RUN GOPROXY=https://proxy.golang.org go mod download


