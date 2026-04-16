FROM golang:1.20-alpine AS builder
WORKDIR /app
# Copy semua file ke dalam Docker
COPY . .
# Paksa Docker untuk download dan perbaiki dependencies yang kurang otomatis
RUN go env -w GOPROXY=direct
RUN go mod tidy
# Rakit aplikasinya
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Setup container untuk production
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
