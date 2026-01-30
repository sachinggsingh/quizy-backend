FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/main

EXPOSE 8080

CMD ["./main"]