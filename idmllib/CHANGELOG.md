# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Trabajo de fidelidad: medir cuánta información sobrevive al ciclo de lectura y
escritura, y empezar a corregirlo. Ver [`docs/FIDELIDAD.md`](docs/FIDELIDAD.md) para
las cifras, cómo reproducirlas y los defectos conocidos.

### Added

- Arnés de fidelidad estructural en `pkg/idml/golden_test.go`: aplica el ciclo parseo → serialización → comparación a cada XML del corpus y reporta las diferencias por categoría, con un total agregado
- Corpus de fidelidad versionado en `testdata/`: `documento_referencia/` (43 XML, página de periódico real) y `archivo_evidencia_imagenes.idml` (única fuente de verdad del formato de imagen embebida)
- El arnés recorre los cinco IDML del repositorio: los dos anteriores más `plain.idml`, `example.idml` y `tripple.idml`, que ya estaban versionados y no se medían
- Rutas del corpus sustituibles por variable de entorno: `IDMLLIB_REFERENCE_DIR`, `IDMLLIB_IMAGES_FIXTURE`, `IDMLLIB_PLAIN_IDML`, `IDMLLIB_EXAMPLE_IDML`, `IDMLLIB_TRIPPLE_IDML`, más `IDMLLIB_MAX_DIFFS` para levantar el tope de diferencias por archivo
- Categorías de diferencia con nombre en `internal/xmlutil/compare.go`, contables y filtrables una por una: `atributo-ausente`, `atributo-valor-distinto`, `atributo-sobrante`, `elemento-ausente`, `elemento-sobrante`, `orden-elementos-distinto`, `texto-distinto`, `etiqueta-distinta`, `namespace-distinto`
- Marca de reporte truncado cuando un archivo alcanza el tope de diferencias
- `internal/xmlutil/childorder.go`: registro del orden documental de los hijos de un elemento, para reproducirlo al serializar sin convertir los campos por tipo en un contenedor
- `internal/xmlutil/attrs.go`: `UnmarshalAttrs` y `MarshalAttrs`, que reparten los atributos de un elemento entre los campos declarados del struct y un campo comodín `OtherAttrs`. Recorren los structs embebidos, respetan `omitempty` y admiten `string`, `int`, `uint`, `float` y `bool`
- `internal/testutil/childorder.go`: `AssertFieldOrderCovers`, que evita el fallo silencioso de olvidar una clase de hijo en el orden de campos
- `NewFromTemplate()` emite los 4 archivos que le faltaban: `Spreads/`, `Stories/`, `XML/BackingStory.xml` y `META-INF/metadata.xml`, con sus plantillas en `pkg/idml/templates/minimal/`
- Constantes de ruta `PathMetadata`, `PathSpread` y `PathStory`
- Sonda para verificación manual en InDesign: `IDMLLIB_PROBE_DIR=<dir> go test ./pkg/idml/ -run SondaParaInDesign`

### Changed

- **`NewFromTemplate()` produce un documento con página.** Antes emitía 9 archivos y su `designmap.xml` no declaraba ninguna referencia `idPkg:Spread`, es decir un documento sin ninguna página, una forma que ninguna exportación de InDesign tiene. Ahora emite las 13 entradas de `testdata/plain.idml` y declara `idPkg:Spread`, `idPkg:Story` e `idPkg:BackingStory`
- **`DocumentPreference` se emite en `Resources/Preferences.xml`** y no en `designmap.xml`, que es donde lo pone InDesign. Con él llegan `MarginPreference` y `ViewPreference`, que faltaban en la plantilla
- El cierre referencial del paquete generado está completo y verificado: `ItemLayer` → `Layer`, `ParentStory` → story emitida, `PageStart` del `Section` → página emitida, `AppliedMaster` → master spread, y `StoryList` con la story y la backing story
- La geometría del marco de texto y `ColumnsPositions` se calculan de las opciones en lugar de estar fijos
- El comparador empareja los hijos por nombre y ordinal en lugar de por posición absoluta, de modo que un reordenamiento no desalinea la comparación ni produce diferencias de atributo entre elementos que no se corresponden
- Los atributos se comparan y se emiten en orden alfabético, para que dos ejecuciones den el mismo reporte
- `FormatDifferences` emite sus mensajes en español, de forma consistente con el resto de la salida del arnés
- Fixtures dorados de `pkg/idms/testdata/golden/` regenerados: fijaban la salida con los hijos reordenados

### Fixed

- **Orden documental de los hijos en `designmap.xml`, `Resources/Styles.xml` y `Resources/Graphic.xml`.** Se emitían agrupados por el orden de los campos del struct en lugar de en el orden del documento: el elemento `Document` con sus 118 hijos, `TOCStyle` saliendo en sexta posición cuando iba en tercera, y `Gradient`, `Swatch` y `PastedSmoothShade` reagrupados
- **Orden documental en la exportación IDMS.** Los dos snippets `.idms` de `testdata/` se reescribían con los hijos de su `<Document>` reordenados. Ahora la secuencia es idéntica a la de la entrada
- El desorden de hijos se reporta como **una** diferencia en el nodo padre en lugar de una cascada por cada posición desplazada. El `designmap.xml` del Archivo_Evidencia_Imagenes producía 96 diferencias por un único defecto
- `TestFormatDifferences` estaba en rojo: `compare.go` devolvía un mensaje en español y el test lo esperaba en inglés. El fallo estaba oculto por la caché de tests de Go
- `NewFromTemplate()` rechaza con un error legible los márgenes que no dejan área utilizable y el número de columnas que no cabe en el ancho disponible, en lugar de emitir un marco de tamaño negativo
- El desplazamiento vertical de la página del master spread mínimo sigue la convención de InDesign para una página única

