FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/server .

COPY configs/ ./configs/
COPY .env .

RUN apk add --no-cache ca-certificates

EXPOSE 7878

CMD ["./server"]
