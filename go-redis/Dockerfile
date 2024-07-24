FROM golang:1.22.2-alpine AS builder

WORKDIR /app

COPY . .
COPY .env .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main main.go



FROM alpine:latest

EXPOSE 3729
WORKDIR /app

COPY --from=builder /app/.env /app/.env

COPY --from=builder /app/main /app/main

CMD ["./main"]
