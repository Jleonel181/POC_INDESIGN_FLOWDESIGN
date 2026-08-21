package idml

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// cachedTemplate parsea una plantilla embebida una sola vez y guarda el resultado,
// incluido el error, para las llamadas siguientes. Sustituye a los pares de
// variables sueltas que había por plantilla: ahora son seis y el patrón repetido
// costaba cuatro declaraciones cada uno.
type cachedTemplate struct {
	name   string
	source []byte

	once sync.Once
	tmpl *template.Template
	err  error
}

func (c *cachedTemplate) get() (*template.Template, error) {
	c.once.Do(func() {
		c.tmpl, c.err = template.New(c.name).Parse(string(c.source))
	})
	return c.tmpl, c.err
}

// Plantillas de la forma mínima que llevan valores calculados. Las que no llevan
// ninguno (mimetype, container.xml, Graphic.xml, Fonts.xml, Styles.xml, Tags.xml) se
// copian tal cual y no pasan por aquí.
var (
	designmapTmpl    = &cachedTemplate{name: "designmap", source: minimalDesignMap}
	masterspreadTmpl = &cachedTemplate{name: "masterspread", source: minimalMasterSpread}
	spreadTmpl       = &cachedTemplate{name: "spread", source: minimalSpread}
	storyTmpl        = &cachedTemplate{name: "story", source: minimalStory}
	backingStoryTmpl = &cachedTemplate{name: "backingstory", source: minimalBackingStory}
	preferencesTmpl  = &cachedTemplate{name: "preferences", source: minimalPreferences}
	metadataTmpl     = &cachedTemplate{name: "metadata", source: minimalMetadata}
)

// Archivos de template embebidos en tiempo de compilación.
// Proveen estructuras mínimas válidas para crear documentos IDML desde cero.

//go:embed templates/minimal/mimetype
var minimalMimetype []byte

//go:embed templates/minimal/container.xml
var minimalContainer []byte

//go:embed templates/minimal/designmap.xml
var minimalDesignMap []byte

//go:embed templates/minimal/Preferences.xml
var minimalPreferences []byte

//go:embed templates/minimal/MasterSpread_ub4.xml
var minimalMasterSpread []byte

//go:embed templates/minimal/Graphic.xml
var minimalGraphic []byte

//go:embed templates/minimal/Fonts.xml
var minimalFonts []byte

//go:embed templates/minimal/Styles.xml
var minimalStyles []byte

//go:embed templates/minimal/Tags.xml
var minimalTags []byte

//go:embed templates/minimal/Spread_ud3.xml
var minimalSpread []byte

//go:embed templates/minimal/Story_ue1.xml
var minimalStory []byte

//go:embed templates/minimal/BackingStory.xml
var minimalBackingStory []byte

//go:embed templates/minimal/metadata.xml
var minimalMetadata []byte

// DocumentPreset define tamaños de página y configuraciones estándar.
type DocumentPreset string

const (
	// PresetLetterUS crea un documento US Letter (8.5" × 11" / 612 × 792 pt)
	PresetLetterUS DocumentPreset = "letter-us"

	// PresetA4 crea un documento A4 (210 × 297 mm / 595.276 × 841.89 pt)
	PresetA4 DocumentPreset = "a4"

	// PresetTabloid crea un documento Tabloid (11" × 17" / 792 × 1224 pt)
	PresetTabloid DocumentPreset = "tabloid"

	// PresetLegalUS crea un documento Legal (8.5" × 14" / 612 × 1008 pt)
	PresetLegalUS DocumentPreset = "legal"

	// PresetCustom permite dimensiones personalizadas
	PresetCustom DocumentPreset = "custom"
)

// PageDimensions contiene el ancho y alto de página en puntos.
type PageDimensions struct {
	Width  float64 // Ancho en puntos (1 punto = 1/72 pulgada)
	Height float64 // Alto en puntos
}

// StandardPresets provee dimensiones de página comunes.
var StandardPresets = map[DocumentPreset]PageDimensions{
	PresetLetterUS: {Width: 612, Height: 792},        // 8.5" × 11"
	PresetA4:       {Width: 595.276, Height: 841.89}, // 210mm × 297mm
	PresetTabloid:  {Width: 792, Height: 1224},       // 11" × 17"
	PresetLegalUS:  {Width: 612, Height: 1008},       // 8.5" × 14"
}

