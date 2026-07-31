package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// GraphicFile representa el archivo Resources/Graphic.xml que contiene colores, muestras y recursos gráficos.
//
// El elemento raíz es <idPkg:Graphic> con el namespace idPkg.
type GraphicFile struct {
	// XMLName no se establece directamente - se maneja manualmente en MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion es la versión del DOM de InDesign (ej., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// Colors define las muestras de color (CMYK, RGB, LAB, etc.)
	Colors []Color `xml:"Color,omitempty"`

	// Inks define las propiedades de tintas de impresión
	Inks []Ink `xml:"Ink,omitempty"`

	// Gradients define las definiciones de rellenos degradados
	Gradients []Gradient `xml:"Gradient,omitempty"`

	// Swatches define las referencias de muestras con nombre
	Swatches []Swatch `xml:"Swatch,omitempty"`

	// PastedSmoothShades define las definiciones de sombras suaves incrustadas
	PastedSmoothShades []PastedSmoothShade `xml:"PastedSmoothShade,omitempty"`

	// StrokeStyles define las definiciones de estilos de trazo/línea
	StrokeStyles []StrokeStyle `xml:"StrokeStyle,omitempty"`

	// Captura todos los demás elementos no modelados explícitamente
	OtherElements []common.RawXMLElement `xml:",any"`

	// childOrder recuerda el orden en que venían los hijos de <idPkg:Graphic>.
	// InDesign intercala Gradient, Swatch y PastedSmoothShade, y el orden de los
	// campos de arriba los reagrupa. Ver xmlutil.ChildOrder.
	childOrder xmlutil.ChildOrder
}

// Color representa una definición de muestra de color.
// Soporta colores Process (CMYK, RGB, LAB) y colores Spot (tintas especiales).
type Color struct {
	Self                      string `xml:"Self,attr"`
	Model                     string `xml:"Model,attr"`                         // "Process" o "Registration"
	Space                     string `xml:"Space,attr"`                         // "CMYK", "RGB", "LAB"
	ColorValue                string `xml:"ColorValue,attr"`                    // Valores separados por espacio (ej., "0 0 100 0")
	ColorOverride             string `xml:"ColorOverride,attr,omitempty"`       // "Normal", "Specialblack", "Hiddenreserved", etc.
	ConvertToHsb              string `xml:"ConvertToHsb,attr,omitempty"`        // "true" o "false"
	AlternateSpace            string `xml:"AlternateSpace,attr,omitempty"`      // "NoAlternateColor" o espacio de color alternativo
	AlternateColorValue       string `xml:"AlternateColorValue,attr,omitempty"` // Valores de color alternativo
	Name                      string `xml:"Name,attr"`
	ColorEditable             string `xml:"ColorEditable,attr,omitempty"`  // "true" o "false"
	ColorRemovable            string `xml:"ColorRemovable,attr,omitempty"` // "true" o "false"
	Visible                   string `xml:"Visible,attr,omitempty"`        // "true" o "false"
	SwatchCreatorID           string `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string `xml:"SwatchColorGroupReference,attr,omitempty"`
}

// Ink representa una definición de tinta de impresión.
// Controla el comportamiento de la tinta en la separación de color e impresión.
type Ink struct {
	Self             string `xml:"Self,attr"`
	Name             string `xml:"Name,attr"`
	Angle            string `xml:"Angle,attr,omitempty"`            // Ángulo de trama en grados
	ConvertToProcess string `xml:"ConvertToProcess,attr,omitempty"` // "true" o "false"
	Frequency        string `xml:"Frequency,attr,omitempty"`        // Frecuencia de trama en lpi
	NeutralDensity   string `xml:"NeutralDensity,attr,omitempty"`   // Valor de densidad óptica
	PrintInk         string `xml:"PrintInk,attr,omitempty"`         // "true" o "false"
	TrapOrder        string `xml:"TrapOrder,attr,omitempty"`        // Número de orden de trampa
	InkType          string `xml:"InkType,attr,omitempty"`          // "Normal", "Transparent", "Opaque"
}

// Gradient representa una definición de relleno degradado.
// Soporta tipos de degradado Lineal y Radial con múltiples paradas de color.
type Gradient struct {
	Self                      string         `xml:"Self,attr"`
	Type                      string         `xml:"Type,attr"` // "Linear" o "Radial"
	Name                      string         `xml:"Name,attr"`
	ColorEditable             string         `xml:"ColorEditable,attr,omitempty"`  // "true" o "false"
	ColorRemovable            string         `xml:"ColorRemovable,attr,omitempty"` // "true" o "false"
	Visible                   string         `xml:"Visible,attr,omitempty"`        // "true" o "false"
	SwatchCreatorID           string         `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string         `xml:"SwatchColorGroupReference,attr,omitempty"`
	GradientStops             []GradientStop `xml:"GradientStop,omitempty"`
}

// GradientStop representa una parada de color en un degradado.
type GradientStop struct {
	Self      string `xml:"Self,attr"`
	StopColor string `xml:"StopColor,attr"`          // Referencia a un Color (ej., "Color/Black")
	Location  string `xml:"Location,attr"`           // Posición 0-100
	Midpoint  string `xml:"Midpoint,attr,omitempty"` // Posición del punto medio (0-100)
}

// Swatch representa una referencia de muestra con nombre.
// Típicamente referencia "None" u otras muestras del panel de muestras.
type Swatch struct {
	Self                      string `xml:"Self,attr"`
	Name                      string `xml:"Name,attr"`
	ColorEditable             string `xml:"ColorEditable,attr,omitempty"`  // "true" o "false"
	ColorRemovable            string `xml:"ColorRemovable,attr,omitempty"` // "true" o "false"
	Visible                   string `xml:"Visible,attr,omitempty"`        // "true" o "false"
	SwatchCreatorID           string `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string `xml:"SwatchColorGroupReference,attr,omitempty"`
}

// PastedSmoothShade representa una definición de sombra suave incrustada.
// Contiene datos de sombra embebidos en la propiedad Contents.
type PastedSmoothShade struct {
	Self                      string             `xml:"Self,attr"`
	ContentsVersion           string             `xml:"ContentsVersion,attr,omitempty"`
	ContentsType              string             `xml:"ContentsType,attr,omitempty"` // "ConstantShade", etc.
	SpotColorList             string             `xml:"SpotColorList,attr,omitempty"`
	ContentsEncoding          string             `xml:"ContentsEncoding,attr,omitempty"` // "Ascii64Encoding"
	ContentsMatrix            string             `xml:"ContentsMatrix,attr,omitempty"`   // Matriz de transformación
	Name                      string             `xml:"Name,attr"`
	ColorEditable             string             `xml:"ColorEditable,attr,omitempty"`  // "true" o "false"
	ColorRemovable            string             `xml:"ColorRemovable,attr,omitempty"` // "true" o "false"
	Visible                   string             `xml:"Visible,attr,omitempty"`        // "true" o "false"
	SwatchCreatorID           string             `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string             `xml:"SwatchColorGroupReference,attr,omitempty"`
	Properties                *common.Properties `xml:"Properties,omitempty"` // Contiene <Contents> CDATA
}

// StrokeStyle representa una definición de estilo de trazo/línea.
// Define patrones para líneas y trazos (sólido, discontinuo, punteado, etc.).
type StrokeStyle struct {
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr"`
	// Aquí irían propiedades adicionales del trazo
	OtherElements []common.RawXMLElement `xml:",any"`
}
