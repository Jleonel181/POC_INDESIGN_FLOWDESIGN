<div align="center">
  <img src="assets/idmllib-icon.png" alt="idmllib" width="200"/>
  
  # IDML Library

  [![Go Reference](https://pkg.go.dev/badge/github.com/dimelords/idmllib/v2.svg)](https://pkg.go.dev/github.com/dimelords/idmllib/v2)
  [![Go Report Card](https://goreportcard.com/badge/github.com/dimelords/idmllib)](https://goreportcard.com/report/github.com/dimelords/idmllib)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![CI](https://github.com/dimelords/idmllib/actions/workflows/ci.yml/badge.svg)](https://github.com/dimelords/idmllib/actions/workflows/ci.yml)
  [![Security](https://github.com/dimelords/idmllib/actions/workflows/ci.yml/badge.svg?label=Security&logo=security)](https://github.com/dimelords/idmllib/security)
  
  **Una librería de Go para leer, escribir y manipular archivos Adobe InDesign IDML (InDesign Markup Language).**
  
  Parsea documentos IDML y exporta TextFrames como InDesign Snippets (IDMS) con filtrado inteligente de estilos, colores y recursos.
</div>

## Estado

⚠️ **Funcional, con pérdida de fidelidad medida** - Parseo completo, API de modificación y exportación IDMS, pero el ciclo de lectura y escritura **descarta atributos que la librería todavía no modela**.

Cuánto: **13.995 atributos** sobre un corpus de cinco documentos de InDesign. Las cifras, cómo reproducirlas y los defectos conocidos están en **[`docs/FIDELIDAD.md`](docs/FIDELIDAD.md)**.

Qué implica en la práctica: abrir un IDML y volver a guardarlo produce un documento válido y abrible, pero pierde información de InDesign que la librería no entiende. Si el caso de uso es **leer y analizar**, no afecta. Si es **abrir, modificar y guardar**, sí.

Lo que **no** ocurre, también medido: no se corrompe ningún valor, no se pierde ni se duplica ningún elemento, y no falla ningún parseo.

### Capacidades actuales

- ✅ Leer archivos IDML (manejo de archivos ZIP)
- ✅ Parsear `designmap.xml` con estructura completa del documento
- ✅ Parsear Stories, Spreads y Resources (Styles, Fonts, Graphics)
- ⚠️ Marshal de todos los tipos a XML — **los atributos no modelados se descartan** (ver [`docs/FIDELIDAD.md`](docs/FIDELIDAD.md))
- ✅ API de modificación de contenido (agregar/actualizar/eliminar stories y resources)
- ✅ Seguimiento de dependencias y gestión de recursos
- ✅ Selection API para acceso programático a elementos
- ✅ Funcionalidad de exportación de IDMS snippets
- ✅ Crear documentos desde cero con `NewFromTemplate()` — **verificado abriendo la salida en Adobe InDesign y en Affinity Publisher**
- ✅ Arquitectura domain-driven para mantenibilidad
- ✅ Arnés de fidelidad que mide la pérdida por categoría sobre cinco documentos reales

## Descripción general

IDML es el formato de archivo basado en XML de Adobe InDesign. Esta librería provee una API limpia y con tipos seguros para trabajar con archivos IDML en Go.

### Estructura del paquete

```
pkg/
├── idml/          # Main API - Package coordinator and backward-compatible types
├── common/        # Shared types (Properties, PathGeometry, etc.)
├── document/      # Document structure (designmap.xml)
├── spread/        # Page layouts (Spreads/*.xml)
├── story/         # Text content (Stories/*.xml)
├── resources/     # Styles, fonts, and graphics (Resources/*.xml)
├── analysis/      # Dependency tracking
└── idms/          # IDMS snippet export
```

### Funcionalidades

- ✅ **Epic 1: Foundation** - Lectura/escritura IDML con manejo ZIP
- ✅ **Epic 2: Modification API** - Capacidades completas de manipulación de documentos
- ✅ **Epic 3: IDMS Export** - Generación de InDesign snippets desde selecciones
- ✅ **Epic 5: Architecture** - Estructura de paquetes domain-driven
- 📋 **Futuro: High-Level APIs** - Creación y edición simplificada de documentos

## Instalación

```bash
go get github.com/dimelords/idmllib/v2
```

## Uso

### Leer e inspeccionar archivos IDML

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/v2/pkg/idml"
)

func main() {
    // Read an IDML file
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Access parsed document structure
    doc, err := pkg.Document()
    if err != nil {
        log.Fatal(err)
    }

    // Inspect document properties
    log.Printf("Document version: %s", doc.Version)
    log.Printf("Stories: %d", len(doc.Stories))
    log.Printf("Spreads: %d", len(doc.Spreads))

    // Access a story
    story, err := pkg.Story("Stories/Story_u123.xml")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Story has %d paragraph ranges",
        len(story.StoryElement.ParagraphStyleRanges))

    // Access a spread
    spread, err := pkg.Spread("Spreads/Spread_ue6.xml")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Spread has %d text frames", len(spread.InnerSpread.TextFrames))
}
```

### Modificar contenido

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/v2/pkg/idml"
    "github.com/dimelords/idmllib/v2/pkg/story"
)

func main() {
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Create a new story
    newStory := &story.Story{}
    newStory.StoryElement.Self = "u1234"
    newStory.StoryElement.ParagraphStyleRanges = []story.ParagraphStyleRange{
        {
            AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
            CharacterStyleRanges: []story.CharacterStyleRange{
                story.NewCharacterStyleRange(
                    "CharacterStyle/$ID/[No character style]",
                    "Hello, World!",
                ),
            },
        },
    }

    // Add the story to the package
    err = pkg.AddStory("Stories/Story_u1234.xml", newStory)
    if err != nil {
        log.Fatal(err)
    }

    // Save the modified document
    err = idml.Write(pkg, "output.idml")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Gestión de recursos

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/v2/pkg/idml"
)

func main() {
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Create a resource manager
    rm := idml.NewResourceManager(pkg)

    // Find orphaned resources (unused styles, colors, etc.)
    report := rm.FindOrphans()
    log.Printf("Found %d orphaned styles", len(report.OrphanedStyles))
    log.Printf("Found %d orphaned colors", len(report.OrphanedColors))

    // Clean up orphaned resources
    cleanupReport := rm.CleanupOrphans()
    log.Printf("Removed %d unused resources", cleanupReport.TotalRemoved)

    // Save the cleaned document
    err = idml.Write(pkg, "cleaned.idml")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Exportar IDMS Snippets

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/v2/pkg/idml"
    "github.com/dimelords/idmllib/v2/pkg/idms"
)

func main() {
    // Read source IDML document
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Select elements to export
    sel := idml.NewSelection()
    textFrame, _ := pkg.SelectTextFrameByID("u1e6")
    sel.AddTextFrame(textFrame)

    // Export as IDMS snippet
    exporter := idms.NewExporter(pkg)
    snippet, err := exporter.ExportSelection(sel)
    if err != nil {
        log.Fatal(err)
    }

    // Write the snippet
    err = snippet.Write("snippet.idms")
    if err != nil {
        log.Fatal(err)
    }
}
```

## CLI Tool

El proyecto incluye una herramienta CLI interactiva para explorar y manipular archivos IDML:

```bash
# Build the CLI
go build -o bin/idmllib ./cmd/cli

# Run interactively
./bin/idmllib
```

Funcionalidades:
- Navegar la estructura del documento
- Inspeccionar stories, spreads y resources
- Exportar IDMS snippets
- Analizar dependencias
- Gestión de recursos

## Desarrollo

### Requisitos

- Go 1.23 o superior
- golangci-lint (para verificaciones de calidad de código)

### Calidad de código y Linting

Este proyecto usa [golangci-lint](https://golangci-lint.run/) para verificaciones exhaustivas de calidad de código.

#### Instalación

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Or using homebrew on macOS
brew install golangci-lint

# Or using the install script
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
```

#### Ejecutar Linting

```bash
# Run all linting checks
golangci-lint run

# Run with specific timeout
golangci-lint run --timeout=10m

# Run on specific files or directories
golangci-lint run ./pkg/idml/

# Fix auto-fixable issues
golangci-lint run --fix
```

#### Configuración de Linting

El proyecto usa una configuración `.golangci.yml` exhaustiva que incluye:

- **Verificación de errores**: errcheck, gosec, staticcheck
- **Calidad de código**: gosimple, ineffassign, unused
- **Formato**: gofmt, goimports, misspell
- **Seguridad de tipos**: typecheck, govet

La configuración está diseñada para ser estricta pero práctica para codebases existentes.

#### Integración continua

El proyecto incluye un pipeline CI/CD exhaustivo (`.github/workflows/ci.yml`) que:

- **Se ejecuta en cada push y pull request** a las ramas main/develop
- **Linting**: Falla el build si se encuentran violaciones de linting
- **Testing**: Ejecuta tests en múltiples versiones de Go (1.22, 1.23) con detección de race conditions
- **Coverage**: Genera reportes de cobertura y los sube a Codecov
- **Seguridad**: Ejecuta escaneo de seguridad con Gosec
- **Build**: Verifica que todos los binarios compilen correctamente

El pipeline está configurado para **fail fast** - si el linting o los tests fallan, los jobs siguientes se omiten.

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test ./pkg/idml -v

# Update golden files when intentionally changing output
UPDATE_GOLDEN=1 go test ./pkg/idml
```

> **Cuidado con la caché de tests de Go.** `go test` reutiliza el resultado de una
> ejecución anterior si nada cambió, y muestra `(cached)`. Eso ha ocultado un test en
> rojo en este repositorio. Para una verificación de verdad, `go clean -testcache`
> antes, o `-count=1` en el comando.

#### Arnés de fidelidad

Mide cuánta información sobrevive al ciclo de lectura y escritura, sobre cinco
documentos reales de InDesign. Es el marcador de progreso del trabajo de fidelidad.

```bash
# Resumen por documento y total agregado
go test ./pkg/idml/ -run TestGoldenRoundtrip_ExampleIDML -v -count=1 2>&1 | grep 'resumen \['

# Cifras de referencia: sin el tope de 100 diferencias por archivo
IDMLLIB_MAX_DIFFS=0 go test ./pkg/idml/ -run TestGoldenRoundtrip_ExampleIDML -v -count=1 2>&1 | grep 'resumen \['

# Medir contra un IDML propio (la ruta debe ser ABSOLUTA)
IDMLLIB_MAX_DIFFS=0 IDMLLIB_IMAGES_FIXTURE="/ruta/absoluta/a/tu.idml" \
  go test ./pkg/idml/ -run 'TestGoldenRoundtrip_ExampleIDML/archivo_evidencia_imagenes' -v -count=1
```

Las diferencias se registran, no hacen fallar el test: hoy son la medición que hay que
llevar a cero. Detalle completo, variables de entorno y línea base en
**[`docs/FIDELIDAD.md`](docs/FIDELIDAD.md)**.

#### Patrones de limpieza en tests

Este proyecto sigue patrones estrictos de limpieza en tests para mantener el repositorio limpio:

**Limpieza automática**:
```go
func TestExample(t *testing.T) {
    // Use t.TempDir() for automatic cleanup
    tempDir := t.TempDir()
    outputPath := filepath.Join(tempDir, "output.idml")
    
    // Test logic here...
    // No manual cleanup needed - t.TempDir() handles it
}
```

**Modo debug con limpieza**:
```go
func TestExampleWithDebug(t *testing.T) {
    var outputPath string
    
    if *preserveTestOutput {
        outputPath = "debug_output.idml"
        t.Cleanup(func() {
            if !t.Failed() {
                os.Remove(outputPath)
            }
        })
    } else {
        tempDir := t.TempDir()
        outputPath = filepath.Join(tempDir, "output.idml")
    }
    
    // Test logic...
}
```

**Guías para artefactos de test**:
- Nunca commitear archivos de salida de tests al repositorio
- Usar `t.TempDir()` para archivos temporales que deben limpiarse automáticamente
- Colocar datos de test persistentes en directorios `testdata/`
- Usar flags de debug con moderación y asegurar la limpieza después de depurar

### Cobertura de tests

- General: ~74%
- pkg/analysis: 93.6%
- pkg/idml: 61.1%
- pkg/idms: 70.6%
- internal/xmlutil: 79.5%

### Estructura del proyecto

```
idmllib/
├── pkg/
│   ├── idml/          # Main API and coordinator
│   ├── common/        # Shared types
│   ├── document/      # Document structure
│   ├── spread/        # Page layouts
│   ├── story/         # Text content
│   ├── resources/     # Styles, fonts, graphics
│   ├── analysis/      # Dependency tracking
│   └── idms/          # IDMS export
├── internal/
│   ├── xmlutil/       # XML utilities
│   └── testutil/      # Test helpers
├── cmd/
│   ├── cli/           # Interactive CLI tool
│   └── debug-export/  # Debug utilities
├── docs/              # Documentation
└── testdata/          # Test IDML files
```

## Arquitectura

Esta librería usa una arquitectura domain-driven donde los paquetes reflejan la estructura de archivos IDML:

- **pkg/common**: Tipos compartidos entre dominios
- **pkg/document**: Tipos de designmap.xml
- **pkg/spread**: Tipos de Spreads/*.xml (page items, layouts)
- **pkg/story**: Tipos de Stories/*.xml (contenido de texto)
- **pkg/resources**: Tipos de Resources/*.xml (styles, fonts, graphics)
- **pkg/idml**: Coordinador principal y API pública

Ver [ARCHITECTURE.md](ARCHITECTURE.md) para documentación detallada.

## Recursos

- [Adobe IDML Specification](https://www.adobe.com/devnet/indesign/sdk.html)
- [Documentación del proyecto](claude.md)
- [Documentación de arquitectura](ARCHITECTURE.md)

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.

## Contribuciones

¡Las contribuciones son bienvenidas! Ver [CONTRIBUTING.md](CONTRIBUTING.md) para más detalles.

## Autor

Fredrik Gustafsson ([@dimelords](https://github.com/dimelords))

## Changelog

### v2.0.0 (2025-01-13)
- ✅ **BREAKING**: Ruta del módulo actualizada a `github.com/dimelords/idmllib/v2` siguiendo la semántica de versionado de módulos Go
- ✅ Soporte completo de lectura/escritura IDML con fidelidad de roundtrip
- ✅ Arquitectura de paquetes domain-driven que refleja la estructura de archivos IDML
- ✅ Compatibilidad con Go 1.23 con correcciones de seguridad exhaustivas
- ✅ CLI tool con interfaz TUI interactiva
- ✅ ResourceManager para seguimiento, validación y limpieza de recursos
- ✅ Selection API para seleccionar elementos por ID de forma programática
- ✅ Funcionalidad de exportación de IDMS snippets
- ✅ Sin dependencias externas (solo Go stdlib)
- ✅ 74% de cobertura de tests con tests de roundtrip exhaustivos
- ✅ Separación limpia de responsabilidades sin dependencias circulares
- ✅ Pipeline de linting y CI/CD exhaustivo con escaneo de seguridad

### v0.1.0
- ✅ Release inicial con lectura/escritura IDML básica
- ✅ Parseo de Document, Story y Spread
- ✅ Capacidad completa de roundtrip
