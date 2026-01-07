# Stage 1: Builder
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copiar go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Construir binario estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o orgmdns \
    ./cmd/orgmdns

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copiar binario desde builder
COPY --from=builder /build/orgmdns /app/orgmdns

# Nota: El directorio logs/ se creará automáticamente por la aplicación
# o se montará como volumen desde docker-compose

# Usuario no root (distroless ya viene con usuario no root)
USER nonroot:nonroot

# Entrypoint
ENTRYPOINT ["/app/orgmdns"]
