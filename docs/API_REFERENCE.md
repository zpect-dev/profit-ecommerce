# API Reference — Profit E-Commerce Backend

> **Base URL:** `http://localhost:8050/v1`
>
> Todos los endpoints (excepto Login) requieren autenticación mediante **Bearer Token** en el header `Authorization`.

---

## Tabla de Contenidos

| Módulo   | Endpoint                  | Método | Auth            |
| -------- | ------------------------- | ------ | --------------- |
| Auth     | `/v1/auth/login`          | POST   | ❌ Pública      |
| Auth     | `/v1/auth/logout`         | POST   | ✅ Bearer Token |
| Catálogo | `/v1/products`            | GET    | ✅ Bearer Token |
| Catálogo | `/v1/products/{id}`       | GET    | ✅ Bearer Token |
| Catálogo | `/v1/products/categories` | GET    | ✅ Bearer Token |
| Catálogo | `/v1/products/batch`      | POST   | ✅ Bearer Token |
| Carrito  | `/v1/cart`                | GET    | ✅ Bearer Token |
| Carrito  | `/v1/cart`                | POST   | ✅ Bearer Token |

---

## Módulo de Autenticación

### `POST /v1/auth/login`

Autentica al cliente contra la base de datos y devuelve un token de sesión junto con el perfil completo. Este es el **único endpoint público** de la API.

> **Nota:** En la versión actual, `username` y `password` deben ser idénticos al código del cliente (`co_cli`) registrado en el ERP Profit.

**Request:**

```json
{
    "username": "CLI001",
    "password": "CLI001"
}
```

**Response (200 OK):**

```json
{
    "token": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "CLI001",
    "cli_des": "Distribuidora Nacional SA",
    "tipo": "A",
    "mont_cre": 15000.5
}
```

| Campo      | Tipo   | Descripción                                                         |
| ---------- | ------ | ------------------------------------------------------------------- |
| `token`    | string | UUID v4. Usar como Bearer Token en todas las peticiones protegidas. |
| `user_id`  | string | Código del cliente (`co_cli`) en el ERP.                            |
| `cli_des`  | string | Nombre/razón social del cliente.                                    |
| `tipo`     | string | Clasificación del cliente (determina lista de precios).             |
| `mont_cre` | number | Saldo/crédito disponible del cliente.                               |

**Errores:**

| Status | Causa                                                              |
| ------ | ------------------------------------------------------------------ |
| 400    | Body vacío o JSON malformado.                                      |
| 401    | Credenciales inválidas, cliente no encontrado, o cliente inactivo. |

**Ejemplo con cURL:**

```bash
curl -X POST http://localhost:8050/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"CLI001","password":"CLI001"}'
```

---

### `POST /v1/auth/logout`

Invalida la sesión actual eliminando el token de Redis.

**Headers requeridos:**

```
Authorization: Bearer <token>
```

**Response:** `204 No Content`

**Errores:**

| Status | Causa                                |
| ------ | ------------------------------------ |
| 401    | Token faltante, inválido o expirado. |

**Ejemplo con cURL:**

```bash
curl -X POST http://localhost:8050/v1/auth/logout \
  -H "Authorization: Bearer 550e8400-e29b-41d4-a716-446655440000"
```

---

## Módulo de Catálogo (Protegido)

> Todos los endpoints de catálogo requieren `Authorization: Bearer <token>`.

### `GET /v1/products`

Devuelve la lista completa de productos sincronizados desde el ERP Profit.

> **Roadmap:** Este endpoint soportará filtrado por categoría (`?category=CAT01`), búsqueda por texto (`?search=jeringa`) y paginación (`?page=1&limit=20`) en futuras iteraciones.

