FROM golang:alpine AS builder
WORKDIR /workspace
COPY . .
RUN go build -o headscale-oidc-sync

FROM alpine:latest
RUN apk add --no-cache docker-cli
WORKDIR /workspace
COPY --from=builder /workspace/headscale-oidc-sync .
VOLUME /var/run/docker.sock
CMD ["./headscale-oidc-sync"]