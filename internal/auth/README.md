# Auth Module

Dominio de Usuarios y Autenticación.

## Responsabilidades
- Login / Registro
- Sesiones en Redis
- Middleware de autenticación JWT

## Capas (por implementar)
- `handler.go` — HTTP handlers
- `service.go` — Lógica de negocio
- `repository.go` — Acceso a Redis para sesiones
- `models.go` — User, Session
- `ports.go` — Interfaces
