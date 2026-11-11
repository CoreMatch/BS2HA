# syntax=docker/dockerfile:1
ARG GO_VERSION=1.26.0

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bs2ha ./cmd/bs2ha

# bs2ha is a one-shot migration CLI: it reads the Blessing-Skin-Server
# database in the working directory and writes the HRPAuth baseline files.
FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
 && mkdir -p /app \
 && adduser -D -H -u 10001 -h /app app \
 && chown -R app:app /app
WORKDIR /app
COPY --from=builder /out/bs2ha /app/bs2ha
USER app
ENTRYPOINT ["/app/bs2ha"]
