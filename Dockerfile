FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fcstask-api ./internal/cmd/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN adduser -D -s /sbin/nologin fcstask-admin
WORKDIR /home/fcstask-admin
COPY --from=builder /app/fcstask-api ./
COPY --from=builder /app/config/config.yaml ./config/
RUN chown fcstask-admin:fcstask-admin ./fcstask-api && chmod +x ./fcstask-api
RUN chown fcstask-admin:fcstask-admin ./config/config.yaml
USER fcstask-admin
EXPOSE 8080
CMD ["./fcstask-api"]