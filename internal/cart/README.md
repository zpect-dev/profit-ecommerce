# Cart Module

Dominio del Carrito de Compras.

## Responsabilidades
- CRUD del carrito (escritura rápida en Redis)
- Persistencia diferida en BD
- Uso OBLIGATORIO de Redis `TxPipeline` para escrituras múltiples

## Capas (por implementar)
- `handler.go` — HTTP handlers
- `service.go` — Lógica de negocio
- `repository.go` — Redis + BD
- `models.go` — Cart, CartItem
- `ports.go` — Interfaces
