# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /haru ./cmd/server

# Stage 2: Runtime
FROM alpine:3.19

RUN adduser -D -u 1000 appuser
COPY --from=builder /haru /haru

USER appuser
EXPOSE 8080
CMD ["/haru"]
