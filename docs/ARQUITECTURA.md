# Arquitectura de FlowDesign POC

## Visión General

```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENTE / USUARIO                        │
└───────────────────────────────┬─────────────────────────────┘
                                │
                    HTTP/HTTPS  │
                                ▼
                    ┌───────────────────┐
                    │   Frontend        │
                    │   (Next.js)       │
                    │   :3000           │
                    └────────┬──────────┘
                             │
                  JSON/REST  │
                             ▼
        ┌────────────────────────────────────────┐
        │        Backend (Node.js/Express)       │
        │        :3001                           │
        │                                        │
        │  ┌──────────────┐  ┌──────────────┐    │
        │  │   Layout     │  │  Editions    │    │
        │  │   Module     │  │  Module      │    │
        │  └──────────────┘  └──────────────┘    │
        │                                        │
        │  ┌──────────────────────────────────┐  │
        │  │  Ventas Module (HTTP adapter)    │  │
        │  │  Consumes ESB via HTTP           │  │
        │  └──────────────┬───────────────────┘  │
        └─────────────────│──────────────────────┘
                          │
            HTTP JSON/REST│
                          ▼
        ┌──────────────────────────────────────┐
        │   ESB (Node.js/Express)              │
        │   :4000                              │
        │                                      │
        │  ┌──────────────┐  ┌──────────────┐  │
        │  │  Ventas      │  │  MySQL       │  │
        │  │  Module      │  │  Adapter     │  │
        │  └──────────────┘  └──────┬───────┘  │
        └───────────────────────────│──────────┘
                                    │
                          TCP 3306  │
                                    ▼
                    ┌────────────────────────────┐
                    │   MySQL Ventas             │
                    │   (on-premise)             │
                    │   192.168.28.33            │
                    │   vw_pauta_dataplan        │
                    └────────────────────────────┘


┌──────────────────────────────────────────────────┐
│  Base de datos del Backend (PostgreSQL :5432)    │
│                                                  │
│  ├── designflow_db                               │
│  │   ├── editions (ediciones)                    │
│  │   ├── pages (páginas)                         │
│  │   └── pautas (anuncios diagramados)           │
└──────────────────────────────────────────────────┘
```

## Componentes

### 1. Frontend (Next.js)

**Responsabilidad:** Interfaz de usuario para diagramadores

**Ubicación:** `/frontend`

**Cómo se comunica:**
```
Frontend ──HTTP──> Backend (:3001)
           GET /api/layout
           GET /api/editions
           POST /api/editions
           GET /api/ventas/ads
```

**Tecnologías:**
- Next.js 14+
- React
- TypeScript
- Tailwind CSS (posible)

Ver [FRONTEND.md](./FRONTEND.md) para detalles.

---

### 2. Backend (Node.js/Express)

**Responsabilidad:** API principal, lógica de negocio de diagramación

**Ubicación:** `/backend`

**Arquitectura:** Hexagonal (capas: domain, application, infrastructure)

**Módulos:**
- `layout` — Cálculos de grillas y layouts
- `editions` — Gestión de ediciones
- `pages` — Gestión de páginas
- `pautas` — Anuncios diagramados
- `ventas` — Integración con ESB (HTTP adapter)

**Base de datos:** PostgreSQL (:5432)

**Cómo se comunica:**
```
Backend ──HTTP──> ESB (:4000)
         GET /api/ventas/ads?date=YYYY-MM-DD

Backend ──SQL──> PostgreSQL
        SELECT/INSERT/UPDATE en designflow_db
```

Ver [BACKEND.md](./BACKEND.md) para detalles.

---

### 3. ESB (Enterprise Service Bus)

**Responsabilidad:** Integración con sistemas externos, transformación de datos

**Ubicación:** `/esb`

**Módulos:**
- `ventas` — Lectura desde MySQL de ventas
  - Puerto: `AdDataSourcePort` (interfaz abstracta)
  - Adaptador: `MysqlAdDataSourceAdapter` (conexión real)
  - Transformer: `AdImportTransformer` (conversión de datos)
  - Caso de uso: `GetImportedAdsUseCase`

**Base de datos:** MySQL on-premise (192.168.28.33:3306)

**Acceso:** Solo lectura a `vw_pauta_dataplan`

**Cómo se comunica:**
```
ESB ──SQL readonly──> MySQL Ventas
    SELECT FROM vw_pauta_dataplan
    WHERE publication_date = ?

ESB ──HTTP──> Backend (:3001)
    GET /api/ventas/ads → responde al backend
```

