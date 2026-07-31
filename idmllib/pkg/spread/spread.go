package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// PageItemBase contiene los atributos comunes compartidos por todos los elementos de página.
// Este struct se embebe en tipos específicos de elementos de página para reducir duplicación
// y proveer una implementación de interfaz consistente.
type PageItemBase struct {
	// Identificación principal
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr,omitempty"`

	// Capa y visibilidad
	ItemLayer string `xml:"ItemLayer,attr,omitempty"`
	Visible   string `xml:"Visible,attr,omitempty"` // "true" o "false"

	// Geometría y transformación
	GeometricBounds string `xml:"GeometricBounds,attr,omitempty"` // formato "y1 x1 y2 x2"
	ItemTransform   string `xml:"ItemTransform,attr,omitempty"`   // matriz de transformación de 6 valores
}

// GetSelf retorna el identificador único de este elemento de página.
func (p *PageItemBase) GetSelf() string {
	return p.Self
}

// GetItemLayer retorna el ID de la capa en la que está este elemento de página.
func (p *PageItemBase) GetItemLayer() string {
	return p.ItemLayer
}

// GetGeometricBounds retorna el bounding box en formato "y1 x1 y2 x2".
func (p *PageItemBase) GetGeometricBounds() string {
	return p.GeometricBounds
}

// GetItemTransform retorna la matriz de transformación de 6 valores.
func (p *PageItemBase) GetItemTransform() string {
	return p.ItemTransform
}

// GetVisible retorna el estado de visibilidad ("true" o "false").
func (p *PageItemBase) GetVisible() string {
	return p.Visible
}

// GetName retorna el nombre de visualización del elemento de página.
func (p *PageItemBase) GetName() string {
	return p.Name
}

// Spread representa un spread (layout de página) en un documento IDML.
// Los spreads contienen páginas, guías y elementos de página como frames de texto e imágenes.
//
// El elemento raíz es <idPkg:Spread> con el namespace idPkg.
//
// DECISIÓN DE DISEÑO: Enfoque de estructura dual
// Este tipo usa una estructura dual para manejar el patrón de wrapper con namespace de IDML:
// - Spread externo: Maneja el wrapper <idPkg:Spread> con namespace y DOMVersion
// - SpreadElement interno: Contiene el contenido real de <Spread> con los elementos de página
// Esta separación permite un marshaling XML limpio mientras provee métodos de acceso convenientes.
// La alternativa sería marshaling custom complejo para cada método de acceso.
type Spread struct {
	// XMLName no se establece directamente - se maneja manualmente en MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion es la versión del DOM de InDesign (ej: "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// Elemento spread interno (el Spread real, no el wrapper)
	// DECISIÓN DE DISEÑO: El struct embebido provee acceso directo al contenido
	// mientras mantiene la estructura del wrapper con namespace para compatibilidad XML.
	InnerSpread SpreadElement `xml:"-"`
}

// TextFrames retorna todos los frames de texto en este spread.
// Es un método de conveniencia para acceder a los frames sin navegar por InnerSpread.
func (s *Spread) TextFrames() []SpreadTextFrame {
	return s.InnerSpread.TextFrames
}

// Pages retorna todas las páginas en este spread.
// Es un método de conveniencia para acceder a las páginas sin navegar por InnerSpread.
func (s *Spread) Pages() []Page {
	return s.InnerSpread.Pages
}

// Rectangles retorna todos los rectángulos en este spread.
// Es un método de conveniencia para acceder a los rectángulos sin navegar por InnerSpread.
func (s *Spread) Rectangles() []Rectangle {
	return s.InnerSpread.Rectangles
}

// Images retorna todas las imágenes en este spread.
// Es un método de conveniencia para acceder a las imágenes sin navegar por InnerSpread.
func (s *Spread) Images() []Image {
	return s.InnerSpread.Images
}

// Ovals retorna todas las elipses en este spread.
// Es un método de conveniencia para acceder a las elipses sin navegar por InnerSpread.
func (s *Spread) Ovals() []Oval {
	return s.InnerSpread.Ovals
}

// Polygons retorna todos los polígonos en este spread.
// Es un método de conveniencia para acceder a los polígonos sin navegar por InnerSpread.
func (s *Spread) Polygons() []Polygon {
	return s.InnerSpread.Polygons
}

// GraphicLines retorna todas las líneas gráficas en este spread.
// Es un método de conveniencia para acceder a las líneas sin navegar por InnerSpread.
func (s *Spread) GraphicLines() []GraphicLine {
	return s.InnerSpread.GraphicLines
}

