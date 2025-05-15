FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o receipt-processor .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/receipt-processor .

EXPOSE 8080

CMD ["./receipt-processor"]