// TemplateOptions configura la creación de documentos desde templates.
type TemplateOptions struct {
	// DOMVersion especifica la versión del DOM de InDesign (ej. "20.4")
	// Si está vacío, usa "20.4" por defecto
	DOMVersion string

	// Preset especifica un tamaño de documento estándar
	// Por defecto: PresetLetterUS
	Preset DocumentPreset

	// CustomDimensions permite tamaños de página personalizados cuando Preset es PresetCustom
	CustomDimensions *PageDimensions

	// Orientation determina la orientación de la página
	// "Portrait" (por defecto) o "Landscape"
	Orientation string

	// Margins en puntos (por defecto: 36 puntos = 0.5 pulgadas)
	Margins struct {
		Top    float64
		Bottom float64
		Left   float64
		Right  float64
	}

	// ColumnCount para columnas de texto (por defecto: 1)
	ColumnCount int

	// ColumnGutter espaciado entre columnas en puntos (por defecto: 12)
	ColumnGutter float64

	// Intent define la intención del documento: "PrintIntent" (por defecto) o
	// "WebIntent" para documentos digitales.
	Intent string

	// PageBinding define el sentido de encuadernación: "LeftToRight" (por defecto)
	// o "RightToLeft" para idiomas RTL.
	PageBinding string

	// FacingPages activa páginas enfrentadas (por defecto: false).
	FacingPages bool

	// MeasurementUnits define las unidades de medida para todas las reglas y diálogos.
	// Valor por defecto: "Points". Opciones comunes: "Points", "Picas", "Millimeters",
	// "Centimeters", "Inches".
	MeasurementUnits string
}

// DefaultTemplateOptions devuelve valores por defecto razonables para US Letter portrait.
func DefaultTemplateOptions() *TemplateOptions {
	opts := &TemplateOptions{
		DOMVersion:   "20.4",
		Preset:       PresetLetterUS,
		Orientation:  "Portrait",
		ColumnCount:  1,
		ColumnGutter: 12,
	}
	opts.Margins.Top = 36
	opts.Margins.Bottom = 36
	opts.Margins.Left = 36
	opts.Margins.Right = 36
	return opts
}

// GetDimensions devuelve las dimensiones de página según el preset y la orientación.
func (opts *TemplateOptions) GetDimensions() PageDimensions {
	var dims PageDimensions

	if opts.Preset == PresetCustom && opts.CustomDimensions != nil {
		dims = *opts.CustomDimensions
	} else {
		dims = StandardPresets[opts.Preset]
	}

	// Intercambiar dimensiones para landscape
	if opts.Orientation == "Landscape" {
		dims.Width, dims.Height = dims.Height, dims.Width
	}

	return dims
}

// NewFromTemplate crea un nuevo paquete IDML desde templates embebidos.
// Útil para crear documentos IDML desde cero de forma programática.
//
// Ejemplo:
//
//	// Crear US Letter portrait por defecto
//	pkg, err := idml.NewFromTemplate(nil)
//
//	// Crear A4 landscape
//	pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
//	    Preset:      idml.PresetA4,
//	    Orientation: "Landscape",
//	})
//
//	// Crear tamaño personalizado
//	pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
//	    Preset: idml.PresetCustom,
//	    CustomDimensions: &idml.PageDimensions{
//	        Width:  720,  // 10 pulgadas
//	        Height: 1080, // 15 pulgadas
//	    },
//	})
//
// El paquete creado tendrá:
//   - designmap.xml con la estructura del documento configurada
//   - MasterSpreads/MasterSpread_ub4.xml con la página maestra
//   - Resources/Preferences.xml con valores por defecto razonables
//   - Estructura de directorios requerida
func NewFromTemplate(opts *TemplateOptions) (*Package, error) {
	if opts == nil {
		opts = DefaultTemplateOptions()
	}

	// Asegurar valores por defecto
	if opts.DOMVersion == "" {
		opts.DOMVersion = "20.4"
	}
	if opts.Orientation == "" {
		opts.Orientation = "Portrait"
	}
	if opts.ColumnCount <= 0 {
		opts.ColumnCount = 1
	}
	if opts.ColumnGutter <= 0 {
		opts.ColumnGutter = 12
	}

	data, err := newTemplateData(opts, opts.GetDimensions())
	if err != nil {
		return nil, err
	}

	// El orden de esta lista es el orden de las entradas del ZIP, y mimetype tiene
	// que ir primero y sin comprimir. El resto sigue el orden de plain.idml, que es
	// una exportación de InDesign de esta misma forma: 1 página, 1 marco de texto.
	//
	// Las plantillas con valores calculados se renderizan; las demás se copian.
	files := []struct {
		path string
		tmpl *cachedTemplate
		raw  []byte
	}{
		{path: PathMimetype, raw: minimalMimetype},
		{path: PathDesignmap, tmpl: designmapTmpl},
		{path: PathContainer, raw: minimalContainer},
		{path: PathMetadata, tmpl: metadataTmpl},
		{path: PathGraphic, raw: minimalGraphic},
		{path: PathFonts, raw: minimalFonts},
		{path: PathStyles, raw: minimalStyles},
		{path: PathPreferences, tmpl: preferencesTmpl},
		{path: PathTags, raw: minimalTags},
		{path: PathMasterSpread, tmpl: masterspreadTmpl},
		{path: PathSpread, tmpl: spreadTmpl},
		{path: PathBackingStory, tmpl: backingStoryTmpl},
		{path: PathStory, tmpl: storyTmpl},
	}

	pkg := New()
	for _, f := range files {
		content := f.raw
		if f.tmpl != nil {
			tmpl, tmplErr := f.tmpl.get()
			content, err = renderTemplate(f.tmpl.name, tmpl, tmplErr, data)
			if err != nil {
				return nil, common.WrapErrorWithPath("idml", "create from template", f.path, err)
			}
		}
		if err := pkg.addFileFromTemplate(f.path, content); err != nil {
			return nil, common.WrapErrorWithPath("idml", "create from template", f.path, err)
		}
	}

	return pkg, nil
}

