# IDML Build - Documentación de Arquitectura

**Proyecto:** idmlbuild  
**Lenguaje:** Go 1.23+  
**Propósito:** Biblioteca de Go de nivel productivo para parsear y manipular archivos IDML (InDesign Markup Language)

---

## Tabla de Contenidos

1. [Descripción General](#descripción-general)
2. [Arquitectura Central](#arquitectura-central)
3. [Estructura de Paquetes](#estructura-de-paquetes)
4. [Modelos de Datos](#modelos-de-datos)
5. [Patrones de Diseño](#patrones-de-diseño)
6. [Estrategia de Pruebas](#estrategia-de-pruebas)
7. [Guías de Desarrollo](#guías-de-desarrollo)

---

## Descripción General

### ¿Qué es IDML?

IDML (InDesign Markup Language) es el formato de archivo basado en XML de Adobe InDesign. Esencialmente es un archivo ZIP que contiene:
- **designmap.xml** - Manifiesto y estructura del documento
- **Stories/** - Archivos XML de contenido de texto
- **Spreads/** - Archivos XML de maquetación de páginas
- **Resources/** - Definiciones de estilos, fuentes y gráficos
- **META-INF/** - Metadatos del paquete

### Objetivos del Proyecto

1. **Parsear** archivos IDML con 100% de fidelidad
2. **Manipular** la estructura del documento de forma programática
3. **Generar** archivos IDML válidos desde structs de Go
4. **Preservar** elementos desconocidos (compatibilidad hacia adelante)
5. Código **listo para producción** con pruebas exhaustivas

---

## Arquitectura Central

### Diseño en Capas

```
┌─────────────────────────────────────┐
│   APIs de Alto Nivel (Futuro)       │
│   - Manipulación de documentos      │
│   - Gestión de estilos              │
│   - Creación de contenido           │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   pkg/idml (Foco Actual)            │
│   - Struct Document (designmap.xml) │
│   - Funciones Parse/Marshal         │
│   - Acceso type-safe a elementos    │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   internal/xmlutil                  │
│   - Utilidades XML                  │
│   - Manejo de namespaces            │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   Biblioteca Estándar de Go         │
│   - encoding/xml                    │
│   - archive/zip                     │
└─────────────────────────────────────┘
```

### Principios Clave

1. **Separación de Responsabilidades**
   - Parseo de documentos (pkg/idml)
   - Manejo de ZIP (read.go, write.go)
   - Utilidades XML (internal/xmlutil)

2. **Compatibilidad hacia Adelante**
   - Elementos desconocidos preservados via `RawXMLElement`
   - Sin pérdida de datos durante roundtrips
   - Soporte para versiones futuras de InDesign

3. **Seguridad de Tipos**
   - Structs explícitos para los elementos principales
   - Verificaciones en tiempo de compilación
   - Soporte de autocompletado en el IDE

4. **Cero Dependencias**
   - Solo la biblioteca estándar de Go
   - Sin parsers externos de XML/ZIP
   - Mínima carga de mantenimiento

---

## Estructura de Paquetes

```
idmlbuild/
├── pkg/
│   ├── idml/              # Coordinador central + API pública
│   │   ├── package.go     # Coordinador del paquete (I/O de archivos, caché)
│   │   ├── read.go        # Lectura de ZIP
│   │   ├── write.go       # Escritura de ZIP
│   │   ├── selection.go   # API de selección para exportación IDMS
│   │   ├── resourcemgr*.go  # Gestión de recursos
│   │   ├── errors.go      # Tipos de error
│   │   └── *_test.go      # Pruebas
│   │
│   ├── common/            # Tipos compartidos entre todos los dominios
│   │   └── types.go       # RawXMLElement, Properties, PathGeometry
│   │
│   ├── document/          # Tipos de documento (designmap.xml)
│   │   ├── document.go    # Document + 30 tipos (Language, Layer, etc.)
│   │   ├── metadata.go    # ProcessingInstruction, DocumentWithMetadata
│   │   ├── parse.go       # ParseDocument*(), MarshalDocument*()
│   │   └── designmap.go   # Tipos Designmap heredados (deprecados)
│   │
│   ├── spread/            # Tipos de Spread (Spreads/*.xml)
│   │   ├── spread.go      # Spread, SpreadElement, Page
│   │   ├── pageitems.go   # TextFrame, Rectangle, Oval, Image, etc.
│   │   ├── graphicline.go # GraphicLine
│   │   └── parse.go       # ParseSpread(), MarshalSpread()
│   │
│   ├── story/             # Tipos de Story (Stories/*.xml)
│   │   ├── story.go       # Story, StoryElement, ParagraphStyleRange
│   │   └── parse.go       # ParseStory(), MarshalStory()
│   │
│   ├── resources/         # Tipos de recursos (Resources/*.xml)
│   │   ├── graphics.go    # GraphicFile, Color, Swatch, Gradient, etc.
│   │   ├── fonts.go       # FontsFile, FontFamily, Font, etc.
│   │   ├── styles.go      # StylesFile, CharacterStyle, ParagraphStyle
│   │   ├── parse_*.go     # Funciones de parseo para cada tipo de recurso
│   │   └── errors.go      # Errores específicos de recursos
│   │
│   ├── analysis/          # Seguimiento de dependencias para exportación IDMS
│   │   └── tracker.go     # DependencyTracker, DependencySet
│   │
│   └── idms/              # Funcionalidad de exportación de snippets IDMS
│       ├── package.go     # Paquete IDMS (archivo XML único)
│       ├── exporter.go    # Exporter, selección → IDMS
│       └── *.go           # I/O y marshaling específicos de IDMS
│
├── internal/
│   ├── xmlutil/           # Utilidades XML
│   └── testutil/          # Helpers de prueba
│
├── cmd/cli/               # Herramienta CLI interactiva con TUI Bubbletea
│
├── testdata/              # Fixtures de prueba
│   ├── *.idml             # Archivos IDML de muestra
│   ├── *.xml              # Archivos XML de muestra
│   └── golden/            # Salidas esperadas
│
└── docs/
    ├── ARCHITECTURE.md    # Este archivo
    └── EPIC*.md           # Documentación de épicas
```

**Arquitectura tras la Refactorización del Épico 5:**

La estructura de paquetes sigue un diseño orientado al dominio que refleja la estructura de archivos IDML:
- **pkg/common**: Tipos compartidos usados en todos los dominios
- **pkg/document**: Tipos y parseo de designmap.xml
- **pkg/spread**: Tipos XML de Spread y elementos de página
- **pkg/story**: Tipos XML de Story y contenido de texto
- **pkg/resources**: Tipos XML de recursos (estilos, fuentes, gráficos)
- **pkg/idml**: Paquete coordinador que orquesta todos los dominios
- **pkg/analysis**: Análisis de dependencias entre dominios
- **pkg/idms**: Exportación IDMS usando tipos de todos los paquetes

**Diseño de Paquetes:**
- Cada paquete provee tipos específicos del dominio y funciones de parseo
- pkg/idml coordina todos los paquetes y provee la API pública principal
- Importaciones directas desde paquetes de dominio (document, story, spread, resources)
- Cero dependencias circulares mediante el patrón del paquete common

### Responsabilidades de Archivos

**document.go** (2.463 líneas)
- Definición del struct Document (50+ atributos, 14+ tipos de elementos)
- 35+ structs de soporte (Language, Layer, Section, etc.)
- Todas las definiciones de tipos de elementos
- Modelo de datos central

**document_parse.go**
- `ParseDocument(data []byte) (*Document, error)`
- `MarshalDocument(doc *Document) ([]byte, error)`
- Manejo de namespaces
- Parseo/generación de XML

**read.go**
- `OpenIDML(path string) (*Package, error)`
- Lectura de archivos ZIP
- Extracción de archivos
- Manejo de tipos MIME

**write.go**
- `WriteIDML(pkg *Package, path string) error`
- Escritura de archivos ZIP
- Compresión
- Orden de archivos (mimetype primero, sin comprimir)

**errors.go**
- Tipos de error personalizados
- Envoltura de errores
- Errores con contexto

---

## Modelos de Datos

### Struct Central: Document

El struct `Document` representa designmap.xml, el archivo de manifiesto en un paquete IDML.

```go
type Document struct {
    // 50+ atributos (identidad, maquetación, colores, etc.)
    XMLName    xml.Name
    Xmlns      string
    DOMVersion string
    Self       string
    // ... muchos más ...

    // Elementos hijo explícitos (14 categorías, 35+ tipos)
    Properties      *Properties
    Languages       []Language
    GraphicResource *ResourceRef
    // ... todos los elementos principales ...

    // Catch-all para elementos desconocidos
    OtherElements []RawXMLElement
}
```

### Categorías de Elementos

**Implementados (95%+ de documentos típicos):**

1. **Metadatos** - Properties, Labels, KeyValuePairs
2. **Idiomas** - Configuraciones de localización
3. **Recursos** - Graphic, Fonts, Styles, Preferences, Tags
4. **Contenido** - MasterSpreads, Spreads, Stories, BackingStory
5. **Maquetación** - Layers, Sections
6. **Tipografía** - NumberingList, NamedGrid, GridDataInformation
7. **Usuarios** - DocumentUser (colaboración)
8. **Colores** - ColorGroup, ColorGroupSwatch
9. **Viñetas** - Definiciones ABullet
10. **Flujo de trabajo** - Assignment (InCopy)
11. **Variables** - TextVariable (10 tipos)

**Aún no implementados (elementos poco frecuentes):**

- KinsokuTable, MojikumiTable (tipografía CJK)
- CrossReferenceFormat
- ConditionalTextPreference
- IndexingSortOption
- Varias preferencias de exportación

Todos los elementos no implementados se preservan via `OtherElements []RawXMLElement`.

---

## Patrones de Diseño

### 1. Patrón RawXMLElement

**Problema:** Necesidad de preservar elementos XML desconocidos sin structs explícitos.

**Solución:**
```go
type RawXMLElement struct {
    XMLName xml.Name
    Attrs   []xml.Attr `xml:",any,attr"`
    Content []byte     `xml:",innerxml"`
}

type Document struct {
    // Elementos explícitos
    Languages []Language
    
    // Catch-all
    OtherElements []RawXMLElement `xml:",any"`
}
```

**Beneficios:**
- Sin pérdida de datos durante roundtrips
- Compatible hacia adelante con versiones futuras de InDesign
- Fácil migración de elementos del catch-all a explícitos

### 2. Patrón de Referencia a Recursos

**Problema:** IDML usa referencias con namespace a archivos XML externos.

**Solución:**
```go
type ResourceRef struct {
    XMLName xml.Name
    Src     string `xml:"src,attr"`
}

// En Document:
GraphicResource *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`
```

**Beneficios:**
- Referencias type-safe
- Manejo consistente de todos los tipos de recursos
- Fácil adición de nuevos tipos de recursos

### 3. Patrón de Preferencias

**Problema:** Los TextVariables tienen diferentes tipos de preferencia según el tipo de variable.

**Solución:**
```go
type TextVariable struct {
    VariableType string
    
    // Solo uno de estos estará poblado
    DatePreference       *DateVariablePreference
    FileNamePreference   *FileNameVariablePreference
    PageNumberPreference *PageNumberVariablePreference
    // ...
    
    OtherElements []RawXMLElement `xml:",any"`
}
```

**Beneficios:**
- Acceso type-safe a preferencias
- El compilador detecta combinaciones inválidas
- Extensible para nuevos tipos de preferencia

### 4. Patrón Contenedor de Properties

**Problema:** Muchos elementos tienen un hijo Properties con contenido variado.

**Solución:**
```go
type Properties struct {
    XMLName       xml.Name `xml:"Properties"`
    Label         *Label
    OtherElements []RawXMLElement `xml:",any"`
}

// Usado por: Document, Layer, Section, Assignment, ABullet, etc.
```

**Beneficios:**
- Patrón consistente en todos los elementos
- Maneja propiedades conocidas (Label) + desconocidas
- Fácil de extender

---

## Estrategia de Pruebas

### Cobertura de Pruebas

**30+ funciones de prueba que cubren:**

1. **Pruebas de Parseo** - Verifican la población de structs
2. **Pruebas de Roundtrip** - Igualdad Parse → Marshal → Parse
3. **Pruebas de Golden Files** - Preservación byte a byte del ZIP
4. **Pruebas de Estructura XML** - Comparación del árbol XML
5. **Pruebas por Elemento** - Validación de cada tipo de elemento

### Datos de Prueba

**testdata/ contiene:**
- `plain.idml` - IDML válido mínimo
- `example.idml` - IDML complejo del mundo real
- `designmap.xml` - XML de manifiesto completo
- `designmap_minimal.xml` - Manifiesto mínimo
- Archivos XML individuales para pruebas específicas

### Ejecutar Pruebas

```bash
# Todas las pruebas
go test ./pkg/idml/

# Salida detallada
go test -v ./pkg/idml/

# Prueba específica
go test -v -run TestDocumentRoundtrip ./pkg/idml/

# Cobertura
go test -cover ./pkg/idml/
```

### Pruebas con Golden Files

Los golden files garantizan roundtrips byte a byte:

```go
func TestGoldenRoundtrip(t *testing.T) {
    // Leer IDML original
    original := readIDML("plain.idml")
    
    // Parsear y reescribir
    pkg, _ := idml.OpenIDML("plain.idml")
    idml.WriteIDML(pkg, "output.idml")
    
    // Comparar byte a byte
    assertIdentical(t, original, output)
}
```

---

## Guías de Desarrollo

### Agregar Nuevos Elementos

**1. Investigar la estructura del elemento:**
```bash
grep -A 10 "ElementName" testdata/*.xml
```

**2. Definir el struct:**
```go
type NewElement struct {
    XMLName       xml.Name `xml:"NewElement"`
    Self          string   `xml:"Self,attr"`
    Name          string   `xml:"Name,attr"`
    // ... atributos ...
    OtherElements []RawXMLElement `xml:",any"`
}
```

**3. Agregar al Document:**
```go
type Document struct {
    // ... campos existentes ...
    
    // Paso N: Nuevos Elementos
    NewElements []NewElement `xml:"NewElement,omitempty"`
    
    OtherElements []RawXMLElement `xml:",any"`
}
```

**4. Actualizar el comentario del catch-all:**
Eliminar el elemento de la documentación de OtherElements.

**5. Escribir pruebas:**
```go
func TestDocumentNewElements(t *testing.T) { /* ... */ }
func TestDocumentNewElementsRoundtrip(t *testing.T) { /* ... */ }
```

**6. Ejecutar pruebas:**
```bash
go test ./pkg/idml/
```

### Estilo de Código

**Struct Tags:**
```go
// Atributos
Self string `xml:"Self,attr"`
Name string `xml:"Name,attr,omitempty"` // Opcional

// Elementos hijo
Languages []Language `xml:"Language,omitempty"` // Múltiples
Properties *Properties `xml:"Properties,omitempty"` // Único opcional

// Elementos con namespace
Graphic *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`

// Catch-all
OtherElements []RawXMLElement `xml:",any"`
```

**Documentación:**
```go
// MyElement representa un elemento de documento para XYZ.
// Este elemento controla ABC y se usa para DEF.
type MyElement struct {
    // Identificación
    Self string `xml:"Self,attr"` // Identificador único (ej. "abc123")
    Name string `xml:"Name,attr"` // Nombre para mostrar
    
    // Configuración
    Enabled string `xml:"Enabled,attr,omitempty"` // Habilitar función ("true"/"false")
    
    // Elementos hijo
    Properties *Properties `xml:"Properties,omitempty"` // Propiedades opcionales
    
    // Catch-all
    OtherElements []RawXMLElement `xml:",any"`
}
```

**Convenciones de Nombres:**
- Structs: PascalCase (ej. `DocumentUser`)
- Campos: PascalCase (ej. `PageNumberStart`)
- Funciones: PascalCase para públicas, camelCase para privadas
- Funciones de prueba: `TestSujeto_Escenario` o `TestSujeto`

### Manejo de Errores

```go
// Envolver errores con contexto
if err := doSomething(); err != nil {
    return nil, &Error{
        Op:  "nombre de la operación",
        Err: err,
    }
}

// Tipo de error personalizado
type Error struct {
    Op  string // Operación que falló
    Err error  // Error subyacente
}
```

### Consideraciones de Rendimiento

1. **Parseo diferido** - Solo parsear lo necesario
2. **Streaming** - Usar io.Reader/Writer donde sea posible
3. **Eficiencia de memoria** - Evitar almacenamiento duplicado de datos
4. **Mínimas asignaciones** - Reutilizar buffers cuando sea apropiado

---

## Hoja de Ruta Futura

### Fase 3: Parseo de Contenido (Próximo)

**Stories:**
- Parsear archivos XML de Story
- Contenido de texto y formato
- Estilos de párrafo/carácter

**Spreads:**
- Parsear archivos XML de Spread
- Maquetación de página y marcos
- Elementos posicionados

**Estilos:**
- Parsear Styles.xml
- Estilos de carácter
- Estilos de párrafo
- Estilos de objeto

### Fase 4: APIs de Alto Nivel

**Manipulación de documentos:**
```go
// Crear nuevo documento
doc := idml.NewDocument()

// Agregar story
story := doc.AddStory("Mi Story")
story.AddParagraph("¡Hola, mundo!")

// Agregar página
spread := doc.AddSpread()
page := spread.AddPage()

// Guardar
doc.SaveAs("output.idml")
```

**Extracción de contenido:**
```go
// Obtener todo el texto
text := doc.ExtractAllText()

// Buscar contenido específico
results := doc.Search("palabra clave")
```

**Gestión de estilos:**
```go
// Listar estilos
styles := doc.ListParagraphStyles()

// Aplicar estilo
paragraph.ApplyStyle("Título 1")

// Crear estilo
doc.CreateParagraphStyle("Personalizado", styleOpts)
```

---

## Métricas de Rendimiento

**Actual (Fase 2):**
- Parsear designmap.xml: ~0,5ms (archivo de 9KB)
- Apertura completa de IDML: ~20ms (incluye extracción ZIP)
- Memoria: ~2MB para un documento típico
- Suite de pruebas: <0,4s para 30+ pruebas

**Objetivos (Fase 3):**
- Parseo completo del documento: <100ms
- Memoria: <10MB para un documento típico
- Soporte de streaming para archivos grandes

---

## Estabilidad de la API

**Estado Actual:** Alpha

- La API de pkg/idml es relativamente estable
- El struct Document puede evolucionar
- Posibles cambios que rompan compatibilidad hasta v1.0

**Versionado:**
- Seguir versionado semántico
- Incremento de versión mayor para cambios que rompan compatibilidad
- Versión menor para nuevas funcionalidades
- Versión de parche para corrección de errores

---

## Contribuir

### Configuración

```bash
# Clonar
git clone https://github.com/dimelords/idmlbuild
cd idmlbuild

# Probar
go test ./...

# Lint (si golangci-lint está instalado)
golangci-lint run
```

### Proceso de Pull Request

1. Escribir pruebas primero
2. Implementar la funcionalidad
3. Ejecutar la suite de pruebas completa
4. Actualizar la documentación
5. Enviar PR con descripción clara

### Lista de Verificación para Revisión de Código

- [ ] Las pruebas pasan
- [ ] Documentación actualizada
- [ ] Sin cambios que rompan compatibilidad (o justificados)
- [ ] El código sigue la guía de estilo
- [ ] Manejo de errores incluido
- [ ] Rendimiento considerado

---

## Referencias

### Especificación IDML

- [Adobe InDesign Interchange (INX) & Markup (IDML)](https://www.adobe.com/devnet/indesign/sdk.html)
- [IDML Cookbook](https://wwwimages2.adobe.com/content/dam/acom/en/devnet/indesign/sdk/cs6/idml/idml-cookbook.pdf)
- [Especificación IDML](https://wwwimages2.adobe.com/content/dam/acom/en/devnet/indesign/sdk/cs6/idml/idml-specification.pdf)

### Recursos de Go

- [Paquete encoding/xml](https://pkg.go.dev/encoding/xml)
- [Paquete archive/zip](https://pkg.go.dev/archive/zip)
- [Testing en Go](https://go.dev/doc/tutorial/add-a-test)

---

## Licencia

[Tu Licencia Aquí]

---

**Última Actualización:** 2025-11-24  
**Versión:** Épico 5 Completo - Arquitectura Orientada al Dominio  
**Estado:** Listo para producción para operaciones completas IDML/IDMS con estructura de paquetes limpia
