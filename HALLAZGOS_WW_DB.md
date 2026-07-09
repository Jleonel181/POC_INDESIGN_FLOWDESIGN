# Hallazgos: Base de Datos WoodWing Enterprise

## Contexto

**WoodWing Enterprise** es el sistema de gestión editorial que usa el equipo de redacción. Desde WoodWing, los periodistas crean artículos y los diseñadores los colocan en páginas de InDesign. Toda esa información queda registrada en una base de datos MySQL.

Este documento explica cómo leer esa base de datos para saber **qué artículos existen, en qué página están diagramados y cómo están distribuidos sus frames en InDesign**.

---

## Tablas involucradas

Todas las tablas tienen el prefijo `smart_`:

| Tabla | Rol |
|---|---|
| `smart_objects` | Repositorio central. Contiene todos los objetos: artículos, layouts, dossiers, imágenes |
| `smart_objectrelations` | Relaciones entre objetos (quién contiene a quién, qué está colocado dónde) |
| `smart_placements` | Frames de InDesign: posición, tamaño y estado de cada elemento en la página |
| `smart_pages` | Páginas físicas de cada layout con sus dimensiones y master page |

---

## Jerarquía de objetos en WoodWing

Cada pieza editorial sigue esta jerarquía:

```
Dossier         → agrupa todo lo relacionado a una edición/nota
    └── Layout  → el archivo .indd de InDesign (una o más páginas)
            └── Article → el archivo .icml con el contenido editorial
                    └── placements → los frames individuales en la página
```

- Un **Dossier** es el contenedor organizativo (carpeta editorial)
- Un **Layout** es el archivo InDesign (`.indd`) con las páginas físicas
- Un **Article** es el contenido editorial (`.icml`) que se coloca dentro del layout
- Los **placements** son cada frame de texto o gráfico dentro de la página

---

## Ejemplo completo — "QPASA GUATE 04 20260625 FE"

### Paso 1 — Encontrar el artículo en `smart_objects`

```sql
SELECT id, type, name, publication, section, state, format, storename, majorversion
FROM smart_objects
WHERE id = 3894855;
```

| Campo | Valor | Descripción |
|---|---|---|
| id | 3894855 | Identificador único del objeto |
| type | Article | Es un artículo editorial (.icml) |
| name | QPASA GUATE 04 20260625 FE | Nombre del artículo |
| publication | 2 | ID de la publicación |
| section | 11 | ID de la sección |
| state | 29 | Estado del flujo editorial |
| format | application/incopyicml | Formato InCopy |
| storename | 3/89/48/3894855 | Ruta física del archivo en el servidor |
| majorversion | 3 | Versión actual del archivo |

---

### Paso 2 — Encontrar los layouts de esa sección y fecha

```sql
SELECT id, name, pagerange, state, storename
FROM smart_objects
WHERE type = 'Layout'
AND publication = 2
AND section = 11;
```

Resultado — layouts de la edición 20260625, sección 11:

| id | name | pagerange | storename |
|---|---|---|---|
| 3893845 | QPASA GUATE 02 20260625 | 002 | 3/89/38/3893845 |
| 3893847 | QPASA GUATE 03 20260625 | 001 | 3/89/38/3893847 |
| **3893849** | **QPASA GUATE 04 20260625** | **004** | **3/89/38/3893849** |
| 3893851 | QPASA GUATE 05 20260625 | 005 | 3/89/38/3893851 |
| 3893853 | QPASA GUATE 06 20260625 | 006 | 3/89/38/3893853 |
| 3893855 | QPASA GUATE 07 20260625 | 007 | 3/89/38/3893855 |

El layout `3893849` es el candidato porque comparte el número "04" con el artículo.

---

### Paso 3 — Confirmar la relación artículo → layout en `smart_objectrelations`

```sql
SELECT * FROM smart_objectrelations WHERE child = 3894855;
```

| id | parent | child | type | pagerange |
|---|---|---|---|---|
| 5888699 | 3893850 | 3894855 | Contained | |
| 5888700 | 3893849 | 3894855 | Placed | ,4 |

- `Contained` (parent=`3893850`) → el artículo pertenece al **Dossier** `3893850`
- `Placed` (parent=`3893849`) → el artículo está diagramado en el **Layout** `3893849`, página `4`

---

### Paso 4 — Información del Dossier contenedor

```sql
SELECT id, type, name, state FROM smart_objects WHERE id = 3893850;
```

| id | type | name | state |
|---|---|---|---|
| 3893850 | Dossier | QPASA GUATE 04 20260625 FE | 54 |

---

### Paso 5 — Página física del layout en `smart_pages`

```sql
SELECT * FROM smart_pages WHERE objid = 3893849;
```

| id | objid | width | height | pagenumber | pageorder | master | instance |
|---|---|---|---|---|---|---|---|
| 3194972 | 3893849 | 774.0 | 972.0 | 07-ORI-4 | 4 | Folio QUÉ PASA | Production |

- Tamaño: **774 × 972 pt** (~27.3 × 34.3 cm, formato tabloide)
- Master page: **Folio QUÉ PASA**
- Número de página física: **07-ORI-4**, orden `4`

---

### Paso 6 — Frames del artículo en la página (`smart_placements`)

```sql
SELECT * FROM smart_placements WHERE child = 3894855 AND parent = 3893849;
```

33 frames en página `4`, todos en el layer `APERTURA`:

