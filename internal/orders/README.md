# Orders Module

Dominio de Pedidos.

## Responsabilidades
- Creación de pedidos contra la API externa de Profit
- Reutilizar un único `http.Client` global con timeouts

## Capas (por implementar)
- `handler.go` — HTTP handlers
- `service.go` — Lógica de negocio
- `client.go` — Cliente HTTP para la API externa
- `models.go` — Order, OrderItem
- `ports.go` — Interfaces
