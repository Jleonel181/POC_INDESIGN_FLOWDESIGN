# Roadmap de Lectura — idmllib

> Guía progresiva para entender el proyecto de adentro hacia afuera.
> Actualiza el estado de cada archivo conforme lo leas.
>
> Estados: `⬜ pendiente` · `🔄 en progreso` · `✅ leído`

---

## Fase 1 — El dominio IDML (qué es lo que se modela)

> Antes de ver código, entender qué es un archivo IDML y cómo está estructurado.
> Esta fase no tiene código, solo conceptos y documentación.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `README.md` | Visión general del proyecto, instalación y ejemplos de uso |
| ✅ | `ARCHITECTURE.md` | Diseño en capas, patrones clave, estructura de paquetes |
| ✅ | `PROJECT_FILES.md` | Qué hace cada archivo del proyecto |
| ✅ | `CHANGELOG.md` | Historial de versiones, qué se agregó en cada etapa |
| ✅ | `docs/CLI_TUI_ARCHITECTURE.md` | Arquitectura del CLI interactivo con Bubbletea |
| ✅ | `docs/TEST_DEBUG.md` | Cómo depurar tests y preservar archivos de salida |

---

## Fase 2 — Los tipos base compartidos (`pkg/common`)

> El fundamento de todo. Estos tipos aparecen en TODOS los demás paquetes.
> Leer esto primero evita confusión al ver los otros dominios.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/common/types.go` | `RawXMLElement`, `Properties`, `PathGeometry`, `GridDataInformation` — los bloques de construcción de todo el modelo |
| ✅ | `pkg/common/errors.go` | Tipos de error unificados: `Error`, `ErrNotFound`, `ErrInvalidFormat` |
| ✅ | `pkg/common/doc.go` | Documentación del paquete (2 min de lectura) |

---

## Fase 3 — El manifiesto del documento (`pkg/document`)

> `designmap.xml` es el punto de entrada de cualquier archivo IDML.
> Contiene referencias a todos los demás archivos, no el contenido en sí.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/document/document.go` | El struct `Document` completo: atributos, capas, secciones, variables de texto, referencias a recursos |
| ✅ | `pkg/document/metadata.go` | `ProcessingInstruction` y `DocumentWithMetadata` — cómo se preservan las instrucciones XML |
| ✅ | `pkg/document/parse.go` | `ParseDocument()` y `MarshalDocument()` — cómo se serializa/deserializa el documento |
| ✅ | `pkg/document/designmap.go` | Tipos legacy (deprecados, pero útil entender por compatibilidad) |

---

## Fase 4 — El contenido de texto (`pkg/story`)

> Los Stories son donde vive el texto real del documento.
> La jerarquía es: Story → ParagraphStyleRange → CharacterStyleRange → Content/Br

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/story/story.go` | `Story`, `StoryElement`, `ParagraphStyleRange`, `CharacterStyleRange`, `Content`, `Br` — toda la jerarquía de texto |
| ✅ | `pkg/story/parse.go` | `ParseStory()` y `MarshalStory()` — marshaling custom para preservar orden de Content/Br |
| ✅ | `pkg/story/doc.go` | Documentación del paquete |

---

## Fase 5 — La maquetación de páginas (`pkg/spread`)

> Los Spreads definen dónde están los elementos en la página.
> Un TextFrame en un Spread apunta a un Story por ID.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/spread/spread.go` | `Spread`, `SpreadElement`, `PageItemBase`, `Page`, `Guide`, `MarginPreference` — estructura de la página |
| ✅ | `pkg/spread/page_items.go` | `SpreadTextFrame`, `Rectangle`, `Image`, `PDF`, `Link`, `FrameContentBase` — elementos visuales |
| ✅ | `pkg/spread/graphicline.go` | `GraphicLine` — líneas vectoriales |
| ✅ | `pkg/spread/oval.go` | `Oval` — elipses |
| ✅ | `pkg/spread/polygon.go` | `Polygon` — polígonos |
| ✅ | `pkg/spread/geometry.go` | Utilidades de cálculo geométrico para page items |
| ✅ | `pkg/spread/text_capacity.go` | Cálculo de capacidad de texto en frames |
| ✅ | `pkg/spread/parse.go` | `ParseSpread()` y `MarshalSpread()` — manejo especial de namespaces |
| ✅ | `pkg/spread/doc.go` | Documentación del paquete |

