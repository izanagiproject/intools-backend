# backend/Dockerfile

# Build Stage
FROM golang:1.20.10-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Final Stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/app /app/app

CMD ["./app"]
