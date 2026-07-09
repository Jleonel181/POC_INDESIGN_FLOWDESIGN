# PLANTILLAS_DB_REDAPP

Documentación del sistema de plantillas de página en la base de datos `redaccion`.
Resultado del análisis exploratorio de tablas y datos reales.

---

## Contexto

La BD `redaccion` pertenece a un sistema editorial/periodístico llamado **RedApp**.
Dentro de ella existe un subsistema de **planificación de páginas** que permite organizar
qué noticias (notas) van en cada página de una edición, usando plantillas visuales
que definen cuántos artículos caben y cómo se distribuyen espacialmente.

---

## Tablas involucradas

### `plantilla_tipo` — Catálogo de plantillas

Define los tipos de plantilla disponibles. Es una tabla de catálogo: **sus datos no cambian**, son fijos del sistema.

| Columna         | Tipo    | Descripción |
|-----------------|---------|-------------|
| `idplantilla_tipo` | INT  | PK |
| `nombre`        | VARCHAR | Nombre legible, ej: `Tipo 1-1`, `Tipo 4-5` |
| `cantidad`      | INT     | **Número de slots (espacios para notas)** que tiene esta plantilla |
| `html`          | TEXT    | HTML del thumbnail visual (previsualización en UI) |
| `plantilla_html`| TEXT    | HTML de la plantilla real, con placeholders `{{{1}}}`, `{{{2}}}`, etc. |
| `icono`         | VARCHAR | Icono opcional |
| `estado`        | VARCHAR | `Activo` / `Inactivo` |

**Convención de nombres**: `Tipo X-Y` donde `X` = cantidad de slots, `Y` = variante de layout.
- `Tipo 1-1` → 1 slot
- `Tipo 2-3` → 2 slots, variante 3
- `Tipo 4-5` → 4 slots, variante 5

**Placeholders en `plantilla_html`**: cada slot se representa como `{{{N}}}` donde N es el número de orden del slot (1, 2, 3...). Al renderizar, se reemplaza cada placeholder con el contenido de la nota asignada a ese slot.

Hay **47 plantillas** activas en total, agrupadas por cantidad de slots del 1 al 6.

---

### `planificacion` — Cabecera de planificación

Representa una planificación editorial (una edición/fecha). Soporta jerarquía padre/hijo mediante `idpadre` (auto-referencial).

---

### `planificacion_pagina` — Páginas de una planificación

Cada fila es una página dentro de una planificación.

| Columna              | Descripción |
|----------------------|-------------|
| `idplanificacion_pagina` | PK |
| `idplanificacion`    | FK a `planificacion` |
| `numero_pagina`      | Número de página (1, 2, 3...) |
| `idplantilla`        | Referencia a `plantilla` (sin FK formal) |
| `idplantilla_tipo`   | **Referencia a `plantilla_tipo.idplantilla_tipo`** (sin FK formal) |

> **Importante**: `idplantilla` e `idplantilla_tipo` no tienen FK declarada en la BD.
> La integridad referencial se maneja en la capa de aplicación.

---

### `planificacion_pagina_nota` — Slots de cada página

**Esta es la tabla central del sistema de slots.** Cada fila representa un espacio (slot) dentro de una página, con o sin nota asignada.

| Columna                      | Descripción |
|------------------------------|-------------|
| `idplanificacion_pagina_nota`| PK |
| `idplanificacion_pagina`     | FK a `planificacion_pagina` |
| `idnota`                     | FK a la nota/artículo asignado. **NULL = slot vacío** |
| `orden`                      | Número de slot dentro de la página (1, 2, 3...) |
| `comentario`                 | Comentario opcional |
| `canal_impreso`              | `Activo` / `Inactivo` |
| `canal_digital`              | `Activo` / `Inactivo` |
| `canal_web`                  | `Activo` / `Inactivo` |

La cantidad de filas por página en esta tabla **debe coincidir** con `plantilla_tipo.cantidad` de la plantilla asignada a esa página.

---

## Relación entre tablas (sin FKs formales)

```
plantilla_tipo
  idplantilla_tipo  ←──────────────────────────────────┐
  cantidad (N slots)                                    │ (sin FK formal)
  plantilla_html ({{{1}}}, {{{2}}}, ..., {{{N}}})       │
                                                        │
planificacion                                           │
  idplanificacion                                       │
       │                                                │
       ▼                                                │
planificacion_pagina                                    │
  idplanificacion_pagina                                │
  numero_pagina                                         │
  idplantilla_tipo ─────────────────────────────────────┘
       │
       ▼
planificacion_pagina_nota  (N filas = N slots)
  idplanificacion_pagina
  orden (1..N)
  idnota  →  nota/artículo (NULL si vacío)
  canal_impreso / canal_digital / canal_web
```

---

## Ejemplo real (planificación 2026-07-01)

```
Página 1 → idplantilla_tipo=1  (Tipo 1-1, 1 slot)  → 1 fila en pagina_nota, orden=1, idnota=569566
Página 2 → idplantilla_tipo=1  (Tipo 1-1, 1 slot)  → 1 fila en pagina_nota, orden=1, idnota=NULL
Página 3 → idplantilla_tipo=1  (Tipo 1-1, 1 slot)  → 1 fila en pagina_nota, orden=1, idnota=NULL
Página 4 → idplantilla_tipo=1  (Tipo 1-1, 1 slot)  → 1 fila en pagina_nota, orden=1, idnota=NULL
Página 5 → idplantilla_tipo=30 (Tipo 4-5, 4 slots) → 4 filas en pagina_nota, orden=1,2,3,4, idnota=NULL
```

---

## Cómo leer una página completa

Para obtener una página con su plantilla y sus slots:

```sql
USE redaccion;

SELECT
  pp.numero_pagina,
  pt.nombre        AS plantilla_nombre,
  pt.cantidad      AS total_slots,
  pt.plantilla_html,
  ppn.orden        AS slot,
  ppn.idnota,
  ppn.canal_impreso,
  ppn.canal_digital,
  ppn.canal_web
FROM planificacion p
JOIN planificacion_pagina pp     ON pp.idplanificacion = p.idplanificacion
JOIN plantilla_tipo pt           ON pt.idplantilla_tipo = pp.idplantilla_tipo
JOIN planificacion_pagina_nota ppn ON ppn.idplanificacion_pagina = pp.idplanificacion_pagina
WHERE DATE(p.fecha) = '2026-07-01'
ORDER BY pp.numero_pagina, ppn.orden;
```

---

## Reglas del sistema (confirmadas con datos reales)

1. Cada página tiene **exactamente una plantilla** (`idplantilla_tipo`).
2. La plantilla define cuántos slots tiene la página (`cantidad`).
3. Los slots se crean en `planificacion_pagina_nota` con `orden` de 1 a N.
4. Un slot sin nota asignada tiene `idnota = NULL`.
5. El `plantilla_html` usa `{{{1}}}` ... `{{{N}}}` como placeholders para renderizar el contenido de cada slot.
6. Las plantillas son un catálogo fijo: no cambian entre planificaciones.
7. No hay FKs formales entre `planificacion_pagina` y `plantilla_tipo`; la integridad la garantiza la aplicación.

---

## Tablas descartadas (vacías o no usadas en este flujo)

| Tabla                        | Estado   | Motivo |
|------------------------------|----------|--------|
| `planificacion_pagina_detalle` | Sin datos | No se usa en planificaciones actuales |
| `plantilla_tipo_detalle`     | Sin datos | No se usa en planificaciones actuales |
| `planificacion_pagina_archivo_objeto` | No investigada | Posiblemente para adjuntos/imágenes |