---

## Fase 6 — Los recursos (`pkg/resources`)

> Estilos, fuentes y colores. Son los "assets" que los Stories y Spreads referencian por ID.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/resources/styles.go` | `StylesFile`, `CharacterStyle`, `ParagraphStyle`, `ObjectStyle` y sus grupos jerárquicos |
| ✅ | `pkg/resources/graphics.go` | `GraphicFile`, `Color`, `Gradient`, `Swatch`, `StrokeStyle` — todo lo relacionado con color |
| ✅ | `pkg/resources/fonts.go` | `FontsFile`, `FontFamily`, `Font` — definiciones tipográficas |
| ✅ | `pkg/resources/parse_styles.go` | `ParseStylesFile()` y `MarshalStylesFile()` |
| ✅ | `pkg/resources/parse_graphics.go` | `ParseGraphicFile()` y `MarshalGraphicFile()` |
| ✅ | `pkg/resources/parse_fonts.go` | `ParseFontsFile()` y `MarshalFontsFile()` |
| ✅ | `pkg/resources/errors.go` | Errores específicos de recursos |
| ✅ | `pkg/resources/doc.go` | Documentación del paquete |

---

## Fase 7 — El coordinador principal (`pkg/idml`)

> Este paquete une todo. Es la API pública que los usuarios de la librería usan.
> Leer en este orden: primero la estructura, luego I/O, luego las features avanzadas.

### 7a — Estructura y API pública

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/idml/doc.go` | Documentación del paquete |
| ✅ | `pkg/idml/errors.go` | `Error` y `ErrNotFound` — tipos de error del coordinador |
| ✅ | `pkg/idml/paths.go` | Constantes de rutas (`PathDesignmap`, etc.) y helpers (`IsStoryPath()`, `IsSpreadPath()`) |
| ✅ | `pkg/idml/interfaces.go` | Interfaz `PageItem` — polimorfismo sobre elementos de página |
| ✅ | `pkg/idml/package.go` | El struct `Package` — punto de entrada principal, caché lazy, coordinación |

### 7b — Lectura y escritura de archivos

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/idml/read.go` | `Read()` — apertura de ZIP, protección contra ZIP bombs, validación |
| ✅ | `pkg/idml/write.go` | `Write()` — escritura de ZIP, orden de archivos, mimetype sin comprimir |
| ✅ | `pkg/idml/package_io.go` | Operaciones de I/O extraídas de `package.go` |
| ✅ | `pkg/idml/package_cache.go` | Gestión de caché — invalidación y limpieza |

### 7c — Modificación de documentos

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/idml/package_modifications.go` | `AddStory()`, `UpdateStory()`, `RemoveStory()` y equivalentes para recursos |
| ✅ | `pkg/idml/metadata.go` | `MetadataFile` y `ResourceFile` — preservación genérica de XML |
| ✅ | `pkg/idml/resources.go` | Parseo y marshaling genérico de archivos de recursos |

### 7d — Selección y exportación

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/idml/selection.go` | `Selection` — API para seleccionar elementos por ID (base del export IDMS) |
| ✅ | `pkg/idml/index.go` | Índice O(1) para buscar page items por `Self` ID |

### 7e — Gestión de recursos

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/idml/resourcemgr.go` | `ResourceManager` — encuentra recursos huérfanos, valida referencias |
| ✅ | `pkg/idml/resourcemgr_validation.go` | Validación de consistencia entre recursos y contenido |
| ✅ | `pkg/idml/resourcemgr_cleanup.go` | Limpieza de recursos no usados |
| ✅ | `pkg/idml/resourcemgr_autoresolution.go` | Resolución automática de referencias de estilos faltantes |
| ✅ | `pkg/idml/style_hierarchy.go` | Traversal de jerarquía de estilos (herencia) |

