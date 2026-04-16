FROM golang:1.22-alpine AS builder

# Install git agar go mod bisa download pustaka
RUN apk add --no-cache git

WORKDIR /app
COPY . .

# Paksa bikin go.mod baru dan rakit
RUN go mod init wastewatch || true
RUN go env -w GOPROXY=direct
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Setup container untuk production
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
