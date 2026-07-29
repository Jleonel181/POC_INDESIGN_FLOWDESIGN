# Templates IDML

Este directorio contiene archivos de template para crear documentos IDML desde cero.

## Propósito

Estos templates proveen estructuras XML mínimas válidas que InDesign acepta. Sirven como punto de partida para generar documentos IDML de forma programática sin necesitar un documento existente para modificar.

## Estructura

### `minimal/`

Contiene los archivos mínimos requeridos para un documento IDML válido:

- **designmap.xml** - Estructura mínima del documento con una página
- **Preferences.xml** - Preferencias mínimas con valores por defecto razonables

Estos templates se embeben en el binario de Go en tiempo de compilación usando directivas `go:embed`, haciéndolos disponibles sin dependencias de archivos externos.

## Uso

### Desde código Go

```go
// Crear un nuevo documento IDML desde templates
pkg, err := idml.NewFromTemplate(nil) // usa valores por defecto

// O con opciones personalizadas
pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
    DOMVersion:          "20.4",
    UseMinimalTemplates: true,
})

// Modificar el documento según sea necesario
doc, _ := pkg.Document()
// ... hacer cambios ...

// Escribir a archivo
err = idml.Write(pkg, "output.idml")
```

## Archivos de template explicados

### Preferences.xml

Contiene configuraciones por defecto esenciales:

- **Preferencias de impresión** - Configuración básica de impresión (orientación, tamaño de papel, etc.)
- **Valores por defecto de elementos de página** - Trazo, relleno y esquinas por defecto para marcos
- **Preferencias de texto** - Configuración tipográfica, manejo de comillas
- **Valores por defecto de texto** - Fuente, tamaño, alineación y espaciado por defecto
- **Ajuste de marcos** - Cómo el contenido se ajusta dentro de los marcos
- **Preferencias de story** - Dirección y orientación del flujo de texto

La versión mínima incluye solo las configuraciones más esenciales. InDesign aplicará sus propios valores por defecto para cualquier preferencia faltante.

### designmap.xml

Define la estructura del documento:

- **Propiedades del documento** - Tamaño de página, orientación, márgenes
- **Referencia a spreads** - Enlaces a archivos de spread (mínimo: un spread)
- **Master spreads** - Plantillas de página (mínimo: master vacío)
- **Idiomas** - Configuración de idioma del texto
- **Preferencias de vista** - Unidades de medida

## Extender los templates

Para agregar más variaciones de templates:

1. Crear un nuevo directorio (ej. `templates/standard/`)
2. Agregar archivos de template al directorio
3. Actualizar `templates.go` con directivas `//go:embed`
4. Agregar opciones a `TemplateOptions` para seleccionar templates

Ejemplo:

```go
//go:embed templates/standard/Preferences.xml
var standardPreferences []byte

//go:embed templates/standard/Styles.xml
var standardStyles []byte
```

## Validación

Todos los archivos de template deben validarse:

1. Creando un IDML usando los templates
2. Abriéndolo en Adobe InDesign
3. Verificando que no aparezcan errores ni advertencias

Si InDesign rechaza un template, probablemente significa:
- Elementos requeridos faltantes
- Valores de atributos inválidos
- Declaraciones de namespace incorrectas
- Recursos referenciados faltantes

## Buenas prácticas

1. **Mantener los templates mínimos** - Incluir solo lo necesario
2. **Usar valores por defecto válidos** - Todos los valores de atributos deben coincidir con las expectativas de InDesign
3. **Probar en InDesign** - Siempre verificar que los templates abran correctamente
4. **Documentar dependencias** - Indicar qué templates requieren otros recursos
5. **Compatibilidad de versiones** - Indicar qué versiones de InDesign soportan cada template

## Referencias

- Adobe InDesign IDML Cookbook (CS5/CS6)
- Documentación del SDK de InDesign
- Especificación del formato IDML: https://www.adobe.com/devnet/indesign/