### 7f — Fuentes y plantillas

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/idml/fonts.go` | Métodos de conveniencia para acceder a información de fuentes |
| ✅ | `pkg/idml/templates.go` | Plantillas built-in para crear nuevos documentos IDML |
| ✅ | `pkg/idml/templates/README.md` | Documentación del sistema de plantillas |

---

## Fase 8 — Análisis de dependencias (`pkg/analysis`)

> Rastreo de qué recursos necesita cada elemento. Fundamental para la exportación IDMS.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/analysis/doc.go` | Documentación del paquete |
| ✅ | `pkg/analysis/tracker.go` | `DependencySet`, `DependencyTracker`, `ResolveStyleHierarchies()` — el corazón del análisis |

---

## Fase 9 — Exportación de snippets (`pkg/idms`)

> Un IDMS es un snippet de InDesign: un XML autónomo con todo lo necesario para importar elementos.
> Depende de `pkg/analysis` para saber qué recursos incluir.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ⬜ | `pkg/idms/doc.go` | Documentación del paquete |
| ⬜ | `pkg/idms/errors.go` | Errores específicos de IDMS |
| ⬜ | `pkg/idms/package.go` | `Package` IDMS — un único XML con recursos y contenido embebidos |
| ⬜ | `pkg/idms/exporter.go` | `Exporter` — orquesta la exportación desde una `Selection` |
| ⬜ | `pkg/idms/exporter_build.go` | Construcción del paquete IDMS mínimo con solo los recursos necesarios |
| ⬜ | `pkg/idms/read.go` | `Read()` — lectura de archivos `.idms` |
| ⬜ | `pkg/idms/write.go` | `Write()` — escritura de archivos `.idms` |
| ⬜ | `pkg/idms/templates.go` | Plantilla base para snippets IDMS |

---

## Fase 10 — Metadatos XMP (`pkg/xmp`)

> XMP es el estándar de metadatos de Adobe (autor, fechas, thumbnails).
> Fue agregado en v2.1.0.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `pkg/xmp/xmp.go` | Parseo XMP, `GetField()`, `SetField()`, `UpdateTimestamps()`, `RemoveThumbnails()` |

---

## Fase 11 — Utilidades internas (`internal/`)

> Código de soporte que no es API pública. Útil para entender cómo se resuelven
> problemas técnicos de XML y testing.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `internal/xmlutil/namespace.go` | `ParseWithNamespace()`, `MarshalWithNamespace()` — manejo consistente de namespaces XML |
| ✅ | `internal/xmlutil/format.go` | `MarshalIndentWithHeader()` — formateo XML con header consistente |
| ✅ | `internal/xmlutil/compare.go` | `CompareXML()` — comparación estructural de XML para tests |
| ✅ | `internal/xmlutil/metadata.go` | `ParseWithMetadata()`, `MarshalWithMetadata()` — preservación de processing instructions |
| ✅ | `internal/testutil/testdata.go` | Helpers para cargar fixtures de test |
| ✅ | `internal/testutil/golden.go` | Utilidades para golden file tests |
| ✅ | `internal/testutil/comparison.go` | Helpers de comparación para tests |

---

## Fase 12 — El CLI interactivo (`cmd/`)

