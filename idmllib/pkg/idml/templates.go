package idml

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"sync"
	"text/template"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Inicialización de templates con sync.Once para carga diferida thread-safe
var (
	designmapTmpl     *template.Template
	designmapTmplErr  error
	designmapTmplOnce sync.Once

	masterspreadTmpl     *template.Template
	masterspreadTmplErr  error
	masterspreadTmplOnce sync.Once
)

// getDesignmapTemplate devuelve el template de designmap ya parseado.
// El template se parsea una sola vez y se cachea para llamadas posteriores.
func getDesignmapTemplate() (*template.Template, error) {
	designmapTmplOnce.Do(func() {
		designmapTmpl, designmapTmplErr = template.New("designmap").Parse(string(minimalDesignMap))
	})
	return designmapTmpl, designmapTmplErr
}

// getMasterspreadTemplate devuelve el template de masterspread ya parseado.
// El template se parsea una sola vez y se cachea para llamadas posteriores.
func getMasterspreadTemplate() (*template.Template, error) {
	masterspreadTmplOnce.Do(func() {
		masterspreadTmpl, masterspreadTmplErr = template.New("masterspread").Parse(string(minimalMasterSpread))
	})
	return masterspreadTmpl, masterspreadTmplErr
}

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

	pkg := New()

	// Obtener dimensiones de página
	dims := opts.GetDimensions()

	// Generar designmap.xml personalizado
	designmap, err := generateDesignMap(opts, dims)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathDesignmap, err)
	}
	if err := pkg.addFileFromTemplate(PathDesignmap, designmap); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathDesignmap, err)
	}

	// Generar MasterSpread personalizado
	masterSpread, err := generateMasterSpread(opts, dims)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMasterSpread, err)
	}
	if err := pkg.addFileFromTemplate(PathMasterSpread, masterSpread); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMasterSpread, err)
	}

	// Agregar Preferences.xml mínimo
	if err := pkg.addFileFromTemplate(PathPreferences, minimalPreferences); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathPreferences, err)
	}

	// CRÍTICO: Agregar archivo mimetype (debe ser el primero y sin comprimir)
	if err := pkg.addFileFromTemplate(PathMimetype, minimalMimetype); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMimetype, err)
	}

	// Agregar archivos META-INF
	if err := pkg.addFileFromTemplate(PathContainer, minimalContainer); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathContainer, err)
	}

	// Agregar archivos de recursos requeridos
	if err := pkg.addFileFromTemplate(PathGraphic, minimalGraphic); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathGraphic, err)
	}

	if err := pkg.addFileFromTemplate(PathFonts, minimalFonts); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathFonts, err)
	}

	if err := pkg.addFileFromTemplate(PathStyles, minimalStyles); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathStyles, err)
	}

	// Add XML/Tags.xml
	if err := pkg.addFileFromTemplate(PathTags, minimalTags); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathTags, err)
	}

	return pkg, nil
}

// generateDesignMap crea un designmap.xml personalizado según las opciones.
func generateDesignMap(opts *TemplateOptions, dims PageDimensions) ([]byte, error) {
	tmpl, err := getDesignmapTemplate()
	if err != nil {
		return nil, common.WrapError("idml", "generate design map", fmt.Errorf("error al parsear template de designmap: %w", err))
	}

	data := struct {
		DOMVersion   string
		PageWidth    float64
		PageHeight   float64
		Orientation  string
		ColumnCount  int
		ColumnGutter float64
	}{
		DOMVersion:   opts.DOMVersion,
		PageWidth:    dims.Width,
		PageHeight:   dims.Height,
		Orientation:  opts.Orientation,
		ColumnCount:  opts.ColumnCount,
		ColumnGutter: opts.ColumnGutter,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, common.WrapError("idml", "generate design map", fmt.Errorf("error al ejecutar template de designmap: %w", err))
	}

	return buf.Bytes(), nil
}

// generateMasterSpread crea un MasterSpread personalizado según las opciones.
func generateMasterSpread(opts *TemplateOptions, dims PageDimensions) ([]byte, error) {
	tmpl, err := getMasterspreadTemplate()
	if err != nil {
		return nil, common.WrapError("idml", "generate master spread", fmt.Errorf("error al parsear template de masterspread: %w", err))
	}

	data := struct {
		DOMVersion   string
		PageWidth    float64
		PageHeight   float64
		CenterX      float64
		CenterY      float64
		ColumnCount  int
		ColumnGutter float64
	}{
		DOMVersion:   opts.DOMVersion,
		PageWidth:    dims.Width,
		PageHeight:   dims.Height,
		CenterX:      dims.Width / 2,
		CenterY:      dims.Height / 2,
		ColumnCount:  opts.ColumnCount,
		ColumnGutter: opts.ColumnGutter,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, common.WrapError("idml", "generate master spread", fmt.Errorf("error al ejecutar template de masterspread: %w", err))
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