Ver [ESB.md](./ESB.md) para detalles.

---

### 4. Base de Datos Principal (PostgreSQL)

**Ubicación:** localhost:5432 en local, RDS en AWS

**Base de datos:** `designflow_db`

**Tablas principales:**
- `editions` — Ediciones de publicaciones
- `pages` — Páginas dentro de ediciones
- `pautas` — Anuncios ya diagramados

**Usuario:** `designflow` / `designflow123`

Ver [DATABASE.md](./DATABASE.md) para esquema completo.

---

### 5. Base de Datos de Ventas (MySQL)

**Ubicación:** 192.168.28.33:3306

**Base de datos:** `ventas`

**Vista:** `vw_pauta_dataplan` (datos de anuncios)

**Usuario:** `dataplan` (solo lectura)

**Acceso:** El ESB consulta esta BD, el Backend nunca

---

## Flujo de Datos Completo

### Caso: Obtener anuncios de una fecha

```
1. Usuario abre Frontend
   ↓
2. Frontend llama: GET http://localhost:3001/api/ventas/ads?date=2026-06-11
   ↓
3. Backend VentasController recibe la fecha
   ↓
4. Backend HttpEsbAdapter llama: GET http://esb:4000/api/ventas/ads?date=2026-06-11
   ↓
5. ESB VentasController recibe la fecha
   ↓
6. ESB GetImportedAdsUseCase valida formato
   ↓
7. ESB MysqlAdDataSourceAdapter ejecuta:
   SELECT ... FROM vw_pauta_dataplan WHERE publication_date = '2026-06-11'
   ↓
8. MySQL retorna filas crudas
   ↓
9. ESB AdImportTransformer transforma:
   - Convierte pulgadas → milímetros
   - Asigna grilla 6x8
   - Estructura en dominio AdImport
   ↓
10. ESB retorna JSON al Backend
    ↓
11. Backend retorna al Frontend
    ↓
12. Frontend muestra anuncios en UI
```

---

## Decisiones Arquitectónicas

### ¿Por qué Hexagonal?

El backend usa arquitectura hexagonal porque:
- Desacoplamiento de la persistencia (cambiar BD es fácil)
- Casos de uso claros y testables
- Fácil de extender sin romper lo existente

### ¿Por qué ESB separado?

El ESB es un servicio independiente porque:
- Desacoplamiento del backend
- Las credenciales de ventas no llegan al backend
- Escalable por separado
- Podría consumirlo otro sistema en el futuro

### ¿Por qué HttpEsbAdapter en backend?

En lugar de hablar directamente a MySQL desde el backend:
- Backend no conoce credenciales de ventas
- Cambios en el ESB no afectan el backend
- En AWS: ESB es Lambda, backend es ECS, se comunican por HTTP

### ¿Por qué Docker Compose?

Para simular la VPC de AWS localmente:
- Los servicios se descubren por nombre (`esb`, `postgres`)
- Igual que en AWS ECS/RDS
- Fácil cambiar de local a AWS sin tocar código

---

## Escalado a AWS

Cuando el POC esté listo para AWS:

| Componente Local | Componente AWS |
|---|---|
| Frontend (npm dev) | CloudFront + S3 (estático) o App Runner |
| Backend (:3001) | ECS Task (o App Runner) |
| ESB (:4000) | Lambda + API Gateway |
| PostgreSQL (:5432) | RDS PostgreSQL |
| Docker Compose | Terraform / CDK |
| `.env` | AWS Secrets Manager |

**El código no cambia, solo la infraestructura.**

---

## Seguridad

- **Base de datos de ventas:** Usuario solo lectura
- **Secretos:** `.env` en local, Secrets Manager en AWS
- **Red:** VPC privada en AWS, Site-to-Site VPN a on-premise
- **CORS:** Configurado por ambiente
- **Imágenes Docker:** Multi-stage, sin credenciales

---

## Performance

- **Queries a MySQL:** Optimizadas con índices en `vw_pauta_dataplan`
- **Cache:** PostgreSQL en memoria (16GB en local)
- **CDN:** CloudFront (futuro, para frontend)
- **Serverless ESB:** Lambda con provisioned concurrency si es necesario

---

## Monitoreo (futuro)

- CloudWatch para logs
- X-Ray para tracing distribuido
- CloudFormation/Terraform para IaC

---

**Última actualización:** 2026-06-30
