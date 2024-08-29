# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o homium .

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app/

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/homium .

EXPOSE 4577

CMD ["./homium"]
