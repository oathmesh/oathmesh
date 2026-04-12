FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy the entire module 
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the chi-api statically
RUN CGO_ENABLED=0 GOOS=linux go build -o /chi-api ./examples/chi-api

# Create minimal runtime container
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /chi-api .

EXPOSE 8080
CMD ["./chi-api"]
