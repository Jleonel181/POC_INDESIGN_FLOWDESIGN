# Database — PostgreSQL

## Visión General

La base de datos principal de FlowDesign es PostgreSQL. Almacena la estructura de ediciones, páginas y pautas (anuncios diagramados).

**Base de datos:** `designflow_db`

**Host:** localhost:5432 (local), RDS en AWS

**Usuario:** `designflow` / `designflow123`

## Schema

### Tabla: editions

Ediciones (publicaciones).

```sql
CREATE TABLE editions (
  id SERIAL PRIMARY KEY,
  no_paginas INTEGER NOT NULL,
  ancho_mm DECIMAL(10, 2) NOT NULL,
  alto_mm DECIMAL(10, 2) NOT NULL,
  cuadros_ancho INTEGER NOT NULL,
  cuadros_alto INTEGER NOT NULL,
  facing_pages BOOLEAN DEFAULT FALSE,
  margen_superior_mm DECIMAL(10, 2) DEFAULT 0,
  margen_inferior_mm DECIMAL(10, 2) DEFAULT 0,
  margen_izquierdo_mm DECIMAL(10, 2) DEFAULT 0,
  margen_derecho_mm DECIMAL(10, 2) DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Campos:**
- `id` — ID único
- `no_paginas` — Número de páginas en la edición
- `ancho_mm`, `alto_mm` — Dimensiones en milímetros
- `cuadros_ancho`, `cuadros_alto` — Tamaño de grilla (ej: 6×8)
- `facing_pages` — Si es "vistazo doble"
- `margen_*` — Márgenes en 4 lados
- `created_at`, `updated_at` — Timestamps

**Ejemplo:**
```sql
INSERT INTO editions 
  (no_paginas, ancho_mm, alto_mm, cuadros_ancho, cuadros_alto) 
VALUES 
  (24, 216, 279, 6, 8);
