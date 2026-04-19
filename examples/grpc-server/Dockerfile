# Build stage
FROM golang:1.26.2 as builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the gRPC server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpc-server ./examples/grpc-server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/grpc-server .

# Expose gRPC port
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD test -f grpc-server || exit 1

# Run the gRPC server
ENV OATHMESH_AUDIENCE="grpc-example-service"
ENV OATHMESH_TRUSTED_ISSUERS="https://issuer.example.com"
ENV OATHMESH_PORT="50051"

CMD ["./grpc-server"]
