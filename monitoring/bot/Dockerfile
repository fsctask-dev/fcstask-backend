FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/main.go

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /app/bot .

RUN apk --no-cache add ca-certificates

CMD ["./bot"]