// templateData son los valores que reciben todas las plantillas de la forma
// mínima. Es un único struct compartido, y no uno por plantilla, porque el cierre
// referencial entre los archivos depende de que la geometría sea la misma en todos:
// el `PageStart` del `Section` apunta a la página del spread, y el marco de texto se
// coloca dentro de los márgenes que declaran las preferencias.
//
// Las medidas van como cadena ya formateada y no como float64 porque el formato de
// un número en una plantilla de texto depende de cómo lo imprima cada una, y aquí
// interesa que el mismo valor salga idéntico en los cuatro archivos que lo llevan.
type templateData struct {
	DOMVersion  string
	Orientation string

	PageWidthStr  string
	PageHeightStr string
	HalfHeightStr string

	ColumnCount      int
	ColumnGutterStr  string
	ColumnWidthStr   string
	ColumnsPositions string

	MarginTopStr    string
	MarginBottomStr string
	MarginLeftStr   string
	MarginRightStr  string

	FrameCenterXStr    string
	FrameCenterYStr    string
	FrameHalfWidthStr  string
	FrameHalfHeightStr string

	// DocumentPreference options
	Intent      string
	PageBinding string
	FacingPages string

	// ViewPreference: unidades de medida
	MeasurementUnits string

	Timestamp  string
	InstanceID string
	DocumentID string
}

