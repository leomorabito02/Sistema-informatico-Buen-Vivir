FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server/main.go

FROM alpine:3.21

# apk upgrade trae los parches de seguridad más recientes de los paquetes
# base (ej. openssl) en cada build — el mismo criterio que auth/Dockerfile,
# necesario para que el scan de Trivy del pipeline no bloquee por CVEs de
# la capa OS ya corregidas upstream.
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server .

RUN chown appuser:appgroup /app/server && \
    chmod 500 /app/server

USER appuser

EXPOSE 8210

CMD ["./server"]
