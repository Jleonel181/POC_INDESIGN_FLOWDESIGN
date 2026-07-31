package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Rectangle representa un elemento de página rectangular en un spread.
// Los rectángulos pueden contener texto, gráficos, o ser frames vacíos.
type Rectangle struct {
	PageItemBase

	// Contenido y visualización
	ContentType             string `xml:"ContentType,attr,omitempty"` // "TextType", "GraphicType", "Unassigned"
	StoryTitle              string `xml:"StoryTitle,attr,omitempty"`
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`

	// Restricciones de layout
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"` // ej: "FlexibleDimension FixedDimension FlexibleDimension"
	VerticalLayoutConstraints   string `xml:"VerticalLayoutConstraints,attr,omitempty"`

	// Propiedades de degradado
	GradientFillStart          string `xml:"GradientFillStart,attr,omitempty"`
	GradientFillLength         string `xml:"GradientFillLength,attr,omitempty"`
	GradientFillAngle          string `xml:"GradientFillAngle,attr,omitempty"`
	GradientFillHiliteLength   string `xml:"GradientFillHiliteLength,attr,omitempty"`
	GradientFillHiliteAngle    string `xml:"GradientFillHiliteAngle,attr,omitempty"`
	GradientStrokeStart        string `xml:"GradientStrokeStart,attr,omitempty"`
	GradientStrokeLength       string `xml:"GradientStrokeLength,attr,omitempty"`
	GradientStrokeAngle        string `xml:"GradientStrokeAngle,attr,omitempty"`
	GradientStrokeHiliteLength string `xml:"GradientStrokeHiliteLength,attr,omitempty"`
	GradientStrokeHiliteAngle  string `xml:"GradientStrokeHiliteAngle,attr,omitempty"`

	// Capa y bloqueo
	Locked              string `xml:"Locked,attr,omitempty"`              // "true" o "false"
	LocalDisplaySetting string `xml:"LocalDisplaySetting,attr,omitempty"` // "Default", etc.

	// Estilo y transformación
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Seguimiento de versión (para documentos complejos)
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr"`

	// Elementos hijo
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	FrameFittingOption *FrameFittingOption `xml:"FrameFittingOption,omitempty"`
	ObjectExportOption *ObjectExportOption `xml:"ObjectExportOption,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	InCopyExportOption *InCopyExportOption `xml:"InCopyExportOption,omitempty"`
	Image              *Image              `xml:"Image,omitempty"`
	PDF                *PDF                `xml:"PDF,omitempty"`

	// Comodín para otros elementos

	// OtherAttrs recoge los atributos que este tipo todavía no declara, para que no
	// se pierdan en el ciclo de lectura y escritura. La etiqueta `,any,attr` es de
	// encoding/xml: al leer recoge solo los atributos que no encajaron en ningún otro
	// campo, en su orden, y al escribir los emite después de los declarados.
	//
	// Límite conocido: encoding/xml corrompe los atributos con prefijo de namespace al
	// re-emitirlos. No aplica aquí: se inspeccionaron los 590 elementos de estos tipos
	// en los cinco documentos del corpus y ninguno lleva un atributo con prefijo. Si
	// algún día aparece uno, este es el sitio que hay que mirar.
	OtherAttrs    []xml.Attr             `xml:",any,attr"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// FrameFittingOption controla cómo el contenido se ajusta dentro de un frame.
// Es crítico para rectángulos con imágenes.
type FrameFittingOption struct {
	AutoFit             string `xml:"AutoFit,attr,omitempty"` // "true" o "false"
	LeftCrop            string `xml:"LeftCrop,attr,omitempty"`
	TopCrop             string `xml:"TopCrop,attr,omitempty"`
	RightCrop           string `xml:"RightCrop,attr,omitempty"`
	BottomCrop          string `xml:"BottomCrop,attr,omitempty"`
	FittingOnEmptyFrame string `xml:"FittingOnEmptyFrame,attr,omitempty"` // "None", "FitContentProportionally", etc.
	FittingAlignment    string `xml:"FittingAlignment,attr,omitempty"`    // "TopLeftAnchor", "CenterAnchor", etc.

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// FrameContentBase contiene los atributos comunes compartidos por el contenido Image y PDF.
// Estos representan contenido colocado dentro de un frame (Rectangle, Oval, Polygon).
type FrameContentBase struct {
	// Identificación principal
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr,omitempty"`

	// Visualización y estilo
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"`
	ImageTypeName           string `xml:"ImageTypeName,attr,omitempty"` // ej: "$ID/Portable Network Graphics (PNG)" o "$ID/Adobe Portable Document Format (PDF)"
	AppliedObjectStyle      string `xml:"AppliedObjectStyle,attr,omitempty"`
	Visible                 string `xml:"Visible,attr,omitempty"`

	// Restricciones de layout
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"`
	VerticalLayoutConstraints   string `xml:"VerticalLayoutConstraints,attr,omitempty"`

	// Transformación (posición, rotación, escala)
	ItemTransform string `xml:"ItemTransform,attr,omitempty"`

	// Seguimiento de versión
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr"`
}

// Image representa una imagen colocada dentro de un frame (típicamente un Rectangle).
// Las imágenes están vinculadas a archivos externos.
type Image struct {
	FrameContentBase

	// Propiedades específicas de la imagen
	Space                string `xml:"Space,attr,omitempty"`
	ActualPpi            string `xml:"ActualPpi,attr,omitempty"`            // formato "72 72"
	EffectivePpi         string `xml:"EffectivePpi,attr,omitempty"`         // formato "394 394"
	ImageRenderingIntent string `xml:"ImageRenderingIntent,attr,omitempty"` // "UseColorSettings", etc.

	// Propiedades de degradado
	GradientFillStart        string `xml:"GradientFillStart,attr,omitempty"`
	GradientFillLength       string `xml:"GradientFillLength,attr,omitempty"`
	GradientFillAngle        string `xml:"GradientFillAngle,attr,omitempty"`
	GradientFillHiliteLength string `xml:"GradientFillHiliteLength,attr,omitempty"`
	GradientFillHiliteAngle  string `xml:"GradientFillHiliteAngle,attr,omitempty"`

	// Elementos hijo
	Properties           *common.Properties    `xml:"Properties,omitempty"`
	ClippingPathSettings *ClippingPathSettings `xml:"ClippingPathSettings,omitempty"`
	ImageIOPreference    *ImageIOPreference    `xml:"ImageIOPreference,omitempty"`
	TextWrapPreference   *TextWrapPreference   `xml:"TextWrapPreference,omitempty"`
	Link                 *Link                 `xml:"Link,omitempty"`

	// Comodín para otros elementos

	// OtherAttrs recoge los atributos que este tipo todavía no declara, para que no
	// se pierdan en el ciclo de lectura y escritura. La etiqueta `,any,attr` es de
	// encoding/xml: al leer recoge solo los atributos que no encajaron en ningún otro
	// campo, en su orden, y al escribir los emite después de los declarados.
	//
	// Límite conocido: encoding/xml corrompe los atributos con prefijo de namespace al
	// re-emitirlos. No aplica aquí: se inspeccionaron los 590 elementos de estos tipos
	// en los cinco documentos del corpus y ninguno lleva un atributo con prefijo. Si
	// algún día aparece uno, este es el sitio que hay que mirar.
	OtherAttrs    []xml.Attr             `xml:",any,attr"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Link representa un vínculo a un archivo externo (imagen, etc.).
type Link struct {
	Self                       string `xml:"Self,attr"`
	AssetURL                   string `xml:"AssetURL,attr,omitempty"`
	AssetID                    string `xml:"AssetID,attr,omitempty"`
	LinkResourceURI            string `xml:"LinkResourceURI,attr,omitempty"` // "file:/path/to/image.jpg"
	LinkResourceFormat         string `xml:"LinkResourceFormat,attr,omitempty"`
	StoredState                string `xml:"StoredState,attr,omitempty"` // "Normal", "Modified", "Missing"
	LinkClassID                string `xml:"LinkClassID,attr,omitempty"`
	LinkClientID               string `xml:"LinkClientID,attr,omitempty"`
	LinkResourceModified       string `xml:"LinkResourceModified,attr,omitempty"` // "true" o "false"
	LinkObjectModified         string `xml:"LinkObjectModified,attr,omitempty"`
	ShowInUI                   string `xml:"ShowInUI,attr,omitempty"`
	CanEmbed                   string `xml:"CanEmbed,attr,omitempty"`
	CanUnembed                 string `xml:"CanUnembed,attr,omitempty"`
	CanPackage                 string `xml:"CanPackage,attr,omitempty"`
	ImportPolicy               string `xml:"ImportPolicy,attr,omitempty"` // "NoAutoImport"
	ExportPolicy               string `xml:"ExportPolicy,attr,omitempty"`
	LinkImportStamp            string `xml:"LinkImportStamp,attr,omitempty"`
	LinkImportModificationTime string `xml:"LinkImportModificationTime,attr,omitempty"`
	LinkImportTime             string `xml:"LinkImportTime,attr,omitempty"`
	LinkResourceSize           string `xml:"LinkResourceSize,attr,omitempty"`
	RenditionData              string `xml:"RenditionData,attr,omitempty"` // "Actual"

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ClippingPathSettings controla el comportamiento de recorte (clipping) de la imagen.
type ClippingPathSettings struct {
	ClippingType           string `xml:"ClippingType,attr,omitempty"` // "None", "PhotoshopPath", etc.
	InvertPath             string `xml:"InvertPath,attr,omitempty"`
	IncludeInsideEdges     string `xml:"IncludeInsideEdges,attr,omitempty"`
	RestrictToFrame        string `xml:"RestrictToFrame,attr,omitempty"`
	UseHighResolutionImage string `xml:"UseHighResolutionImage,attr,omitempty"`
	Threshold              string `xml:"Threshold,attr,omitempty"`
	Tolerance              string `xml:"Tolerance,attr,omitempty"`
	InsetFrame             string `xml:"InsetFrame,attr,omitempty"`
	AppliedPathName        string `xml:"AppliedPathName,attr,omitempty"`
	Index                  string `xml:"Index,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ImageIOPreference controla la configuración de importación/exportación de la imagen.
type ImageIOPreference struct {
	ApplyPhotoshopClippingPath string `xml:"ApplyPhotoshopClippingPath,attr,omitempty"`
	AllowAutoEmbedding         string `xml:"AllowAutoEmbedding,attr,omitempty"`
	AlphaChannelName           string `xml:"AlphaChannelName,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ContourOption controla el ajuste de contorno (contour wrapping).
type ContourOption struct {
	ContourType        string `xml:"ContourType,attr,omitempty"`
	IncludeInsideEdges string `xml:"IncludeInsideEdges,attr,omitempty"`
	ContourPathName    string `xml:"ContourPathName,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// TextWrapPreference controla cómo el texto se ajusta alrededor de los objetos.
type TextWrapPreference struct {
	Inverse               string             `xml:"Inverse,attr,omitempty"`
	ApplyToMasterPageOnly string             `xml:"ApplyToMasterPageOnly,attr,omitempty"`
	TextWrapSide          string             `xml:"TextWrapSide,attr,omitempty"` // "BothSides", "LeftSide", "RightSide"
	TextWrapMode          string             `xml:"TextWrapMode,attr,omitempty"` // "None", "BoundingBoxTextWrap", etc.
	Properties            *common.Properties `xml:"Properties,omitempty"`
	ContourOption         *ContourOption     `xml:"ContourOption,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// InCopyExportOption controla la configuración de exportación a InCopy.
type InCopyExportOption struct {
	XMLName               xml.Name `xml:"InCopyExportOption"`
	IncludeGraphicProxies string   `xml:"IncludeGraphicProxies,attr,omitempty"`
	IncludeAllResources   string   `xml:"IncludeAllResources,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ObjectExportOption controla la configuración de exportación para publicación web/digital.
type ObjectExportOption struct {
	XMLName xml.Name `xml:"ObjectExportOption"`
	// Placeholder para las opciones de exportación - la definición completa llegará en la Fase 5
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PDF representa un archivo PDF colocado dentro de un frame (típicamente un Rectangle).
// Los PDF pueden usarse para anuncios, documentos importados, o gráficos.
// Es similar a Image pero específico para contenido PDF.
type PDF struct {
	FrameContentBase

	// Configuración de política de color específica de PDF
	GrayVectorPolicy string `xml:"GrayVectorPolicy,attr,omitempty"` // "IgnoreAll", "HonorAllProfiles"
	RGBVectorPolicy  string `xml:"RGBVectorPolicy,attr,omitempty"`  // "IgnoreAll", "HonorAllProfiles"
	CMYKVectorPolicy string `xml:"CMYKVectorPolicy,attr,omitempty"` // "IgnoreAll", "HonorAllProfiles"

	// Elementos hijo
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	PDFAttribute       *PDFAttribute       `xml:"PDFAttribute,omitempty"`
	Link               *Link               `xml:"Link,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`

	// Comodín para otros elementos

	// OtherAttrs recoge los atributos que este tipo todavía no declara, para que no
	// se pierdan en el ciclo de lectura y escritura. La etiqueta `,any,attr` es de
	// encoding/xml: al leer recoge solo los atributos que no encajaron en ningún otro
	// campo, en su orden, y al escribir los emite después de los declarados.
	//
	// Límite conocido: encoding/xml corrompe los atributos con prefijo de namespace al
	// re-emitirlos. No aplica aquí: se inspeccionaron los 590 elementos de estos tipos
	// en los cinco documentos del corpus y ninguno lleva un atributo con prefijo. Si
	// algún día aparece uno, este es el sitio que hay que mirar.
	OtherAttrs    []xml.Attr             `xml:",any,attr"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// PDFAttribute contiene atributos específicos de PDF como el número de página y la configuración de recorte.
type PDFAttribute struct {
	PageNumber            string `xml:"PageNumber,attr,omitempty"`            // "1" (qué página del PDF multipágina mostrar)
	PDFCrop               string `xml:"PDFCrop,attr,omitempty"`               // "CropPDF", "CropContentBox", "CropMediaBox", etc.
	TransparentBackground string `xml:"TransparentBackground,attr,omitempty"` // "true" o "false"

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PageItem es el tipo común de los elementos de página que pueden ser hijos directos
// de un <Spread>: marcos de texto, rectángulos, elipses, polígonos, líneas gráficas,
// grupos y el contenido colocado suelto (Image, PDF).
//
// Por qué se declara aquí y no se reutiliza la de pkg/idml/interfaces.go: pkg/idml
// importa pkg/spread, así que usar la suya crearía un ciclo de importación.
//
// Los métodos llevan **receptor puntero** a propósito. Es la barrera automática contra
// el fallo silencioso de guardar una copia: con receptores valor, un `Rectangle` (valor)
// satisface la interfaz y el compilador acepta `Items = append(Items, rect)`, con lo que
// mutar el elemento a través de Items no afectaría a nada. Con receptor puntero eso no
// compila y hay que escribir `&rect`.
//
// xmlTag no está exportado, así que la interfaz queda sellada: ningún paquete de fuera
// puede declarar un tipo que la satisfaga, y la lista de tipos posibles es la de este
// archivo.
type PageItem interface {
	GetSelf() string
	GetItemLayer() string
	GetItemTransform() string
	GetGeometricBounds() string
	GetVisible() string
	GetName() string

	// xmlTag es el nombre del elemento XML de este tipo, que es lo que decide en qué
	// campo por tipo se guarda y con qué nombre se emite.
	xmlTag() string
}

// Nombres de elemento de los elementos de página. Están como constantes porque los usan
// tres sitios que tienen que coincidir: el reparto al parsear, la emisión y el registro
// de orden. Una errata entre ellos haría desaparecer elementos en silencio.
const (
	TagTextFrame   = "TextFrame"
	TagRectangle   = "Rectangle"
	TagImage       = "Image"
	TagOval        = "Oval"
	TagPolygon     = "Polygon"
	TagGraphicLine = "GraphicLine"
	TagGroup       = "Group"
	TagPDF         = "PDF"
)

func (f *SpreadTextFrame) xmlTag() string { return TagTextFrame }
func (r *Rectangle) xmlTag() string       { return TagRectangle }
func (i *Image) xmlTag() string           { return TagImage }
func (o *Oval) xmlTag() string            { return TagOval }
func (p *Polygon) xmlTag() string         { return TagPolygon }
func (g *GraphicLine) xmlTag() string     { return TagGraphicLine }
func (g *Group) xmlTag() string           { return TagGroup }
func (p *PDF) xmlTag() string             { return TagPDF }

// Los seis accesores de PageItem para el contenido colocado (Image y PDF), que embeben
// FrameContentBase en lugar de PageItemBase.
//
// ItemLayer y GeometricBounds devuelven cadena vacía porque el XML de InDesign no los
// pone en estos elementos: la capa y el bounding box son del marco que los contiene. Se
// devuelve el cero en lugar de omitir los métodos para que Image y PDF puedan estar en
// un []PageItem cuando aparecen como hijos directos de un <Spread>.

// GetSelf retorna el identificador único de este contenido.
func (f *FrameContentBase) GetSelf() string { return f.Self }

// GetName retorna el nombre de visualización.
func (f *FrameContentBase) GetName() string { return f.Name }

// GetItemTransform retorna la matriz de transformación de 6 valores.
func (f *FrameContentBase) GetItemTransform() string { return f.ItemTransform }

// GetVisible retorna el estado de visibilidad ("true" o "false").
func (f *FrameContentBase) GetVisible() string { return f.Visible }

// GetItemLayer retorna siempre cadena vacía: la capa la declara el marco contenedor.
func (f *FrameContentBase) GetItemLayer() string { return "" }

// GetGeometricBounds retorna siempre cadena vacía: el bounding box lo declara el marco
// contenedor.
func (f *FrameContentBase) GetGeometricBounds() string { return "" }

// Comprobaciones en tiempo de compilación de que los ocho tipos satisfacen PageItem con
// receptor puntero, y de que un valor **no** la satisface. Si alguien añade un tipo de
// elemento de página y olvida su xmlTag, esto es lo que falla.
var (
	_ PageItem = (*SpreadTextFrame)(nil)
	_ PageItem = (*Rectangle)(nil)
	_ PageItem = (*Image)(nil)
	_ PageItem = (*Oval)(nil)
	_ PageItem = (*Polygon)(nil)
	_ PageItem = (*GraphicLine)(nil)
	_ PageItem = (*Group)(nil)
	_ PageItem = (*PDF)(nil)
)