```

---

### Tabla: pages

Páginas dentro de ediciones.

```sql
CREATE TABLE pages (
  id SERIAL PRIMARY KEY,
  no_pagina INTEGER NOT NULL,
  edicion_id INTEGER NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pages_edicion_id ON pages(edicion_id);
```

**Campos:**
- `id` — ID único
- `no_pagina` — Número de página dentro de la edición (1, 2, 3...)
- `edicion_id` — FK a `editions`

**Relación:**
```
Edition (1) ──────→ (N) Pages
  1 edición = 24 páginas
```

---

### Tabla: pautas

Anuncios ya diagramados.

```sql
CREATE TABLE pautas (
  id SERIAL PRIMARY KEY,
  descripcion_pauta VARCHAR(255) NOT NULL,
  cuadros_alto INTEGER NOT NULL,
  cuadros_ancho INTEGER NOT NULL,
  ubicacion_cuadros_x INTEGER NOT NULL,
  ubicacion_cuadros_y INTEGER NOT NULL,
  pagina_id INTEGER NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pautas_pagina_id ON pautas(pagina_id);
```

**Campos:**
- `id` — ID único
- `descripcion_pauta` — Descripción del anuncio (ej: "NUESTRO TRABAJO")
- `cuadros_alto`, `cuadros_ancho` — Tamaño en cuadros de grilla
- `ubicacion_cuadros_x`, `ubicacion_cuadros_y` — Posición en grilla
- `pagina_id` — FK a `pages`

**Relación:**
```
Page (1) ──────→ (N) Pautas
  1 página = múltiples anuncios
```

**Ejemplo:**
```sql
INSERT INTO pautas 
  (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) 
VALUES 
  ('NUESTRO TRABAJO', 8, 6, 0, 0, 5);
```

---

## Relaciones Completas

```
editions (1)
    │
    └──→ (N) pages
            │
            └──→ (N) pautas
```

**Ejemplo de estructura completa:**

Edición 1 (24 páginas, 216×279mm, grilla 6×8)
├── Página 1
│   ├── Pauta 1 (NUESTRO TRABAJO, 6×8, pos 0,0)
│   ├── Pauta 2 (CAMPAÑA EDICTOS, 3×4, pos 0,0)
│   └── Pauta 3 (FESTIVAL DE DESCUENTO, 6×4, pos 0,4)
├── Página 2
│   ├── Pauta 4 (...)
│   └── Pauta 5 (...)
...
└── Página 24
    └── Pauta N (...)
```

## Migraciones

Las migraciones se encuentran en `backend/database/init.sql`.

Se ejecutan automáticamente al levantar PostgreSQL con Docker Compose:

```yaml
postgres:
  volumes:
    - ./backend/database/init.sql:/docker-entrypoint-initdb.d/init.sql
```

### Crear Tabla Nueva

Agregar a `backend/database/init.sql`:

```sql
CREATE TABLE nueva_tabla (
  id SERIAL PRIMARY KEY,
  ...
);
```

---

## Queries Útiles

### Obtener todas las ediciones

```sql
SELECT * FROM editions;
```

### Obtener páginas de una edición

```sql
SELECT p.* 
FROM pages p
WHERE p.edicion_id = 1
ORDER BY p.no_pagina;
```

### Obtener pautas de una página

```sql
SELECT p.* 
FROM pautas p
WHERE p.pagina_id = 5
ORDER BY p.id;
```

### Contar pautas por edición

```sql
SELECT 
  e.id, 
  e.no_paginas, 
  COUNT(p.id) as total_pautas
FROM editions e
LEFT JOIN pages pg ON e.id = pg.edicion_id
LEFT JOIN pautas p ON pg.id = p.pagina_id
GROUP BY e.id;
```

### Verificar disponibilidad de grilla

```sql
-- Cuántos cuadros (de 6×8) están usados en página 5
SELECT 
  SUM(cuadros_ancho * cuadros_alto) as cuadros_usados,
  (6 * 8) as cuadros_disponibles
FROM pautas
WHERE pagina_id = 5;
```

---

## Conexión en Local

### Con psql

```bash
psql -h localhost -U designflow -d designflow_db
# Password: designflow123
```

### Con Docker Compose

```bash
docker compose exec postgres psql -U designflow -d designflow_db
```

### Con cliente GUI

- **PgAdmin** — http://localhost:5050 (si lo agregas a docker-compose)
- **DBeaver** — Desktop app
- **VS Code** — Extensión SQLTools

---

## Respaldos

### Hacer backup

```bash
# Local
pg_dump -h localhost -U designflow designflow_db > backup.sql

# Docker
docker compose exec postgres pg_dump -U designflow designflow_db > backup.sql
```

### Restaurar backup

```bash
# Local
psql -h localhost -U designflow designflow_db < backup.sql

# Docker
cat backup.sql | docker compose exec -T postgres psql -U designflow designflow_db
```

---

## Performance

### Índices

Ya existen en:
- `pages.edicion_id`
- `pautas.pagina_id`

### Consultas Lentas

Si una query es lenta, analizar con:

```sql
EXPLAIN ANALYZE
SELECT ...
```

---

## Escalabilidad

### En AWS (RDS)

Solo cambias la conexión:

```typescript
// De: localhost:5432
// A: designflow-db.cxxxxxxx.us-east-1.rds.amazonaws.com:5432
```

El schema y las queries siguen siendo idénticas.

---

## Troubleshooting

### Error: "cannot connect to database"

```bash
# Verificar que PostgreSQL está corriendo
docker compose ps

# Verificar logs
docker compose logs postgres
```

### Error: "password authentication failed"

- Verificar `DB_USER` y `DB_PASSWORD` en `.env`
- Asegurar que coinciden con `POSTGRES_USER` y `POSTGRES_PASSWORD` en docker-compose

### Error: "table does not exist"

- Las tablas se crean desde `backend/database/init.sql`
- Verificar que ese archivo existe
- Limpiar datos: `docker compose down -v` (borra volumen) y levantar de nuevo

---

**Última actualización:** 2026-01-10
