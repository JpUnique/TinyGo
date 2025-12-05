# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o tinygo-api .

# Runtime stage
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /root/

COPY --from=builder --chown=nonroot:nonroot /app/tinygo-api /app/
EXPOSE 8080

ENV PORT=8080
CMD ["app/tinygo-api"]