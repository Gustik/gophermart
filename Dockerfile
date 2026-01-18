# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN cd cmd/gophermart && go build -o gophermart -v

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/cmd/gophermart/gophermart .

# Copy migrations
COPY migrations ./migrations

# Copy accrual binary
COPY cmd/accrual/accrual_linux_amd64 ./accrual_linux_amd64
RUN chmod +x ./accrual_linux_amd64

# Expose port
EXPOSE 8080

# Run
CMD ["./gophermart"]