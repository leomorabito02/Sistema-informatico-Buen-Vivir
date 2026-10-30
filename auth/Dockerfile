FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache upgrade && apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server .

RUN chown appuser:appgroup /app/server && \
    chmod 500 /app/server

USER appuser

EXPOSE 8080 50051

CMD ["./server"]
