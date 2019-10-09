###############################################################################
#### Run the build on alpine - istiod doesn't need more.
# Main docker images for istiod will be distroless and alpine.
FROM golang:1.13-alpine AS build-base

WORKDIR /ws
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOPROXY=https://proxy.golang.org

RUN apk add --no-cache git

# With caching should avoid repeated downloads as long as the sum/mod don't change
COPY go.mod go.sum  ./
RUN go mod download


