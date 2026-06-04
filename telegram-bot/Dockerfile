# Build context must be the repo root (the bot imports internal/ packages):
#   docker build -f telegram-bot/Dockerfile -t fcstask-telegram-bot .
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fcstask-telegram-bot ./telegram-bot/cmd/bot/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN adduser -D -s /sbin/nologin fcstask-bot
WORKDIR /home/fcstask-bot
COPY --from=builder /app/fcstask-telegram-bot ./
RUN chown fcstask-bot:fcstask-bot ./fcstask-telegram-bot && chmod +x ./fcstask-telegram-bot
USER fcstask-bot
CMD ["./fcstask-telegram-bot"]
