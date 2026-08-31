FROM golang:1.22-bookworm AS builder

WORKDIR /app

COPY go.mod ./
COPY . .
RUN go get go.mau.fi/whatsmeow@latest
RUN go get google.golang.org/protobuf@latest
RUN go get modernc.org/sqlite@latest
RUN go get github.com/google/uuid@latest
RUN go get github.com/skip2/go-qrcode@latest
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o agent .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/agent .

EXPOSE 8080

CMD ["./agent"]
