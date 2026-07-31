# Templates IDML

Este directorio contiene archivos de template para crear documentos IDML desde cero.

## Propósito

Estos templates proveen estructuras XML mínimas válidas que InDesign acepta. Sirven como punto de partida para generar documentos IDML de forma programática sin necesitar un documento existente para modificar.

## Estructura

### `minimal/`

Contiene los **13 archivos** de un documento IDML mínimo: una página con un marco de
texto. Es el mismo conjunto de entradas que `testdata/plain.idml`, que es una
exportación de InDesign de esa misma forma y sirve de oráculo.

| Archivo | Contenido | Lleva valores calculados |
|---|---|---|
| `mimetype` | Identificador del formato. Primera entrada del ZIP y sin comprimir | no |
| `designmap.xml` | Manifiesto: capa, sección, y las referencias a los demás archivos | sí |
| `container.xml` | `META-INF/container.xml`, señala cuál es el archivo raíz | no |
| `metadata.xml` | `META-INF/metadata.xml`, paquete XMP | sí |
| `Graphic.xml` | Colores y muestras | no |
| `Fonts.xml` | Fuentes | no |
| `Styles.xml` | Estilos | no |
| `Preferences.xml` | Preferencias, incluido `DocumentPreference` con el tamaño de página | sí |
| `Tags.xml` | Etiquetas XML | no |
| `MasterSpread_ub4.xml` | Página maestra | sí |
| `Spread_ud3.xml` | El spread con su página y su marco de texto | sí |
| `BackingStory.xml` | Story de respaldo de la estructura XML | sí |
| `Story_ue1.xml` | El texto del marco | sí |

Los que llevan valores calculados son plantillas de `text/template`; el resto se copia
tal cual. Todos se embeben en el binario en tiempo de compilación con `go:embed`.

> **Si añades o quitas un archivo de `minimal/`, actualiza `templates.go`.** Un
> `//go:embed` que apunte a un archivo inexistente **rompe la compilación**, no solo
> los tests.

### Identificadores y cierre referencial

Los identificadores están **fijos** en las plantillas, y el documento solo es válido si
coinciden entre archivos. Estas cinco referencias tienen que cerrar:

| Referencia | Apunta a | Valor |
|---|---|---|
| `TextFrame@ItemLayer` | la `Layer` del designmap | `uba` |
| `TextFrame@ParentStory` | la story emitida | `ue1` |
| `Section@PageStart` | la página del spread | `ud8` |
| `Page@AppliedMaster` | el master spread | `ub4` |
| `Document@StoryList` | la story y la backing story | `ue1 u98` |

`TestNewFromTemplate_CierreReferencial` resuelve cada una contra el elemento al que
apunta, así que cambiar un identificador en una plantilla y olvidarlo en otra deja el
test en rojo.

### Geometría

No está fija: se calcula de `TemplateOptions`. El marco de texto se coloca en la caja de
márgenes, y `ColumnsPositions` se deriva del ancho útil, el número de columnas y el
medianil.

`NewFromTemplate()` devuelve error si los márgenes no dejan área utilizable o si el
número de columnas no cabe en el ancho disponible.

## Uso

### Desde código Go

```go
// Crear un nuevo documento IDML desde templates
pkg, err := idml.NewFromTemplate(nil) // usa valores por defecto: US Letter, 1 columna

// O con opciones personalizadas
pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
    DOMVersion:   "20.4",
    Preset:       idml.PresetA4,
    Orientation:  "Portrait",
    ColumnCount:  5,
    ColumnGutter: 12,
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

### Resultado registrado

**Verificado en Adobe InDesign.** Dos sondas A4, una de 1 columna y otra de 5, abren
**sin aviso de daño y sin petición de recuperación**, y muestran la página con su marco
de texto dentro de los márgenes. Abren también en **Affinity Publisher**, así que la
forma emitida no depende de una tolerancia particular de InDesign.

De esa verificación salieron dos conclusiones útiles:

- **El subconjunto curado de `Preferences.xml` es suficiente.** No hizo falta copiar el archivo completo de un documento real. Confirma el principio de esta plantilla: el subconjunto necesita ser **válido, no completo**, porque InDesign aplica sus propios valores por defecto a lo que falte.
- **InDesign respeta el `ColumnsPositions` emitido y no lo recalcula.** Con 5 columnas y medianil 12 se emiten huecos de 12 y se miden 12 al abrir. El cálculo directo es la derivación correcta.

### Regenerar las sondas

```bash
IDMLLIB_PROBE_DIR=~/Desktop/sondas_idml go test ./pkg/idml/ -run SondaParaInDesign -v -count=1
```

Sin la variable, el test se omite: emite archivos para inspección humana, no comprueba
nada por sí mismo.

### Si un cambio rompe la validación

Al modificar una plantilla hay que repetir la verificación manual. Si InDesign la
rechaza, las causas probables son:

- Elementos requeridos faltantes
- Valores de atributos inválidos
- Declaraciones de namespace incorrectas
- Recursos referenciados faltantes, o un identificador que dejó de cerrar

Para el último caso hay un atajo que no necesita InDesign:

```bash
go test ./pkg/idml/ -run TestNewFromTemplate -v
```

Comprueba el conjunto de rutas contra `testdata/plain.idml`, las tres referencias
`idPkg:` del designmap, el cierre referencial de las cinco referencias cruzadas, y que
los 13 XML son bien formados.

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
