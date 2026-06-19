ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY backup-service ./backup-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/backup-service ./backup-service/cmd/backup-service

ARG PG_MAJOR=17
FROM postgres:${PG_MAJOR}-alpine

COPY --from=builder /out/backup-service /usr/local/bin/backup-service
COPY backup-service/config.yaml /etc/backup-service/config.yaml

RUN mkdir -p /var/backups/postgres /var/backups/postgres-wal/base \
        /var/backups/postgres-wal/archive /var/lib/backup-service /var/log \
 && chown -R postgres:postgres /var/backups /var/lib/backup-service /var/log /etc/backup-service

USER postgres
WORKDIR /var/lib/backup-service

ENTRYPOINT ["backup-service"]
CMD ["serve", "--config", "/etc/backup-service/config.yaml"]