**Headers requeridos:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
[
    {
        "co_art": "ART001",
        "art_des": "Jeringa Descartable 10ml",
        "stock_act": 500,
        "prec_vta1": 2.5,
        "prec_vta2": 2.3,
        "prec_vta3": 2.1,
        "prec_vta4": 1.9,
        "prec_vta5": 1.75,
        "tipo_imp": "G",
        "co_lin": "LIN01",
        "co_cat": "CAT01",
        "co_subl": "SUB01",
        "image_url": "https://imagenes.cristmedicals.com/imagenes-v3/imagenes/ART001.jpg",
        "inventory_json": {
            "ALM01": {
                "nombre": "Almacén Principal",
                "stock_total": 300,
                "stock_comprometido": 50,
                "stock_por_llegar": 100
            }
        }
    }
]
```

---

### `GET /v1/products/{id}`

Devuelve un producto específico por su código de artículo.

**Ejemplo:**

```bash
curl http://localhost:8050/v1/products/ART001 \
  -H "Authorization: Bearer <token>"
```

---

### `GET /v1/products/categories`

Devuelve las categorías, líneas y sub-líneas disponibles para filtrado.

---

### `POST /v1/products/batch`

Obtiene múltiples productos por sus IDs en una sola petición (útil para el carrito).

**Request:**

```json
{
    "ids": ["ART001", "ART002", "ART003"]
}
```

---

## Módulo de Carrito (Protegido)

> El carrito está vinculado al `user_id` extraído automáticamente del Bearer Token por el middleware. **No es necesario enviar el ID del usuario en el body ni en headers.**

### `GET /v1/cart`

Devuelve el carrito del usuario autenticado con **validación de stock en tiempo real**:

- Si un producto tiene stock insuficiente, la cantidad se ajusta automáticamente.
- Si un producto tiene stock 0, se elimina del carrito.

**Headers requeridos:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
    "user_id": "CLI001",
    "items": [
        {
            "product_id": "ART001",
            "quantity": 5
        },
        {
            "product_id": "ART002",
            "quantity": 2
        }
    ]
}
```

> **Nota:** Las cantidades devueltas pueden diferir de las solicitadas originalmente si el stock cambió. El Frontend debe usar esta respuesta como fuente de verdad.

---

### `POST /v1/cart`

Agrega un producto al carrito del usuario autenticado. Si el producto ya existe, las cantidades se suman.

**Headers requeridos:**

```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request:**

```json
{
    "product_id": "ART001",
    "quantity": 3
}
```

**Response:** `201 Created`

**Errores:**

| Status | Causa                                  |
| ------ | -------------------------------------- |
| 400    | JSON malformado o campos faltantes.    |
| 401    | Token faltante, inválido o expirado.   |
| 500    | Error interno del servidor (Redis/BD). |

**Ejemplo con cURL:**

```bash
curl -X POST http://localhost:8050/v1/cart \
  -H "Authorization: Bearer 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{"product_id":"ART001","quantity":3}'
```

---

## Autenticación — Guía para Frontend (React)

### Flujo Recomendado

1. **Login:** El usuario ingresa su código de cliente. Llamar a `POST /v1/auth/login`.
2. **Guardar en estado global:** Almacenar `token`, `cli_des`, `tipo` y `mont_cre` en Redux/Context.
3. **Interceptor Axios:** Configurar un interceptor que añada `Authorization: Bearer <token>` a todas las peticiones.
4. **Manejo de 401:** Si cualquier endpoint retorna `401`, redirigir al login.

### Ejemplo de Interceptor (Axios)

```javascript
import axios from "axios";

const api = axios.create({
    baseURL: "http://localhost:8050/v1",
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem("token");
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem("token");
            window.location.href = "/login";
        }
        return Promise.reject(error);
    },
);

export default api;
```

---

## Configuración Docker

```bash
# Levantar todos los servicios
docker compose up --build

# Detener y limpiar volúmenes
docker compose down -v
```

| Servicio   | Puerto | Descripción                     |
| ---------- | ------ | ------------------------------- |
| API        | 8050   | Backend principal               |
| PostgreSQL | 5432   | Base de datos destino           |
| Redis      | 6379   | Cache de sesiones y carrito     |
| Adminer    | 8081   | UI para inspeccionar PostgreSQL |
