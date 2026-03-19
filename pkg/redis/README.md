# Redis Package

Librería compartida para inicialización y configuración del cliente Redis.

## Responsabilidades
- Inicialización del cliente `go-redis`
- Pool de conexiones
- Health checks

## Uso
Será utilizado por `internal/auth/` y `internal/cart/` para acceso a Redis.
No debe contener lógica de negocio.
