# Backend — API Principal

## Visión General

El backend es la API central que orquesta la diagramación de publicaciones. Usa arquitectura hexagonal para separar la lógica de negocio de los detalles técnicos.

**Stack:** Node.js 22 + Express + TypeScript + PostgreSQL

**Puerto:** 3001 (local), configurable

## Endpoints

### Health Check

```
GET /health

Respuesta:
{
  "status": "ok",
  "service": "designflow-backend",
  "database": "connected"
}
```

### Layout Module

```
GET /api/layout
POST /api/layout
  - body: { editionId, pages, pautas }
```

Calcula la grilla y distribuye anuncios en la página.

### Editions Module

```
GET /api/editions
POST /api/editions
  - body: { no_paginas, ancho_mm, alto_mm, cuadros_ancho, cuadros_alto }
```

Gestiona ediciones (publicaciones).

### Ventas Module (ESB Integration)

```
GET /api/ventas/ads?date=YYYY-MM-DD

Respuesta:
{
  "date": "2026-06-11",
  "total": 45,
  "ads": [
    {
      "fid": "3681453",
      "customer": "Diarios Modernos, S.a.",
      "product": "Centro Occidente",
      "slogan": "NUESTRO TRABAJO",
      "coverDate": "2026-06-11",
      "formatFid": "6.00x8.00-MN",
      "cuadrosAncho": 6,
      "cuadrosAlto": 8,
      "widthMm": 254,
      "heightMm": 317.5,
      "colorFid": "Full",
      "placementComment": "Desplegado",
      "statusFid": "Por emitir",
      "printSystemFid": "32542600042",
      "regionalName": "Region Centro Occidente",
      "productCategory": "Departamental",
      "observations": ""
    }
  ]
}
```

Obtiene anuncios del ESB para una fecha específica.

## Arquitectura Hexagonal

```
src/
├── bootstrap/              # Entry point
│   ├── app.ts             # Express app
│   ├── server.ts          # Server startup
│   └── dependency-container.ts  # DI container
│
├── config/
│   ├── environment.config.ts    # Env variables
│   └── database.config.ts       # Database connection
│
├── modules/                # Lógica de negocio
│   ├── layout/
│   │   ├── domain/              # Reglas de negocio puras
│   │   │   ├── entities/
│   │   │   └── services/
│   │   ├── application/         # Casos de uso
│   │   │   ├── contracts/
│   │   │   ├── dto/
│   │   │   └── use-cases/
│   │   └── infraestructure/     # Detalles técnicos
│   │       └── http/
│   │
│   ├── editions/
│   ├── pages/
│   ├── pautas/
│   └── ventas/
│       └── infraestructure/http/
│           ├── HttpEsbAdapter.ts     # Llama al ESB
│           ├── VentasController.ts
│           └── VentasRoutes.ts
│
└── shared/                 # Utilidades compartidas
    ├── application/
    ├── domain/
    │   └── DomainError.ts
    └── infraestructure/
        └── http/
            ├── error-handler.middleware.ts
            └── not-found.middleware.ts
```

## Módulos

### Layout

Calcula posiciones de anuncios en la página basado en una grilla.

**Servicios:**
- `GridLayoutCalculator` — Cálculos matemáticos de grilla
- `LayoutValidator` — Validaciones de layout

**Casos de uso:**
- `GenerateEditionLayoutUseCase` — Calcula layout completo
- `GetAllEditionsLayoutUseCase` — Obtiene layouts de todas las ediciones

### Editions

Gestión de ediciones (publicaciones).

**Entidades de dominio:**
- `Edition` — Una edición con dimensiones, páginas, márgenes

**Casos de uso:**
- `CreateEditionUseCase` — Crear edición

**Persistencia:**
- `PostgresEditionRepository`

### Pages

Gestión de páginas dentro de ediciones.

**Entidades de dominio:**
- `Page` — Una página con número y referencia a edición

**Persistencia:**
- `PostgresPageRepository`

### Pautas

Anuncios ya diagramados en una página.

**Entidades de dominio:**
- `Pauta` — Un anuncio con posición (x, y) y tamaño

**Persistencia:**
- `PostgresPautaRepository`

### Ventas (ESB Integration)

**No tiene lógica de negocio local.** Solo es un adaptador HTTP que consume el ESB.

**Adaptador:**
- `HttpEsbAdapter` — Llama a `http://esb:4000/api/ventas/ads`

**Controlador:**
- `VentasController` — Recibe request, delega al adapter, responde

**DTO:**
- `AdImportDto` — Data Transfer Object para respuestas

## Flujo de una Request

### Ejemplo: GET /api/ventas/ads?date=2026-06-11

```
1. Express Route Handler
   ↓
2. VentasController.getAdsByDate()
   - Valida query params
   ↓
3. HttpEsbAdapter.getAdsByDate(date)
   - Construye URL: http://esb:4000/api/ventas/ads?date=2026-06-11
   - Hace fetch()
   ↓
4. ESB responde con array de anuncios
   ↓
5. VentasController retorna JSON al cliente
```

## Configuración

### Variables de Entorno

```bash
# Puerto
PORT=3001

# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=designflow
DB_PASSWORD=designflow123
DB_NAME=designflow_db
DB_SYNCHRONIZE=false
DB_LOGGING=true

# ESB
ESB_URL=http://localhost:4000

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_CREDENTIALS=true
```

### En Docker Compose

```yaml
backend:
  environment:
    ESB_URL: http://esb:4000    # Nombre del servicio
    DB_HOST: postgres            # Nombre del servicio
```

## Desarrollo

### Levantar en local

```bash
cd backend
npm install
npm run dev
```

El servidor corre en `http://localhost:3001` con hot reload.

### Estructura de un módulo nuevo

Si quieres agregar un nuevo módulo (ej: "productos"):

```
src/modules/productos/
├── domain/
│   ├── entities/
│   │   └── Producto.ts
│   └── repositories/
│       └── ProductoRepository.ts
├── application/
│   └── use-cases/
│       └── GetProductosUseCase.ts
└── infraestructure/
    ├── persistence/
    │   ├── PostgresProductoRepository.ts
    │   └── entities/
    │       └── ProductoEntity.ts
    └── http/
        ├── ProductoController.ts
        └── ProductoRoutes.ts
```

1. Define la entidad en `domain/entities/Producto.ts`
2. Define el repositorio en `domain/repositories/ProductoRepository.ts`
3. Implementa el repositorio en `infraestructure/persistence/PostgresProductoRepository.ts`
4. Crea el caso de uso en `application/use-cases/GetProductosUseCase.ts`
5. Crea el controlador en `infraestructure/http/ProductoController.ts`
6. Registra en `dependency-container.ts`

## Testing

Por ahora no hay tests (será implementado después). Ver [README.md](../README.md) para roadmap.

## Troubleshooting

### Error: "Cannot connect to database"
- Verificar que PostgreSQL está corriendo: `docker compose up postgres`
- Verificar credenciales en `.env`
- Verificar que `DB_HOST=postgres` en Docker Compose

### Error: "ESB connection timeout"
- Verificar que ESB está corriendo: `docker compose up esb` o `cd esb && npm run dev`
- Verificar que `ESB_URL` es correcto
- En local: `http://localhost:4000`, en Docker: `http://esb:4000`

### Error: "Query error in MysqlAdDataSourceAdapter"
- El error es del ESB, no del backend
- Ver [ESB.md](./ESB.md) para troubleshoot

---

**Última actualización:** 2026-06-30