| element | frameid | _left | top | width | height | overset |
|---|---|---|---|---|---|---|
| N1_02_titulo | 115970 | 48.7 | 99.9 | 484.6 | 32.0 | 0 |
| N1_03_subtitulo | 115993 | 48.7 | 140.2 | 519.0 | 14.4 | -2 ⚠️ |
| N1_05_credito_nota | 115794 | 47.8 | 164.4 | 127.6 | 22.5 | -20 ⚠️ |
| N1_06_cuerpo_texto | 116016 | 46.2 | 197.9 | 127.6 | 253.0 | 0 |
| N1_06_cuerpo_texto | 116115 | 187.6 | 384.9 | 382.4 | 66.0 | 0 |
| N1_09_pie_foto_02 | 116190 | 592.0 | 890.9 | 141.0 | 22.0 | -14 ⚠️ |
| N2_02_titulo | 116069 | 247.2 | 486.4 | 146.8 | 51.6 | -3 ⚠️ |
| N2_06_cuerpo_texto | 116092 | 247.5 | 582.9 | 146.5 | 121.0 | 0 |
| N2_06_cuerpo_texto | 116138 | 414.5 | 483.9 | 146.2 | 220.0 | 0 |
| N2_09_pie_foto_01 | 116046 | 48.7 | 681.4 | 183.9 | 22.5 | -19 ⚠️ |
| N5_08_foto_01 | 115719 | 573.0 | 93.1 | 116.1 | 72.2 | 0 |
| graphic | 114877 | 339.6 | 35.4 | 180.0 | 48.0 | 0 |
| graphic | 115723 | 573.4 | 69.7 | 165.7 | 97.6 | 0 |

Convención de nombres de elementos:
- `N1_*` → frames de la nota principal
- `N2_*` → frames de la nota secundaria
- `N5_*` → frames de foto
- `body` → frames genéricos de texto
- `graphic` → frames de imagen

> ⚠️ `overset < 0` = texto desbordado, no cabe en el frame  
> Frames con `page=0` están en el **pasteboard** (overflow fuera de página, `frameorder=2`)

---

### Paso 7 — Verificar si hay imágenes gestionadas por WoodWing

```sql
SELECT id, type, name, state, storename
FROM smart_objects
WHERE type = 'Image'
AND id IN (
    SELECT child FROM smart_objectrelations WHERE parent = 3893850
);
```

Resultado: **vacío**. Las fotos están embebidas directamente en el `.indd`/`.icml` y no son objetos independientes en WoodWing.

---

### Paso 8 — Ver todos los objetos diagramados en el layout

```sql
SELECT DISTINCT p.child, so.type, so.name, so.state, p.page, p.layer
FROM smart_placements p
JOIN smart_objects so ON so.id = p.child
WHERE p.parent = 3893849
ORDER BY p.page, so.type;
```

| child | type | name | state | page | layer |
|---|---|---|---|---|---|
| 3894855 | Article | QPASA GUATE 04 20260625 FE | 29 | 0 | APERTURA |
| 3894855 | Article | QPASA GUATE 04 20260625 FE | 29 | 4 | APERTURA |

- `page=4` → frames reales en la página
- `page=0` → frames en el pasteboard (overflow)
- **Conclusión:** layout de una sola página con un único artículo que la ocupa completa

---

## Jerarquía final confirmada

```
smart_objects: Dossier  3893850  "QPASA GUATE 04 20260625 FE"  (state=54)
    └── smart_objects: Layout   3893849  "QPASA GUATE 04 20260625"  (state=30)
            │   └── smart_pages: página 4, 774×972pt, master "Folio QUÉ PASA"
            └── smart_objectrelations: Placed → child=3894855, page=4
                    └── smart_objects: Article  3894855  "QPASA GUATE 04 20260625 FE"  (state=29)
                            └── smart_placements: 33 frames en página 4, layer APERTURA
                                    ├── texto encadenado: N1_06_cuerpo_texto (frameorder 0,1,2)
                                    ├── título, subtítulo, créditos, pies de foto
                                    ├── frames gráficos: graphic, N5_08_foto_01
                                    └── fotos: embebidas en .indd, no en WoodWing
```

---

## Consultas de referencia reutilizables

```sql
-- 1. Buscar artículo por nombre o fecha
SELECT id, type, name, section, state, storename
FROM smart_objects
WHERE type = 'Article' AND name LIKE '%GUATE 04%';

-- 2. Layouts de una publicación y sección
SELECT id, name, pagerange, state
FROM smart_objects
WHERE type = 'Layout' AND publication = 2 AND section = 11;

-- 3. Jerarquía completa de un artículo
SELECT * FROM smart_objectrelations WHERE child = <article_id>;

-- 4. Página física de un layout
SELECT * FROM smart_pages WHERE objid = <layout_id>;

-- 5. Todos los frames de un artículo en su layout
SELECT * FROM smart_placements
WHERE child = <article_id> AND parent = <layout_id>;

-- 6. Todos los objetos diagramados en un layout
SELECT DISTINCT p.child, so.type, so.name, so.state, p.page, p.layer
FROM smart_placements p
JOIN smart_objects so ON so.id = p.child
WHERE p.parent = <layout_id>
ORDER BY p.page, so.type;

-- 7. Imágenes gestionadas por WoodWing en un dossier
SELECT id, type, name, storename
FROM smart_objects
WHERE type = 'Image'
AND id IN (
    SELECT child FROM smart_objectrelations WHERE parent = <dossier_id>
);
```