### Documentation

- Nuevo [`docs/FIDELIDAD.md`](docs/FIDELIDAD.md): estado medido de la fidelidad, cómo reproducirlo, los dos defectos conocidos y lo que sigue sin verificar
- Corregido en `README.md` y `ARCHITECTURE.md` el reclamo de «roundtrip perfecto» y «preservación byte a byte del ZIP», que las mediciones contradicen
- `pkg/idml/templates/README.md` actualizado a los 13 archivos de la plantilla mínima, con el resultado de la verificación manual en InDesign

### Known issues

- **13995 atributos se pierden** en el ciclo sobre el corpus de cinco documentos, porque los tipos del modelo todavía no declaran el campo comodín. Las funciones que lo resuelven existen; conectarlas es trabajo pendiente
- **71 nodos** conservan sus hijos pero en otro orden: los spreads, las stories, y los tipos `Properties` y `ObjectStyleGroup`
- Los archivos bajo `MasterSpreads/` no se parsean, se copian tal cual
- Solo se ha verificado en InDesign un documento **generado desde cero**, no un documento existente reescrito por la librería

### Security

## [2.2.0] - 2026-01-20

### Added
- PDF support for Rectangle page items with `PDF` struct and `PDFAttribute`
- `FrameContentBase` shared structure for Image and PDF content types
- Support for extracting PDF links from `Rectangle/PDF/Link@LinkResourceURI`
- PDF-specific color policy attributes: `GrayVectorPolicy`, `RGBVectorPolicy`, `CMYKVectorPolicy`
- PDF page selection and cropping attributes via `PDFAttribute`

### Changed
- Refactored `Image` and `PDF` structs to use shared `FrameContentBase` (eliminates 13 duplicated attributes)
- `Rectangle` now supports both `Image` and `PDF` child elements
- Updated test fixtures to reflect new attribute ordering in marshaled XML

### Fixed
- Fixed test compilation errors after FrameContentBase refactoring
- Updated golden test files for new XML attribute ordering

## [2.1.0] - 2025-01-15

### Added
- New `pkg/xmp` package for XMP (Extensible Metadata Platform) metadata support
- XMP metadata parsing and extraction from IDML and IDMS files
- `XMP()` and `SetXMP()` methods on IDML and IDMS Package types
- XMP timestamp management with `UpdateTimestamps()` method
- XMP thumbnail management with `RemoveThumbnails()` and `AddThumbnail()` methods
- XMP field access operations with `GetField()` and `SetField()` methods
- Automatic XMP metadata persistence when writing IDML/IDMS files

### Changed
- Updated `.golangci.yml` to v2 configuration format for golangci-lint v2.8.0 compatibility
- Updated Go version in linter config from 1.21 to 1.23
- IDML and IDMS packages now automatically extract and preserve XMP metadata during read/write operations

### Fixed
- Fixed `.golangci.yml` configuration errors (version field, output format, deprecated linters)
- Removed deprecated linters: `typecheck`, `gofmt`, `goimports`, `gosimple`

### Security

## [2.0.0] - 2025-01-13

### Added

- Complete IDML read/write support with roundtrip fidelity
- Domain-driven package architecture mirroring IDML file structure
- Story, Spread, and Document parsing with full XML support
- ResourceManager for tracking, validating, and cleaning up resources
- Selection API for programmatically selecting elements by ID
- IDMS snippet export functionality
- DependencyTracker for analyzing element dependencies
- Golden file testing infrastructure
- CLI tool with interactive TUI interface
- Zero external dependencies (Go stdlib only)

### Changed
- **BREAKING**: Updated module path to `github.com/dimelords/idmllib/v2` following Go module versioning semantics for v2.0.0
- Updated all import statements to use v2 module path
- Updated README.md examples and installation instructions for v2

### Package Structure

- `pkg/idml` - Main coordinator and public API (file I/O, caching)
- `pkg/document` - Document/designmap types and parsing
- `pkg/story` - Story types and parsing
- `pkg/spread` - Spread types and page item parsing
- `pkg/resources` - Resource file types (Fonts, Styles, Graphics)
- `pkg/analysis` - Cross-domain dependency tracking
- `pkg/idms` - IDMS snippet export
- `pkg/common` - Shared types (Properties, RawXMLElement, etc.)
- `internal/xmlutil` - XML comparison and formatting utilities
- `internal/testutil` - Test helpers and golden file support

### Testing

- Overall test coverage: 74%
- Critical path coverage: 93.6% (pkg/analysis)
- Comprehensive roundtrip tests
- Golden file validation
- 158/161 tests passing