// num formatea una medida en puntos como la escribe InDesign: sin notación
// exponencial y sin ceros de relleno.
//
// Redondea a 9 decimales antes de recortar, que es la precisión que usa InDesign en
// sus propias exportaciones. Sin ese redondeo, un ancho de columna calculado sale
// como "95.05519999999999" en lugar de "95.0552": es el mismo número, pero el ruido
// de la coma flotante no se parece a nada que InDesign produzca.
func num(v float64) string {
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 9, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// newTemplateData calcula la geometría de la forma mínima a partir de las opciones.
//
// El marco de texto se coloca exactamente en la caja de márgenes, que es lo que hace
// InDesign al crear un documento con marco de texto principal, y así el marco queda
// coherente con los márgenes que se emiten en las preferencias.
func newTemplateData(opts *TemplateOptions, dims PageDimensions) (*templateData, error) {
	usableWidth := dims.Width - opts.Margins.Left - opts.Margins.Right
	usableHeight := dims.Height - opts.Margins.Top - opts.Margins.Bottom
	if usableWidth <= 0 || usableHeight <= 0 {
		return nil, common.Errorf("idml", "create from template", "",
			"los márgenes no dejan área utilizable: página %s×%s, márgenes %s/%s/%s/%s",
			num(dims.Width), num(dims.Height),
			num(opts.Margins.Top), num(opts.Margins.Bottom), num(opts.Margins.Left), num(opts.Margins.Right))
	}

	// Ancho de una columna y posiciones de las columnas dentro del área utilizable.
	// InDesign las emite como pares inicio/fin relativos al margen izquierdo.
	columnWidth := (usableWidth - float64(opts.ColumnCount-1)*opts.ColumnGutter) / float64(opts.ColumnCount)
	if columnWidth <= 0 {
		return nil, common.Errorf("idml", "create from template", "",
			"%d columnas con medianil %s no caben en %s puntos de ancho utilizable",
			opts.ColumnCount, num(opts.ColumnGutter), num(usableWidth))
	}

	positions := make([]string, 0, opts.ColumnCount*2)
	for i := 0; i < opts.ColumnCount; i++ {
		start := float64(i) * (columnWidth + opts.ColumnGutter)
		positions = append(positions, num(start), num(start+columnWidth))
	}

	instanceID, err := newUUID()
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMetadata, err)
	}
	documentID, err := newUUID()
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMetadata, err)
	}

	// Opciones de DocumentPreference
	intent := opts.Intent
	if intent == "" {
		intent = "PrintIntent"
	}
	pageBinding := opts.PageBinding
	if pageBinding == "" {
		pageBinding = "LeftToRight"
	}
	facingPages := "false"
	if opts.FacingPages {
		facingPages = "true"
	}
	measurementUnits := opts.MeasurementUnits
	if measurementUnits == "" {
		measurementUnits = "Points"
	}

	return &templateData{
		DOMVersion:  opts.DOMVersion,
		Orientation: opts.Orientation,

		PageWidthStr:  num(dims.Width),
		PageHeightStr: num(dims.Height),
		HalfHeightStr: num(dims.Height / 2),

		ColumnCount:      opts.ColumnCount,
		ColumnGutterStr:  num(opts.ColumnGutter),
		ColumnWidthStr:   num(columnWidth),
		ColumnsPositions: strings.Join(positions, " "),

		MarginTopStr:    num(opts.Margins.Top),
		MarginBottomStr: num(opts.Margins.Bottom),
		MarginLeftStr:   num(opts.Margins.Left),
		MarginRightStr:  num(opts.Margins.Right),

		Intent:           intent,
		PageBinding:      pageBinding,
		FacingPages:      facingPages,
		MeasurementUnits: measurementUnits,

		// El marco va centrado en la caja de márgenes. El origen vertical del spread
		// está en el centro de la página, de ahí el desplazamiento de media altura.
		FrameCenterXStr:    num(opts.Margins.Left + usableWidth/2),
		FrameCenterYStr:    num(opts.Margins.Top + usableHeight/2 - dims.Height/2),
		FrameHalfWidthStr:  num(usableWidth / 2),
		FrameHalfHeightStr: num(usableHeight / 2),

		Timestamp:  time.Now().Format(time.RFC3339),
		InstanceID: instanceID,
		DocumentID: documentID,
	}, nil
}

// newUUID genera un identificador aleatorio con la forma que XMP espera en
// InstanceID y DocumentID. Se usa crypto/rand de la biblioteca estándar para no
// añadir una dependencia por ocho líneas.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("no se pudo generar el identificador XMP: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // versión 4
	b[8] = (b[8] & 0x3f) | 0x80 // variante RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// renderTemplate ejecuta una plantilla de la forma mínima con los datos calculados.
func renderTemplate(name string, tmpl *template.Template, err error, data *templateData) ([]byte, error) {
	if err != nil {
		return nil, common.WrapError("idml", "create from template", fmt.Errorf("error al parsear template de %s: %w", name, err))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, common.WrapError("idml", "create from template", fmt.Errorf("error al ejecutar template de %s: %w", name, err))
	}
	return buf.Bytes(), nil
}

// addFileFromTemplate agrega un archivo al paquete desde datos de template.
// Es un helper para NewFromTemplate.
func (p *Package) addFileFromTemplate(path string, data []byte) error {
	// Crear una copia de los datos para evitar compartir el slice embebido
	fileCopy := make([]byte, len(data))
	copy(fileCopy, data)

	// Crear un header ZIP básico
	header := &zip.FileHeader{
		Name:   path,
		Method: zip.Deflate,
	}
	header.SetMode(0644)

	// Agregar al mapa de archivos con metadatos correctos
	p.files[path] = &fileEntry{
		data:   fileCopy,
		header: header,
	}

	// Preservar el orden de archivos
	p.fileOrder = append(p.fileOrder, path)

	return nil
}
