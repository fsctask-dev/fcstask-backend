FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY ./ ./

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

RUN oapi-codegen -generate types,skip-prune \
    -package api \
    -o internal/api/types.gen.go \
    api/openapi.yaml

RUN oapi-codegen -generate server \
    -package api \
    -o internal/api/server.gen.go \
    api/openapi.yaml

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fcstask-api ./internal/cmd/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fcstask-migrate ./migrate/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN adduser -D -s /sbin/nologin fcstask-admin
WORKDIR /home/fcstask-admin
COPY --from=builder /app/fcstask-api ./
COPY --from=builder /app/fcstask-migrate ./
COPY --from=builder /app/config/config.yaml ./config/
COPY --from=builder /app/internal/db/migration ./internal/db/migration
RUN chown fcstask-admin:fcstask-admin ./fcstask-api && chmod +x ./fcstask-api
RUN chown fcstask-admin:fcstask-admin ./config/config.yaml
USER fcstask-admin
EXPOSE 8080
CMD ["./fcstask-api"]