// SpreadElement representa el elemento <Spread> real con todos sus atributos y elementos hijo.
type SpreadElement struct {
	XMLName xml.Name `xml:"Spread"`

	// Atributos principales
	Self                    string `xml:"Self,attr"`
	PageTransitionType      string `xml:"PageTransitionType,attr,omitempty"`
	PageTransitionDirection string `xml:"PageTransitionDirection,attr,omitempty"`
	PageTransitionDuration  string `xml:"PageTransitionDuration,attr,omitempty"`
	ShowMasterItems         string `xml:"ShowMasterItems,attr,omitempty"`
	PageCount               string `xml:"PageCount,attr,omitempty"`
	BindingLocation         string `xml:"BindingLocation,attr,omitempty"`
	SpreadHidden            string `xml:"SpreadHidden,attr,omitempty"`
	AllowPageShuffle        string `xml:"AllowPageShuffle,attr,omitempty"`
	ItemTransform           string `xml:"ItemTransform,attr,omitempty"`
	FlattenerOverride       string `xml:"FlattenerOverride,attr,omitempty"`

	// Items es la secuencia de elementos de página en Orden_Documental, con punteros a
	// los elementos guardados en los campos por tipo de abajo.
	//
	// Qué manda cada uno en esta fase: **Items manda el orden, los campos por tipo
	// mandan el contenido.** Al emitir, la secuencia sale del orden registrado al
	// parsear y el contenido se lee de los campos por tipo.
	//
	// La razón de ese reparto es un conflicto medido, no una preferencia. El criterio 1
	// de la Tarea 7 pide que Items sea la fuente de verdad de la serialización, y el
	// criterio 3 pide que las 18 escrituras existentes a los campos por tipo sigan
	// funcionando sin tocarlas. Las dos cosas no pueden ser ciertas a la vez:
	// `removeItemFromSpread` de pkg/idml borra del slice `TextFrames` de un spread ya
	// parseado, y si el contenido saliera de Items ese borrado no llegaría al archivo
	// escrito. Peor aún, `TestRemoveTextFrame_Basic` solo comprueba el slice en memoria,
	// así que la regresión sería silenciosa. La fase 2 migra esas escrituras a Append y
	// Remove, y entonces Items pasa a ser también el contenido.
	//
	// ponytail: en esta fase Items es una vista ordenada, no el contenedor. Un llamador
	// que haga `append` a un campo por tipo puede invalidar los punteros de Items, que
	// apuntan al array de ese slice. No afecta a lo que se emite, porque la emisión lee
	// del campo; sí afecta a quien lea Items después. La vía de mejora es la fase 2, que
	// convierte Items en el contenedor y deja los campos como valores derivados.
	Items []PageItem `xml:"-"`

	// childOrder recuerda la secuencia de hijos leída, para reproducirla al emitir.
	// Incluye FlattenerPreference, las Page, los elementos de página y los hijos no
	// modelados. Sin este registro el ciclo reagrupa los hijos por tipo.
	childOrder xmlutil.ChildOrder

	// Elementos hijo
	FlattenerPreference *FlattenerPreference `xml:"FlattenerPreference,omitempty"`
	Pages               []Page               `xml:"Page,omitempty"`
	TextFrames          []SpreadTextFrame    `xml:"TextFrame,omitempty"`
	Rectangles          []Rectangle          `xml:"Rectangle,omitempty"`
	Images              []Image              `xml:"Image,omitempty"`
	Ovals               []Oval               `xml:"Oval,omitempty"`
	Polygons            []Polygon            `xml:"Polygon,omitempty"`
	GraphicLines        []GraphicLine        `xml:"GraphicLine,omitempty"`
	Groups              []Group              `xml:"Group,omitempty"`

	// Comodín para otros elementos aún no modelados explícitamente

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

// FlattenerPreference contiene configuraciones para el aplanado de transparencia.
type FlattenerPreference struct {
	LineArtAndTextResolution    string             `xml:"LineArtAndTextResolution,attr,omitempty"`
	GradientAndMeshResolution   string             `xml:"GradientAndMeshResolution,attr,omitempty"`
	ClipComplexRegions          string             `xml:"ClipComplexRegions,attr,omitempty"`
	ConvertAllStrokesToOutlines string             `xml:"ConvertAllStrokesToOutlines,attr,omitempty"`
	ConvertAllTextToOutlines    string             `xml:"ConvertAllTextToOutlines,attr,omitempty"`
	Properties                  *common.Properties `xml:"Properties,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// Page representa una página dentro de un spread.
type Page struct {
	// Atributos principales
	Self                   string `xml:"Self,attr"`
	TabOrder               string `xml:"TabOrder,attr"`
	AppliedMaster          string `xml:"AppliedMaster,attr,omitempty"`
	OverrideList           string `xml:"OverrideList,attr"`
	MasterPageTransform    string `xml:"MasterPageTransform,attr,omitempty"`
	Name                   string `xml:"Name,attr,omitempty"`
	AppliedTrapPreset      string `xml:"AppliedTrapPreset,attr,omitempty"`
	GeometricBounds        string `xml:"GeometricBounds,attr,omitempty"`
	ItemTransform          string `xml:"ItemTransform,attr,omitempty"`
	AppliedAlternateLayout string `xml:"AppliedAlternateLayout,attr,omitempty"`
	LayoutRule             string `xml:"LayoutRule,attr,omitempty"`
	SnapshotBlendingMode   string `xml:"SnapshotBlendingMode,attr,omitempty"`
	OptionalPage           string `xml:"OptionalPage,attr,omitempty"`
	GridStartingPoint      string `xml:"GridStartingPoint,attr,omitempty"`
	UseMasterGrid          string `xml:"UseMasterGrid,attr,omitempty"`

	// Elementos hijo
	Properties          *common.Properties          `xml:"Properties,omitempty"`
	Guides              []Guide                     `xml:"Guide,omitempty"`
	MarginPreference    *MarginPreference           `xml:"MarginPreference,omitempty"`
	GridDataInformation *common.GridDataInformation `xml:"GridDataInformation,omitempty"`

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

// Guide representa una guía de regla en una página.
type Guide struct {
	Self                    string             `xml:"Self,attr"`
	OverriddenPageItemProps string             `xml:"OverriddenPageItemProps,attr"`
	Orientation             string             `xml:"Orientation,attr,omitempty"`
	Location                string             `xml:"Location,attr,omitempty"`
	FitToPage               string             `xml:"FitToPage,attr,omitempty"`
	ViewThreshold           string             `xml:"ViewThreshold,attr,omitempty"`
	Locked                  string             `xml:"Locked,attr,omitempty"`
	ItemLayer               string             `xml:"ItemLayer,attr,omitempty"`
	PageIndex               string             `xml:"PageIndex,attr,omitempty"`
	GuideType               string             `xml:"GuideType,attr,omitempty"`
	GuideZone               string             `xml:"GuideZone,attr,omitempty"`
	Properties              *common.Properties `xml:"Properties,omitempty"`

	// OtherAttrs recoge los atributos que este tipo todavía no declara, para que no
	// se pierdan en el ciclo de lectura y escritura. La etiqueta `,any,attr` es de
	// encoding/xml: al leer recoge solo los atributos que no encajaron en ningún otro
	// campo, en su orden, y al escribir los emite después de los declarados.
	//
	// Límite conocido: encoding/xml corrompe los atributos con prefijo de namespace al
	// re-emitirlos. No aplica aquí: se inspeccionaron los 590 elementos de estos tipos
	// en los cinco documentos del corpus y ninguno lleva un atributo con prefijo. Si
	// algún día aparece uno, este es el sitio que hay que mirar.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// MarginPreference contiene la configuración de márgenes de página.
type MarginPreference struct {
	ColumnCount      string `xml:"ColumnCount,attr,omitempty"`
	ColumnGutter     string `xml:"ColumnGutter,attr,omitempty"`
	Top              string `xml:"Top,attr,omitempty"`
	Bottom           string `xml:"Bottom,attr,omitempty"`
	Left             string `xml:"Left,attr,omitempty"`
	Right            string `xml:"Right,attr,omitempty"`
	ColumnDirection  string `xml:"ColumnDirection,attr,omitempty"`
	ColumnsPositions string `xml:"ColumnsPositions,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

type BasicFrame struct {
	Self string `xml:"Self,attr"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// SpreadTextFrame representa un frame de texto en un spread.
type SpreadTextFrame struct {
	PageItemBase

	// Atributos principales
	ParentStory             string `xml:"ParentStory,attr,omitempty"`       // ID de la story que contiene el texto
	PreviousTextFrame       string `xml:"PreviousTextFrame,attr,omitempty"` // Frame anterior en la cadena
	NextTextFrame           string `xml:"NextTextFrame,attr,omitempty"`     // Frame siguiente en la cadena
	ContentType             string `xml:"ContentType,attr,omitempty"`
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`

	// Restricciones de layout
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"`
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
	Locked              string `xml:"Locked,attr,omitempty"`
	LocalDisplaySetting string `xml:"LocalDisplaySetting,attr,omitempty"`

	// Estilo y transformación
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Seguimiento de versión
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr"`

	// Elementos hijo
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Comodín para todos los demás atributos y elementos

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

// Oval representa un frame elíptico o circular en un spread.
// Las elipses pueden contener imágenes, texto o ser elementos decorativos vacíos.
type Oval struct {
	PageItemBase

	// Contenido y visualización
	ContentType string `xml:"ContentType,attr,omitempty"` // "Unassigned", "GraphicType", "TextType"

	// Capa y bloqueo
	LockState string `xml:"LockState,attr,omitempty"` // "None", etc.
	Locked    string `xml:"Locked,attr,omitempty"`    // "true" o "false"

	// Propiedades de borde
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"` // Ancho del borde en puntos
	StrokeType   string `xml:"StrokeType,attr,omitempty"`   // "Solid", "Dashed", etc.
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`  // Referencia a muestra de color
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`   // Porcentaje de tinta

	// Propiedades de relleno
	FillColor string `xml:"FillColor,attr,omitempty"`
	FillTint  string `xml:"FillTint,attr,omitempty"`

	// Estilos aplicados
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Configuración de visualización
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"`

	// Elementos hijo
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	Image              *Image              `xml:"Image,omitempty"` // Si la elipse contiene una imagen

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

// Polygon representa una figura de múltiples lados en un spread.
// Los polígonos pueden ser regulares (lados iguales) o irregulares, y pueden contener imágenes o texto.
type Polygon struct {
	PageItemBase

	// Contenido y visualización
	ContentType string `xml:"ContentType,attr,omitempty"` // "Unassigned", "GraphicType", "TextType"

	// Capa y bloqueo
	LockState string `xml:"LockState,attr,omitempty"` // "None", etc.
	Locked    string `xml:"Locked,attr,omitempty"`    // "true" o "false"

	// Propiedades de borde
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"` // Ancho del borde en puntos
	StrokeType   string `xml:"StrokeType,attr,omitempty"`   // "Solid", "Dashed", etc.
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`  // Referencia a muestra de color
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`   // Porcentaje de tinta

	// Propiedades de relleno
	FillColor string `xml:"FillColor,attr,omitempty"`
	FillTint  string `xml:"FillTint,attr,omitempty"`

	// Estilos aplicados
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Configuración de visualización
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"`

	// Elementos hijo
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	Image              *Image              `xml:"Image,omitempty"` // Si el polígono contiene una imagen

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

// GraphicLine representa una línea o path en un spread.
// Las líneas gráficas son elementos de dibujo vectorial con propiedades de borde y puntas de flecha opcionales.
type GraphicLine struct {
	PageItemBase

	// Contenido y visualización
	ContentType string `xml:"ContentType,attr,omitempty"` // "Unassigned", "GraphicType"

	// Capa y bloqueo
	LockState string `xml:"LockState,attr,omitempty"` // "None", etc.
	Locked    string `xml:"Locked,attr,omitempty"`    // "true" o "false"

	// Propiedades de línea
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"` // Ancho de línea en puntos
	StrokeType   string `xml:"StrokeType,attr,omitempty"`   // "Solid", "Dashed", etc.
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`  // Referencia a muestra de color (ej: "Color/Black")
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`   // Porcentaje de tinta

	// Propiedades de relleno (las líneas típicamente no tienen relleno, incluido por completitud)
	FillColor string `xml:"FillColor,attr,omitempty"`
	FillTint  string `xml:"FillTint,attr,omitempty"`

	// Terminación y unión de línea
	EndCap     string `xml:"EndCap,attr,omitempty"`     // "ButtEndCap", "RoundEndCap", "ProjectingEndCap"
	EndJoin    string `xml:"EndJoin,attr,omitempty"`    // "MiterEndJoin", "RoundEndJoin", "BevelEndJoin"
	MiterLimit string `xml:"MiterLimit,attr,omitempty"` // Límite de inglete para esquinas pronunciadas

	// Puntas de flecha
	LeftLineEnd  string `xml:"LeftLineEnd,attr,omitempty"`  // "None", "SimpleArrow", etc.
	RightLineEnd string `xml:"RightLineEnd,attr,omitempty"` // "None", "SimpleArrow", etc.

	// Estilos aplicados
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Configuración de visualización
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"` // "Default", etc.

	// Propiedades de degradado (para bordes y rellenos con degradado)
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

	// Restricciones de layout
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"`
	VerticalLayoutConstraints   string `xml:"VerticalLayoutConstraints,attr,omitempty"`

	// Seguimiento de versión (para documentos complejos)
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr"`

	// Elementos hijo
	PathGeometry       *common.PathGeometry `xml:"PathGeometry,omitempty"`
	Properties         *common.Properties   `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference  `xml:"TextWrapPreference,omitempty"`
	ObjectExportOption *ObjectExportOption  `xml:"ObjectExportOption,omitempty"`

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

// Group representa una colección de elementos de página agrupados.
type Group struct {
	PageItemBase
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

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
