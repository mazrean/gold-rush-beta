# syntax = docker/dockerfile:1.0-experimental

FROM golang:1.16.0-alpine as build

RUN apk add --update --no-cache git

WORKDIR /github.com/traPtitech/anke-to

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -o /main -ldflags "-s -w"

FROM alpine:3.13.2
WORKDIR /app
RUN apk --update --no-cache add ca-certificates \
  && update-ca-certificates \
  && rm -rf /usr/share/ca-certificates

COPY --from=build /main ./
ENTRYPOINT ./main