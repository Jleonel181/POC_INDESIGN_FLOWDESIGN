# idml_generator

Genera archivos IDML a partir del JSON de layout del backend (`/api/layout/:editionId`).

## Arquitectura

Módulo hexagonal con separación estricta de capas:

```
src/
├── modules/idml/
│   ├── domain/
│   │   ├── entities.py          # LayoutDocument, SpreadPage, Frame
│   │   ├── ports.py             # IdmlGeneratorPort (interfaz de salida)
│   │   └── layout_fetcher_port.py  # LayoutFetcherPort (interfaz de entrada)
│   ├── application/
│   │   └── generate_idml_use_case.py  # Orquesta fetch + generate
│   └── infrastructure/
│       ├── idml_xml_builder.py        # Construye XML IDML con lxml
│       ├── idml_generator_adapter.py  # Ensambla el ZIP .idml
│       ├── layout_api_adapter.py      # Fetch desde backend HTTP
│       └── layout_json_adapter.py     # Fetch desde archivo JSON local
└── container.py  # Inyección de dependencias
```

## Setup

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Uso

### Desde el backend (requiere backend corriendo en localhost:3001)

```bash
python main.py --edition 12 --output output/edition_12.idml
```

### Desde un archivo JSON local

```bash
python main.py --json sample_layout.json --output output/edition_12.idml
```

### Variable de entorno

```bash
BACKEND_URL=http://localhost:3001 python main.py --edition 12 --output output/edition_12.idml
```

## Estructura del IDML generado

El archivo `.idml` es un ZIP con:

```
mimetype
designmap.xml
Resources/
  Fonts.xml
  Graphic.xml
  Styles.xml
  Preferences.xml
XML/
  Tags.xml
  BackingStory.xml
Spreads/
  Spread_1.xml   ← página 1 (sola, portada)
  Spread_2.xml   ← páginas 2-3 (spread)
  Spread_3.xml   ← páginas 4-5 (spread)
  ...
Stories/
  Story_<page_id>.xml  ← una por página
```

### Lógica de spreads (facing pages)

- Página 1: spread individual (portada)
- Páginas 2-3, 4-5, 6-7...: spreads de dos páginas
- Sin facing pages: cada página es su propio spread

### Coordenadas

- El backend devuelve bounds en mm con origen top-left de la página
- El generador convierte a puntos (1 mm = 2.834645669 pt)
- Las coordenadas en el Spread XML son relativas al centro del spread:
  - Página sola: `x_offset = -pageWidth/2`
  - Página izquierda: `x_offset = -pageWidth`
  - Página derecha: `x_offset = 0`
- `TextFrame.ItemTransform` = posición absoluta en el spread
- `TextFrame.GeometricBounds` = dimensiones locales del frame (`0 0 height width`)
