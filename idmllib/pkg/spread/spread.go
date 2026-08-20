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

// TextFrames retorna punteros a todos los frames de texto en este spread, en orden documental.
func (s *Spread) TextFrames() []*SpreadTextFrame {
	return s.InnerSpread.TextFrames()
}

// Pages retorna todas las páginas en este spread.
// Es un método de conveniencia para acceder a las páginas sin navegar por InnerSpread.
func (s *Spread) Pages() []Page {
	return s.InnerSpread.Pages
}

// Rectangles retorna punteros a todos los rectángulos en este spread, en orden documental.
func (s *Spread) Rectangles() []*Rectangle {
	return s.InnerSpread.Rectangles()
}

// Images retorna punteros a todas las imágenes en este spread, en orden documental.
func (s *Spread) Images() []*Image {
	return s.InnerSpread.Images()
}

// Ovals retorna punteros a todas las elipses en este spread, en orden documental.
func (s *Spread) Ovals() []*Oval {
	return s.InnerSpread.Ovals()
}

// Polygons retorna punteros a todos los polígonos en este spread, en orden documental.
func (s *Spread) Polygons() []*Polygon {
	return s.InnerSpread.Polygons()
}

// GraphicLines retorna punteros a todas las líneas gráficas en este spread, en orden documental.
func (s *Spread) GraphicLines() []*GraphicLine {
	return s.InnerSpread.GraphicLines()
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
	// los elementos guardados en los campos internos.
	//
	// Items manda el orden y los campos internos mandan el contenido. Al emitir, la
	// secuencia sale del registro de orden y el contenido se lee de los campos.
	//
	// La mutación pasa por Append (que registra el orden) o por los helpers AddX/RemoveXAt
	// (para operaciones sin registro de orden). Los campos son unexported; el acceso
	// externo es por los accesores TextFrames(), Rectangles(), etc. que devuelven []*T.
	Items []PageItem `xml:"-"`

	// childOrder recuerda la secuencia de hijos leída, para reproducirla al emitir.
	// Incluye FlattenerPreference, las Page, los elementos de página y los hijos no
	// modelados. Sin este registro el ciclo reagrupa los hijos por tipo.
	childOrder xmlutil.ChildOrder

	// Elementos hijo
	FlattenerPreference *FlattenerPreference `xml:"FlattenerPreference,omitempty"`
	Pages               []Page               `xml:"Page,omitempty"`
	textFrames          []SpreadTextFrame
	rectangles          []Rectangle
	images              []Image
	ovals               []Oval
	polygons            []Polygon
	graphicLines        []GraphicLine
	groups              []Group

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

// --- Accesores de SpreadElement: devuelven punteros a los elementos del contenedor ---

// TextFrames retorna punteros a los text frames, en orden documental.
func (se *SpreadElement) TextFrames() []*SpreadTextFrame {
	result := make([]*SpreadTextFrame, 0, len(se.textFrames))
	for i := range se.textFrames {
		result = append(result, &se.textFrames[i])
	}
	return result
}

// Rectangles retorna punteros a los rectángulos, en orden documental.
func (se *SpreadElement) Rectangles() []*Rectangle {
	result := make([]*Rectangle, 0, len(se.rectangles))
	for i := range se.rectangles {
		result = append(result, &se.rectangles[i])
	}
	return result
}

// Images retorna punteros a las imágenes, en orden documental.
func (se *SpreadElement) Images() []*Image {
	result := make([]*Image, 0, len(se.images))
	for i := range se.images {
		result = append(result, &se.images[i])
	}
	return result
}

// Ovals retorna punteros a las elipses, en orden documental.
func (se *SpreadElement) Ovals() []*Oval {
	result := make([]*Oval, 0, len(se.ovals))
	for i := range se.ovals {
		result = append(result, &se.ovals[i])
	}
	return result
}

// Polygons retorna punteros a los polígonos, en orden documental.
func (se *SpreadElement) Polygons() []*Polygon {
	result := make([]*Polygon, 0, len(se.polygons))
	for i := range se.polygons {
		result = append(result, &se.polygons[i])
	}
	return result
}

// GraphicLines retorna punteros a las líneas gráficas, en orden documental.
func (se *SpreadElement) GraphicLines() []*GraphicLine {
	result := make([]*GraphicLine, 0, len(se.graphicLines))
	for i := range se.graphicLines {
		result = append(result, &se.graphicLines[i])
	}
	return result
}

// Groups retorna punteros a los grupos, en orden documental.
func (se *SpreadElement) Groups() []*Group {
	result := make([]*Group, 0, len(se.groups))
	for i := range se.groups {
		result = append(result, &se.groups[i])
	}
	return result
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

// MasterSpread representa un master spread (plantilla de página maestra) en un documento IDML.
// Los master spreads definen el contenido base que se hereda en las páginas normales.
//
// Estructura dual como Spread: wrapper <idPkg:MasterSpread> con namespace + elemento
// <MasterSpread> interno.
type MasterSpread struct {
	XMLName xml.Name `xml:"-"`

	// DOMVersion es la versión del DOM de InDesign (ej: "21.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// Elemento master spread interno
	InnerMasterSpread MasterSpreadElement `xml:"-"`
}

// TextFrames retorna punteros a todos los frames de texto en este master spread.
func (ms *MasterSpread) TextFrames() []*SpreadTextFrame {
	return ms.InnerMasterSpread.TextFrames()
}

// Pages retorna todas las páginas en este master spread.
func (ms *MasterSpread) Pages() []Page {
	return ms.InnerMasterSpread.Pages
}

// Rectangles retorna punteros a todos los rectángulos en este master spread.
func (ms *MasterSpread) Rectangles() []*Rectangle {
	return ms.InnerMasterSpread.Rectangles()
}

// GraphicLines retorna punteros a todas las líneas gráficas en este master spread.
func (ms *MasterSpread) GraphicLines() []*GraphicLine {
	return ms.InnerMasterSpread.GraphicLines()
}

// MasterSpreadElement representa el elemento <MasterSpread> interno con todos sus
// atributos y elementos hijo. Reutiliza el mismo contenedor ordenado que SpreadElement.
type MasterSpreadElement struct {
	XMLName xml.Name `xml:"MasterSpread"`

	// Atributos principales
	Self                    string `xml:"Self,attr"`
	Name                    string `xml:"Name,attr,omitempty"`
	NamePrefix              string `xml:"NamePrefix,attr,omitempty"`
	BaseName                string `xml:"BaseName,attr,omitempty"`
	ShowMasterItems         string `xml:"ShowMasterItems,attr,omitempty"`
	PageCount               string `xml:"PageCount,attr,omitempty"`
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr"`
	PrimaryTextFrame        string `xml:"PrimaryTextFrame,attr,omitempty"`
	ItemTransform           string `xml:"ItemTransform,attr,omitempty"`

	// Items es la secuencia de elementos de página en Orden_Documental.
	Items []PageItem `xml:"-"`

	// childOrder recuerda la secuencia de hijos leída, para reproducirla al emitir.
	childOrder xmlutil.ChildOrder

	// Elementos hijo
	FlattenerPreference *FlattenerPreference `xml:"FlattenerPreference,omitempty"`
	Pages               []Page               `xml:"Page,omitempty"`
	textFrames          []SpreadTextFrame
	rectangles          []Rectangle
	images              []Image
	ovals               []Oval
	polygons            []Polygon
	graphicLines        []GraphicLine
	groups              []Group

	// OtherAttrs recoge los atributos que este tipo todavía no declara.
	OtherAttrs    []xml.Attr             `xml:",any,attr"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// --- Accesores de MasterSpreadElement ---

// TextFrames retorna punteros a los text frames, en orden documental.
func (mse *MasterSpreadElement) TextFrames() []*SpreadTextFrame {
	result := make([]*SpreadTextFrame, 0, len(mse.textFrames))
	for i := range mse.textFrames {
		result = append(result, &mse.textFrames[i])
	}
	return result
}

// Rectangles retorna punteros a los rectángulos, en orden documental.
func (mse *MasterSpreadElement) Rectangles() []*Rectangle {
	result := make([]*Rectangle, 0, len(mse.rectangles))
	for i := range mse.rectangles {
		result = append(result, &mse.rectangles[i])
	}
	return result
}

// Images retorna punteros a las imágenes, en orden documental.
func (mse *MasterSpreadElement) Images() []*Image {
	result := make([]*Image, 0, len(mse.images))
	for i := range mse.images {
		result = append(result, &mse.images[i])
	}
	return result
}

// Ovals retorna punteros a las elipses, en orden documental.
func (mse *MasterSpreadElement) Ovals() []*Oval {
	result := make([]*Oval, 0, len(mse.ovals))
	for i := range mse.ovals {
		result = append(result, &mse.ovals[i])
	}
	return result
}

// Polygons retorna punteros a los polígonos, en orden documental.
func (mse *MasterSpreadElement) Polygons() []*Polygon {
	result := make([]*Polygon, 0, len(mse.polygons))
	for i := range mse.polygons {
		result = append(result, &mse.polygons[i])
	}
	return result
}

// GraphicLines retorna punteros a las líneas gráficas, en orden documental.
func (mse *MasterSpreadElement) GraphicLines() []*GraphicLine {
	result := make([]*GraphicLine, 0, len(mse.graphicLines))
	for i := range mse.graphicLines {
		result = append(result, &mse.graphicLines[i])
	}
	return result
}

// Groups retorna punteros a los grupos, en orden documental.
func (mse *MasterSpreadElement) Groups() []*Group {
	result := make([]*Group, 0, len(mse.groups))
	for i := range mse.groups {
		result = append(result, &mse.groups[i])
	}
	return result
}

// Append agrega un elemento de página al final del master spread, en Orden_Documental.
// Misma semántica que SpreadElement.Append.
func (mse *MasterSpreadElement) Append(item PageItem) (PageItem, error) {
	if item == nil {
		return nil, common.Errorf("spread", "append page item", "", "el elemento de página es nil")
	}

	switch v := item.(type) {
	case *SpreadTextFrame:
		mse.textFrames = append(mse.textFrames, *v)
	case *Rectangle:
		mse.rectangles = append(mse.rectangles, *v)
	case *Image:
		mse.images = append(mse.images, *v)
	case *Oval:
		mse.ovals = append(mse.ovals, *v)
	case *Polygon:
		mse.polygons = append(mse.polygons, *v)
	case *GraphicLine:
		mse.graphicLines = append(mse.graphicLines, *v)
	case *Group:
		mse.groups = append(mse.groups, *v)
	default:
		return nil, common.Errorf("spread", "append page item", item.GetSelf(),
			"MasterSpreadElement no tiene campo para un <"+item.xmlTag()+"> a este nivel")
	}

	mse.childOrder.Record(item.xmlTag())
	mse.rebuildItems()

	if len(mse.Items) == 0 {
		return nil, common.Errorf("spread", "append page item", item.GetSelf(),
			"el elemento no quedó registrado en Items")
	}
	return mse.Items[len(mse.Items)-1], nil
}

// ItemTags devuelve la secuencia de nombres de elemento de Items, en Orden_Documental.
func (mse *MasterSpreadElement) ItemTags() []string {
	tags := make([]string, 0, len(mse.Items))
	for _, it := range mse.Items {
		tags = append(tags, it.xmlTag())
	}
	return tags
}

// --- Métodos de mutación directa para compatibilidad con pkg/idml ---
// Estos métodos permiten a los consumidores dentro del módulo manipular los campos
// sin exportar sin pasar por Append (que registra orden). Son para operaciones de
// edición tipo Remove/Set que necesitan acceso posicional al slice interno.

// RemoveTextFrameAt elimina el text frame en la posición indicada del slice interno.
func (se *SpreadElement) RemoveTextFrameAt(i int) {
	se.textFrames = append(se.textFrames[:i], se.textFrames[i+1:]...)
	se.rebuildItems()
}

// RemoveRectangleAt elimina el rectángulo en la posición indicada del slice interno.
func (se *SpreadElement) RemoveRectangleAt(i int) {
	se.rectangles = append(se.rectangles[:i], se.rectangles[i+1:]...)
	se.rebuildItems()
}

// AddTextFrame agrega un text frame al spread (sin registro de orden, para compatibilidad).
func (se *SpreadElement) AddTextFrame(tf SpreadTextFrame) {
	se.textFrames = append(se.textFrames, tf)
	se.rebuildItems()
}

// AddRectangle agrega un rectángulo al spread (sin registro de orden, para compatibilidad).
func (se *SpreadElement) AddRectangle(rect Rectangle) {
	se.rectangles = append(se.rectangles, rect)
	se.rebuildItems()
}

// SetTextFrameAt reemplaza el text frame en la posición indicada.
func (se *SpreadElement) SetTextFrameAt(i int, tf SpreadTextFrame) {
	se.textFrames[i] = tf
	se.rebuildItems()
}

// SetRectangleAt reemplaza el rectángulo en la posición indicada.
func (se *SpreadElement) SetRectangleAt(i int, rect Rectangle) {
	se.rectangles[i] = rect
	se.rebuildItems()
}
