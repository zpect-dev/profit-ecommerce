---
trigger: always_on
---

## Reglas de Base de Datos y Sincronizador

- **Dueño del Esquema:** El catálogo principal de la base de datos es gestionado por [herramienta de migración, ej. golang-migrate]. El Sincronizador SOLO tiene permisos de escritura/actualización (UPSERT) sobre las tablas del catálogo, pero NUNCA debe alterar la estructura de las tablas.
- **Bloqueos (Locks):** Al sincronizar datos pesados, el worker debe implementar bloqueos optimistas (Optimistic Locking) o usar transacciones explícitas de BD para no afectar las lecturas concurrentes de la API del Catálogo.
