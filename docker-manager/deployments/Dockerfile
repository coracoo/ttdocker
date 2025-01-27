# deployments/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/ .
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk add --no-cache docker-cli
COPY --from=builder /app/main /app/main
VOLUME /var/run/docker.sock
CMD ["/app/main"]