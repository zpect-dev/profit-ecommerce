---
trigger: always_on
---

## Mapa de Arquitectura del Proyecto (Para uso de Context7)

- `cmd/api/`: Punto de entrada de la API principal. Solo configuración y arranque del servidor HTTP/Router.
- `cmd/syncer/`: Punto de entrada del Sincronizador de BD. No debe contener lógica HTTP.
- `internal/auth/`: Dominio de Usuarios. Contiene handlers, services y repositorios que interactúan con Redis.
- `internal/catalog/`: Dominio de Productos. Solo lectura desde la BD.
- `internal/cart/`: Dominio del Carrito. Escritura en Redis (rápida) y persistencia diferida en BD.
- `internal/orders/`: Dominio de Pedidos. Cliente HTTP para la API externa.
- `pkg/`: Librerías compartidas (ej. inicialización de Redis, Logger, cliente HTTP base) que no contienen lógica de negocio.
