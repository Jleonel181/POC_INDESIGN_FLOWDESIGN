package resources

import (
	"encoding/xml"
	"strconv"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// StylesFile representa el archivo Resources/Styles.xml que contiene las definiciones de estilos.
//
// El elemento raíz es <idPkg:Styles> con el namespace idPkg.
type StylesFile struct {
	// XMLName no se establece directamente - se maneja manualmente en MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion es la versión del DOM de InDesign (ej., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// RootCharacterStyleGroup contiene todas las definiciones de estilos de carácter
	RootCharacterStyleGroup *CharacterStyleGroup `xml:"RootCharacterStyleGroup,omitempty"`

	// RootParagraphStyleGroup contiene todas las definiciones de estilos de párrafo
	RootParagraphStyleGroup *ParagraphStyleGroup `xml:"RootParagraphStyleGroup,omitempty"`

	// RootCellStyleGroup contiene las definiciones de estilos de celda de tabla
	RootCellStyleGroup *CellStyleGroup `xml:"RootCellStyleGroup,omitempty"`

	// RootTableStyleGroup contiene las definiciones de estilos de tabla
	RootTableStyleGroup *TableStyleGroup `xml:"RootTableStyleGroup,omitempty"`

	// RootObjectStyleGroup contiene las definiciones de estilos de objeto
	RootObjectStyleGroup *ObjectStyleGroup `xml:"RootObjectStyleGroup,omitempty"`

	// TOCStyles contiene las definiciones de estilos de tabla de contenidos
	TOCStyles []TOCStyle `xml:"TOCStyle,omitempty"`

	// Captura todos los demás elementos
	OtherElements []common.RawXMLElement `xml:",any"`

	// childOrder recuerda el orden en que venían los hijos de <idPkg:Styles>.
	// InDesign coloca TOCStyle entre los grupos de estilos, en tercera posición, y el
	// orden de los campos de arriba lo empuja a la sexta. Ver xmlutil.ChildOrder.
	childOrder xmlutil.ChildOrder
}

// CharacterStyleGroup representa un grupo de estilos de carácter.
// Soporta grupos anidados para organización jerárquica (ej., "Naviga:Standard").
// XMLName será "RootCharacterStyleGroup" para la raíz o "CharacterStyleGroup" para anidados.
type CharacterStyleGroup struct {
	XMLName         xml.Name               // Se establece durante el unmarshal
	Self            string                 `xml:"Self,attr"`
	Name            string                 `xml:"Name,attr,omitempty"`
	CharacterStyles []CharacterStyle       `xml:"CharacterStyle,omitempty"`
	NestedGroups    []CharacterStyleGroup  `xml:"CharacterStyleGroup,omitempty"` // Grupos anidados
	OtherElements   []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// CharacterStyle representa una definición de estilo de carácter.
// Los estilos de carácter aplican formato al texto seleccionado dentro de un párrafo.
type CharacterStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	Imported                 string `xml:"Imported,attr,omitempty"`                 // "true" o "false"
	SplitDocument            string `xml:"SplitDocument,attr,omitempty"`            // "true" o "false"
	EmitCss                  string `xml:"EmitCss,attr,omitempty"`                  // "true" o "false"
	StyleUniqueId            string `xml:"StyleUniqueId,attr,omitempty"`            // UUID
	IncludeClass             string `xml:"IncludeClass,attr,omitempty"`             // "true" o "false"
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"` // Atajo de teclado

	// Atributos básicos de formato de texto
	FontStyle   string `xml:"FontStyle,attr,omitempty"`   // Nombre del estilo de fuente
	PointSize   string `xml:"PointSize,attr,omitempty"`   // Tamaño de fuente
	FillColor   string `xml:"FillColor,attr,omitempty"`   // Referencia al color de relleno
	StrokeColor string `xml:"StrokeColor,attr,omitempty"` // Referencia al color de trazo
	Underline   string `xml:"Underline,attr,omitempty"`   // "true" o "false"
	StrikeThru  string `xml:"StrikeThru,attr,omitempty"`  // "true" o "false"

	// Properties contiene configuraciones adicionales del estilo
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Captura todos los demás atributos y elementos
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// GetAppliedFont retorna la fuente aplicada desde Properties, o cadena vacía si no está definida.
// Es un método de conveniencia para acceder a la fuente sin verificar si Properties es nil.
func (cs *CharacterStyle) GetAppliedFont() string {
	if cs.Properties == nil {
		return ""
	}
	return cs.Properties.GetAppliedFont()
}

// GetPointSize retorna el tamaño en puntos como float64, o 0 si no está definido o es inválido.
// Es un método de conveniencia que maneja la conversión de string a float.
func (cs *CharacterStyle) GetPointSize() float64 {
	if cs.PointSize == "" {
		return 0
	}
	size, _ := strconv.ParseFloat(cs.PointSize, 64)
	return size
}

// ParagraphStyleGroup representa un grupo de estilos de párrafo.
// Soporta grupos anidados para organización jerárquica (ej., "Naviga:Standard").
// XMLName será "RootParagraphStyleGroup" para la raíz o "ParagraphStyleGroup" para anidados.
type ParagraphStyleGroup struct {
	XMLName         xml.Name               // Se establece durante el unmarshal
	Self            string                 `xml:"Self,attr"`
	Name            string                 `xml:"Name,attr,omitempty"`
	ParagraphStyles []ParagraphStyle       `xml:"ParagraphStyle,omitempty"`
	NestedGroups    []ParagraphStyleGroup  `xml:"ParagraphStyleGroup,omitempty"` // Grupos anidados
	OtherElements   []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ParagraphStyle representa una definición de estilo de párrafo.
// Los estilos de párrafo aplican formato completo a párrafos enteros.
// Nota: Debido a la gran cantidad de atributos (100+), se usa OtherElements
// para capturar todos los atributos modelando explícitamente solo los más críticos.
type ParagraphStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	Imported                 string `xml:"Imported,attr,omitempty"`
	NextStyle                string `xml:"NextStyle,attr,omitempty"` // Referencia al siguiente estilo de párrafo
	SplitDocument            string `xml:"SplitDocument,attr,omitempty"`
	EmitCss                  string `xml:"EmitCss,attr,omitempty"`
	StyleUniqueId            string `xml:"StyleUniqueId,attr,omitempty"`
	IncludeClass             string `xml:"IncludeClass,attr,omitempty"`
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string `xml:"KeyboardShortcut,attr,omitempty"`

	// Formato común de párrafo
	FontStyle       string `xml:"FontStyle,attr,omitempty"`
	PointSize       string `xml:"PointSize,attr,omitempty"`
	FillColor       string `xml:"FillColor,attr,omitempty"`
	Justification   string `xml:"Justification,attr,omitempty"` // "LeftAlign", "CenterAlign", "RightAlign", etc. (alineación)
	SpaceBefore     string `xml:"SpaceBefore,attr,omitempty"`
	SpaceAfter      string `xml:"SpaceAfter,attr,omitempty"`
	LeftIndent      string `xml:"LeftIndent,attr,omitempty"`
	RightIndent     string `xml:"RightIndent,attr,omitempty"`
	FirstLineIndent string `xml:"FirstLineIndent,attr,omitempty"`

	// Parámetros de ajuste de texto (para cálculos tipográficos precisos)
	Tracking            string `xml:"Tracking,attr,omitempty"`            // Espaciado entre letras (-25 = ajustado, 0 = normal, 25 = suelto)
	KerningMethod       string `xml:"KerningMethod,attr,omitempty"`       // "$ID/Optical" o "$ID/Metrics"
	MinimumWordSpacing  string `xml:"MinimumWordSpacing,attr,omitempty"`  // Porcentaje (90 = 90%)
	MaximumWordSpacing  string `xml:"MaximumWordSpacing,attr,omitempty"`  // Porcentaje (110 = 110%)
	MinimumGlyphScaling string `xml:"MinimumGlyphScaling,attr,omitempty"` // Porcentaje (97 = 97%)
	MaximumGlyphScaling string `xml:"MaximumGlyphScaling,attr,omitempty"` // Porcentaje (103 = 103%)

	// Properties contiene configuraciones adicionales del estilo (AppliedFont, Leading, TabList, etc.)
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Captura los muchos otros atributos (más de 100 atributos en total)
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// CellStyleGroup representa un grupo de estilos de celda de tabla.
type CellStyleGroup struct {
	XMLName       xml.Name               `xml:"RootCellStyleGroup"`
	Self          string                 `xml:"Self,attr"`
	CellStyles    []CellStyle            `xml:"CellStyle,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// CellStyle representa una definición de estilo de celda de tabla.
type CellStyle struct {
	Self                  string                 `xml:"Self,attr"`
	Name                  string                 `xml:"Name,attr"`
	AppliedParagraphStyle string                 `xml:"AppliedParagraphStyle,attr,omitempty"` // Referencia al estilo de párrafo
	Properties            *common.Properties     `xml:"Properties,omitempty"`
	OtherElements         []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// TableStyleGroup representa un grupo de estilos de tabla.
type TableStyleGroup struct {
	XMLName       xml.Name               `xml:"RootTableStyleGroup"`
	Self          string                 `xml:"Self,attr"`
	TableStyles   []TableStyle           `xml:"TableStyle,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// TableStyle representa una definición de estilo de tabla.
// Contiene configuraciones completas de borde, relleno y trazo de tabla.
type TableStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string `xml:"KeyboardShortcut,attr,omitempty"`

	// Espaciado de tabla
	SpaceBefore string `xml:"SpaceBefore,attr,omitempty"`
	SpaceAfter  string `xml:"SpaceAfter,attr,omitempty"`

	// Propiedades de borde (Superior, Izquierdo, Inferior, Derecho)
	TopBorderStrokeWeight  string `xml:"TopBorderStrokeWeight,attr,omitempty"`
	TopBorderStrokeColor   string `xml:"TopBorderStrokeColor,attr,omitempty"`
	LeftBorderStrokeWeight string `xml:"LeftBorderStrokeWeight,attr,omitempty"`
	LeftBorderStrokeColor  string `xml:"LeftBorderStrokeColor,attr,omitempty"`
	// ... muchos más atributos de borde

	Properties    *common.Properties     `xml:"Properties,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ObjectStyleGroup representa un grupo de estilos de objeto.
// Soporta grupos anidados para organización jerárquica.
// XMLName será "RootObjectStyleGroup" para la raíz o "ObjectStyleGroup" para anidados.
type ObjectStyleGroup struct {
	XMLName       xml.Name               // Will be set during unmarshal
	Self          string                 `xml:"Self,attr"`
	Name          string                 `xml:"Name,attr,omitempty"`
	ObjectStyles  []ObjectStyle          `xml:"ObjectStyle,omitempty"`
	NestedGroups  []ObjectStyleGroup     `xml:"ObjectStyleGroup,omitempty"` // Grupos anidados
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ObjectStyle representa una definición de estilo de objeto.
// Se aplica a marcos, cuadros de texto, gráficos y otros objetos de página.
type ObjectStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string `xml:"KeyboardShortcut,attr,omitempty"`
	AppliedParagraphStyle    string `xml:"AppliedParagraphStyle,attr,omitempty"`
	EmitCss                  string `xml:"EmitCss,attr,omitempty"`
	IncludeClass             string `xml:"IncludeClass,attr,omitempty"`

	// Propiedades de trazo y relleno
	FillColor    string `xml:"FillColor,attr,omitempty"`
	FillTint     string `xml:"FillTint,attr,omitempty"`
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"`

	// Propiedades de esquina
	TopLeftCornerOption string `xml:"TopLeftCornerOption,attr,omitempty"`
	TopLeftCornerRadius string `xml:"TopLeftCornerRadius,attr,omitempty"`
	CornerRadius        string `xml:"CornerRadius,attr,omitempty"`

	// Elementos hijo
	TransformAttributeOption *TransformAttributeOption `xml:"TransformAttributeOption,omitempty"`
	ObjectExportOption       *ObjectExportOption       `xml:"ObjectExportOption,omitempty"`
	TextFramePreference      *TextFramePreference      `xml:"TextFramePreference,omitempty"`

	Properties    *common.Properties     `xml:"Properties,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// TransformAttributeOption define los puntos de referencia de transformación para objetos.
type TransformAttributeOption struct {
	TransformAttrLeftReference  string `xml:"TransformAttrLeftReference,attr,omitempty"`
	TransformAttrTopReference   string `xml:"TransformAttrTopReference,attr,omitempty"`
	TransformAttrRefAnchorPoint string `xml:"TransformAttrRefAnchorPoint,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ObjectExportOption define la configuración de exportación de objetos.
type ObjectExportOption struct {
	AltTextSourceType     string                 `xml:"AltTextSourceType,attr,omitempty"`
	ActualTextSourceType  string                 `xml:"ActualTextSourceType,attr,omitempty"`
	CustomAltText         string                 `xml:"CustomAltText,attr,omitempty"`
	CustomActualText      string                 `xml:"CustomActualText,attr,omitempty"`
	ApplyTagType          string                 `xml:"ApplyTagType,attr,omitempty"`
	ImageConversionType   string                 `xml:"ImageConversionType,attr,omitempty"`
	ImageExportResolution string                 `xml:"ImageExportResolution,attr,omitempty"`
	Properties            *common.Properties     `xml:"Properties,omitempty"`
	OtherElements         []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// TextFramePreference define las preferencias del marco de texto.
type TextFramePreference struct {
	TextColumnCount       string                 `xml:"TextColumnCount,attr,omitempty"`
	TextColumnGutter      string                 `xml:"TextColumnGutter,attr,omitempty"`
	FirstBaselineOffset   string                 `xml:"FirstBaselineOffset,attr,omitempty"`
	VerticalJustification string                 `xml:"VerticalJustification,attr,omitempty"`
	AutoSizingType        string                 `xml:"AutoSizingType,attr,omitempty"`
	OtherElements         []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// TOCStyle representa una definición de estilo de tabla de contenidos.
type TOCStyle struct {
	Self                 string                 `xml:"Self,attr"`
	Name                 string                 `xml:"Name,attr"`
	Title                string                 `xml:"Title,attr,omitempty"`
	TitleStyle           string                 `xml:"TitleStyle,attr,omitempty"`           // Referencia al estilo de párrafo
	RunIn                string                 `xml:"RunIn,attr,omitempty"`                // "true" o "false"
	IncludeHidden        string                 `xml:"IncludeHidden,attr,omitempty"`        // "true" o "false"
	IncludeBookDocuments string                 `xml:"IncludeBookDocuments,attr,omitempty"` // "true" o "false"
	CreateBookmarks      string                 `xml:"CreateBookmarks,attr,omitempty"`      // "true" o "false"
	OtherElements        []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// FindParagraphStyle busca un estilo de párrafo por su ID Self.
// Busca en la jerarquía de grupos de estilos de párrafo, incluyendo grupos anidados.
func (sf *StylesFile) FindParagraphStyle(styleID string) *ParagraphStyle {
	if sf.RootParagraphStyleGroup == nil {
		return nil
	}
	return sf.findParagraphStyleInGroup(sf.RootParagraphStyleGroup, styleID)
}

// findParagraphStyleInGroup busca recursivamente un estilo de párrafo en un grupo.
func (sf *StylesFile) findParagraphStyleInGroup(group *ParagraphStyleGroup, styleID string) *ParagraphStyle {
	if group == nil {
		return nil
	}

	// Buscar en estilos directos
	for i := range group.ParagraphStyles {
		if group.ParagraphStyles[i].Self == styleID {
			return &group.ParagraphStyles[i]
		}
	}

	// Buscar en grupos anidados
	for i := range group.NestedGroups {
		if style := sf.findParagraphStyleInGroup(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}