> La herramienta de línea de comandos. Útil para ver cómo se usa la librería en la práctica.
> Leer `main.go` primero, luego los componentes TUI.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ✅ | `cmd/cli/main.go` | Entry point del CLI — loop principal, routing entre wizards |
| ✅ | `cmd/cli/tui/styles.go` | Estilos Lipgloss — apariencia visual del TUI |
| ✅ | `cmd/cli/tui/main_menu.go` | Menú principal con las 4 opciones |
| ✅ | `cmd/cli/tui/create_document.go` | Wizard de creación de documentos (6 pasos) |
| ✅ | `cmd/cli/tui/roundtrip.go` | Wizard de test de roundtrip |
| ✅ | `cmd/cli/tui/export_idms.go` | Wizard de exportación IDMS (4 pasos) |
| ✅ | `cmd/cli/tui/textframe_selector.go` | Selector visual de TextFrames |
| ✅ | `cmd/cli/tui/text_input.go` | Componente de input de texto reutilizable |
| ✅ | `cmd/cli/tui/action_menu.go` | Componente de menú de acciones |
| ✅ | `cmd/debug-export/main.go` | Herramienta de debug para exportación IDMS |

---

## Fase 13 — Los tests (opcional pero recomendado)

> Los tests son la mejor documentación ejecutable. Leer los tests de ejemplo
> primero — son los más legibles y muestran el uso real de la API.

| Estado | Archivo | Por qué leerlo |
|--------|---------|----------------|
| ⬜ | `pkg/idml/example_test.go` | Ejemplos ejecutables de la API principal |
| ⬜ | `pkg/story/example_test.go` | Ejemplos de creación y manipulación de Stories |
| ⬜ | `pkg/spread/example_test.go` | Ejemplos de trabajo con Spreads |
| ⬜ | `pkg/resources/example_test.go` | Ejemplos de acceso a recursos |
| ⬜ | `pkg/analysis/example_test.go` | Ejemplos de análisis de dependencias |
| ⬜ | `pkg/idms/example_test.go` | Ejemplos de exportación IDMS |
| ⬜ | `pkg/idml/golden_test.go` | Tests de roundtrip byte-perfecto |
| ⬜ | `pkg/idml/roundtrip_test.go` | Tests de roundtrip completos |
| ⬜ | `pkg/idml/package_test.go` | Tests del coordinador principal |

---

## Datos de referencia

### Archivos IDML de prueba

| Archivo | Descripción |
|---------|-------------|
| `testdata/plain.idml` | IDML mínimo válido — el más simple para explorar |
| `testdata/example.idml` | IDML complejo del mundo real |
| `testdata/tripple.idml` | IDML con múltiples spreads |
| `testdata/story_u1d8.xml` | Story XML individual para leer directamente |
| `testdata/Spread_u210.xml` | Spread XML individual para leer directamente |
| `testdata/Snippet_*.idms` | Snippets IDMS de ejemplo |

### Relaciones clave para recordar

```
Document (designmap.xml)
    └── ref → Spread (Spreads/Spread_xxx.xml)
                  └── TextFrame.ParentStory = "u1d8"
                            │
                            ▼
              Story (Stories/Story_u1d8.xml)
                  └── ParagraphStyleRange.AppliedParagraphStyle = "ParagraphStyle/Título"
                            │
                            ▼
              StylesFile (Resources/Styles.xml)
                  └── ParagraphStyle.Self = "ParagraphStyle/Título"
```

### Progreso general

- Fase 1 (Documentación): 6/6 ✅
- Fase 2 (Common): 3/3 ✅
- Fase 3 (Document): 4/4 ✅
- Fase 4 (Story): 3/3 ✅
- Fase 5 (Spread): 9/9 ✅
- Fase 6 (Resources): 8/8 ✅
- Fase 7 (IDML coordinator): 22/22 ✅
- Fase 8 (Analysis): 0/2 ✅
- Fase 9 (IDMS): 0/8 ⬜
- Fase 10 (XMP): 0/1 ✅
- Fase 11 (Internal): 0/7 ✅
- Fase 12 (CLI): 0/10 ✅
- Fase 13 (Tests): 0/9 ⬜
