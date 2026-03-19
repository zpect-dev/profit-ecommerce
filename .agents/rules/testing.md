---
trigger: always_on
---

## Reglas de Testing (Go)

- **Aislamiento Estricto:** NUNCA hagas llamadas HTTP reales en los tests del módulo de Pedidos. Usa `httptest.Server` OBLIGATORIAMENTE para simular la API externa.
- **Mocks de Redis:** Para testear el módulo de Auth y el Carrito, usa la librería `alicebob/miniredis` o genera mocks de la interfaz del repositorio con `uber-go/mock` (anteriormente `golang/mock`). No requieras una instancia de Redis corriendo para los tests unitarios.
- **Table-Driven Tests:** Utiliza siempre el patrón de pruebas basadas en tablas de Go. Genera casos de prueba para: `Éxito`, `Error de BD`, `Error de Redis (Timeout)`, y `Datos Inválidos`.
- **Cobertura de Contexto:** Asegúrate de que los tests verifiquen que la cancelación del `context.Context` detiene la ejecución (especialmente en el Sincronizador).
