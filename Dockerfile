FROM golang:1.22-bookworm AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o agent .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/agent .

EXPOSE 8080

CMD ["./agent"]
