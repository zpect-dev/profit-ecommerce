# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema necesarias para cgo (si fuera necesario, aunque aquí usaremos CGO_ENABLED=0 pero git a veces ayuda)
RUN apk add --no-cache git

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar API
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# Final Stage
FROM alpine:3.19

WORKDIR /app

# Instalar certificados CA para https (importante para conectar a AWS/Azure o APIs externas si fuera el caso)
RUN apk --no-cache add ca-certificates tzdata

# Copiar binarios del builder
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/db ./db
# Exponer puerto de la API (informativo)
EXPOSE 8050

# Usuario no root por seguridad
RUN adduser -D nonroot
USER nonroot

# Por defecto corremos la API, pero se puede sobreescribir en docker-compose
CMD ["/app/api"]
