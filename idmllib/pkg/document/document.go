// Package document provee tipos para la estructura del documento InDesign (designmap.xml).
//
// Este paquete contiene el tipo Document y todos los tipos relacionados para representar
// el manifiesto principal del documento, incluyendo capas, secciones, idiomas, grillas,
// grupos de colores, variables de texto y configuraciones a nivel de documento.
//
// Para archivos IDMS (snippets), el tipo Document también soporta recursos inline
// y contenido que normalmente estaría en archivos separados en formato IDML.
package document

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Document representa el elemento raíz de designmap.xml.
// Es el archivo de manifiesto principal de un paquete IDML.
//
// Implementación Fase 2: Parseo completo de todos los atributos del Document y
// elementos hijo principales manteniendo compatibilidad hacia adelante.
type Document struct {
	// XMLName captura el nombre del elemento. Nota: Document no tiene namespace propio,
	// pero declara el namespace idPkg para los elementos hijo.
	XMLName xml.Name `xml:"Document"`

	// Xmlns define el prefijo de namespace idPkg usado por los elementos hijo.
	// Ejemplo: xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"
	Xmlns string `xml:"xmlns:idPkg,attr"`

	// Atributos de identidad principales
	DOMVersion string `xml:"DOMVersion,attr"`     // Versión del DOM de InDesign (ej: "20.4")
	Self       string `xml:"Self,attr"`           // Identificador único (usualmente "d")
	Name       string `xml:"Name,attr,omitempty"` // Nombre del documento

	// Gestión de stories
	StoryList string `xml:"StoryList,attr,omitempty"` // Lista de IDs de stories separados por espacios

	// Layout y posicionamiento
	ZeroPoint   string `xml:"ZeroPoint,attr,omitempty"`   // Punto cero del documento (ej: "0 0")
	ActiveLayer string `xml:"ActiveLayer,attr,omitempty"` // ID de la capa actualmente activa

	// Gestión de color
	CMYKProfile         string `xml:"CMYKProfile,attr,omitempty"`         // Perfil de color CMYK
	RGBProfile          string `xml:"RGBProfile,attr,omitempty"`          // Perfil de color RGB
	SolidColorIntent    string `xml:"SolidColorIntent,attr,omitempty"`    // Intención de renderizado para colores sólidos
	AfterBlendingIntent string `xml:"AfterBlendingIntent,attr,omitempty"` // Intención de renderizado post-mezcla
	DefaultImageIntent  string `xml:"DefaultImageIntent,attr,omitempty"`  // Intención de renderizado de imágenes por defecto
	RGBPolicy           string `xml:"RGBPolicy,attr,omitempty"`           // Política de color RGB
	CMYKPolicy          string `xml:"CMYKPolicy,attr,omitempty"`          // Política de color CMYK
	AccurateLABSpots    string `xml:"AccurateLABSpots,attr,omitempty"`    // Precisión de colores spot LAB ("true"/"false")

	// Configuración MathML (para composición tipográfica matemática)
	AppliedMathMLFontSize    string `xml:"AppliedMathMLFontSize,attr,omitempty"`    // Tamaño de fuente MathML
	AppliedMathMLRgbColor    string `xml:"AppliedMathMLRgbColor,attr,omitempty"`    // Color RGB para MathML
	PreferMathMLInEpubExport string `xml:"PreferMathMLInEpubExport,attr,omitempty"` // Usar MathML en exportación EPUB
	TintValue                string `xml:"TintValue,attr,omitempty"`                // Valor de tinta para MathML

	// Elementos hijo (Fase 2 - Paso 2: Properties)
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Paso 3: Idiomas
	Languages []Language `xml:"Language,omitempty"`

	// Paso 4: Referencias a recursos (namespace idPkg)
	GraphicResource     *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`
	FontsResource       *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Fonts,omitempty"`
	StylesResource      *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Styles,omitempty"`
	PreferencesResource *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Preferences,omitempty"`
	TagsResource        *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Tags,omitempty"`

	// Paso 4b: Referencias a recursos de contenido (namespace idPkg)
	// Apuntan al contenido real del documento (spreads, stories, etc.)
	MasterSpreads []ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging MasterSpread,omitempty"`
	Spreads       []ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Spread,omitempty"`
	Stories       []ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Story,omitempty"`
	BackingStory  *ResourceRef  `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging BackingStory,omitempty"`

	// Paso 5: Elementos principales de layout
	Layers []Layer `xml:"Layer,omitempty"`

	// Paso 6: Numeración y grillas
	NumberingLists []NumberingList `xml:"NumberingList,omitempty"`
	NamedGrids     []NamedGrid     `xml:"NamedGrid,omitempty"`

	// Paso 7: Estructura del documento
	Sections      []Section      `xml:"Section,omitempty"`
	DocumentUsers []DocumentUser `xml:"DocumentUser,omitempty"`

	// Paso 8: Colores, viñetas y asignaciones
	ColorGroups []ColorGroup `xml:"ColorGroup,omitempty"`
	ABullets    []ABullet    `xml:"ABullet,omitempty"`
	Assignments []Assignment `xml:"Assignment,omitempty"`

	// Paso 9: Variables de texto
	TextVariables []TextVariable `xml:"TextVariable,omitempty"`

	// ========================================================================
	// Fase 4.3: Recursos inline para IDMS
	// ========================================================================
	// En archivos IDMS (snippets), los recursos están embebidos inline en el Document
	// en lugar de estar en archivos separados. Estos campos permiten que el struct
	// Document contenga tanto referencias IDML (arriba) como contenido inline IDMS.
	//
	// En archivos IDML, estos campos suelen estar vacíos (se usan ResourceRef en su lugar).
	// En archivos IDMS, estos campos contienen los datos reales de los recursos.

	// Colores y muestras inline (en lugar de GraphicResource)
	// Nota: Colors, Swatches, StrokeStyles usan tipos de resources (Fase 5a completa)
	Colors       []resources.Color       `xml:"Color,omitempty"`
	Swatches     []resources.Swatch      `xml:"Swatch,omitempty"`
	StrokeStyles []resources.StrokeStyle `xml:"StrokeStyle,omitempty"`

	// Grupos de estilos inline (en lugar de StylesResource)
	// Nota: Los grupos de estilos usan tipos de resources (Fase 5c completa)
	RootCharacterStyleGroup *resources.CharacterStyleGroup `xml:"RootCharacterStyleGroup,omitempty"`
	RootParagraphStyleGroup *resources.ParagraphStyleGroup `xml:"RootParagraphStyleGroup,omitempty"`
	RootObjectStyleGroup    *resources.ObjectStyleGroup    `xml:"RootObjectStyleGroup,omitempty"`

	// Contenido inline (en lugar de ResourceRefs de Spreads/Stories)
	// Nota: InlineSpreads usa spread.SpreadElement (Fase 3 completa)
	// Nota: InlineStories usa story.StoryElement (Fase 4 completa)
	InlineSpreads []spread.SpreadElement `xml:"Spread,omitempty"`
	InlineStories []story.StoryElement   `xml:"Story,omitempty"`

	// Elementos requeridos para compatibilidad con InDesign
	TinDocumentDataObject              *TinDocumentDataObject              `xml:"TinDocumentDataObject,omitempty"`
	TransparencyDefaultContainerObject *TransparencyDefaultContainerObject `xml:"TransparencyDefaultContainerObject,omitempty"`

	// Comodín para todos los demás elementos hijo aún no modelados explícitamente.
	// Incluye: KinsokuTable, MojikumiTable, CrossReferenceFormat,
	// ConditionalTextPreference, EndnoteOption, WatermarkPreference, IndexingSortOption,
	// LinkedStoryOption, LinkedPageItemOption, y muchos más.
	// A medida que se agrega soporte explícito para más elementos, se mueven de OtherElements
	// a campos dedicados arriba.
	OtherElements []common.RawXMLElement `xml:",any"`

	// childOrder recuerda en qué orden venían los hijos en el XML de entrada.
	//
	// Hace falta porque los hijos de <Document> viven repartidos en los campos por
	// tipo de arriba, y emitirlos en el orden de esos campos reagrupa el documento.
	// InDesign intercala Layer, Section, TextVariable, ABullet y las referencias
	// idPkg sin agruparlos por tipo: en el designmap del Documento_Referencia son 118
	// hijos, y sin este registro salen todos reordenados.
	//
	// No es un contenedor. El contenido de cada hijo sigue viviendo en su campo por
	// tipo, que es su fuente de verdad; aquí solo está la secuencia. Va sin exportar
	// porque solo el parseo lo escribe.
	childOrder xmlutil.ChildOrder
}

// Clases de hijo de <Document>, una por campo del struct. El registro de orden
// guarda estas cadenas, así que identifican el campo y no la etiqueta XML: por eso
// hay dos entradas para Spread y dos para Story, que existen como referencia idPkg
// y como contenido inline de un IDMS, con la misma etiqueta y campos distintos.
const (
	childProperties          = "Properties"
	childLanguage            = "Language"
	childRefGraphic          = "idPkg:Graphic"
	childRefFonts            = "idPkg:Fonts"
	childRefStyles           = "idPkg:Styles"
	childRefPreferences      = "idPkg:Preferences"
	childRefTags             = "idPkg:Tags"
	childRefMasterSpread     = "idPkg:MasterSpread"
	childRefSpread           = "idPkg:Spread"
	childRefStory            = "idPkg:Story"
	childRefBackingStory     = "idPkg:BackingStory"
	childLayer               = "Layer"
	childNumberingList       = "NumberingList"
	childNamedGrid           = "NamedGrid"
	childSection             = "Section"
	childDocumentUser        = "DocumentUser"
	childColorGroup          = "ColorGroup"
	childABullet             = "ABullet"
	childAssignment          = "Assignment"
	childTextVariable        = "TextVariable"
	childColor               = "Color"
	childSwatch              = "Swatch"
	childStrokeStyle         = "StrokeStyle"
	childRootCharStyleGroup  = "RootCharacterStyleGroup"
	childRootParaStyleGroup  = "RootParagraphStyleGroup"
	childRootObjStyleGroup   = "RootObjectStyleGroup"
	childTinDocumentData     = "TinDocumentDataObject"
	childTransparencyDefault = "TransparencyDefaultContainerObject"
	childInlineSpread        = "Spread"
	childInlineStory         = "Story"
	childOther               = "OtherElement"
)

// documentChildOrder es el orden en que están declarados los campos del struct.
// Se usa para los hijos que el registro de orden no menciona: los de un documento
// construido desde cero, y los que se agregan después de parsear. Reproduce el
// orden que tenía marshalChildren antes de existir el registro, de modo que la
// salida de un documento sin registro no cambia.
//
// Tiene que nombrar todas las clases de arriba. Lo comprueba un test.
var documentChildOrder = []string{
	childProperties,
	childLanguage,
	childRefGraphic,
	childRefFonts,
	childRefStyles,
	childRefPreferences,
	childRefTags,
	childRefMasterSpread,
	childRefSpread,
	childRefStory,
	childRefBackingStory,
	childLayer,
	childNumberingList,
	childNamedGrid,
	childSection,
	childDocumentUser,
	childColorGroup,
	childABullet,
	childAssignment,
	childTextVariable,
	childColor,
	childSwatch,
	childStrokeStyle,
	childRootCharStyleGroup,
	childRootParaStyleGroup,
	childRootObjStyleGroup,
	childTinDocumentData,
	childTransparencyDefault,
	childInlineSpread,
	childInlineStory,
	childOther,
}

// Language representa una definición de idioma en el documento.
// Los idiomas definen configuraciones de localización para el texto.
type Language struct {
	XMLName xml.Name `xml:"Language"`

	// Identificación
	Self string `xml:"Self,attr"`         // Identificador único (ej: "Language/$ID/English%3a UK")
	Name string `xml:"Name,attr"`         // Nombre de visualización (ej: "$ID/English: UK")
	Id   string `xml:"Id,attr,omitempty"` // ID numérico del idioma

	// Componentes del idioma
	PrimaryLanguageName string `xml:"PrimaryLanguageName,attr,omitempty"` // Idioma principal (ej: "$ID/English")
	SublanguageName     string `xml:"SublanguageName,attr,omitempty"`     // Subidioma/región (ej: "$ID/UK")

	// Configuración tipográfica
	SingleQuotes string `xml:"SingleQuotes,attr,omitempty"` // Caracteres de comilla simple (ej: "''")
	DoubleQuotes string `xml:"DoubleQuotes,attr,omitempty"` // Caracteres de comilla doble (ej: """")

	// Proveedores de procesamiento
	HyphenationVendor string `xml:"HyphenationVendor,attr,omitempty"` // Proveedor de separación silábica (ej: "Proximity", "Hunspell")
	SpellingVendor    string `xml:"SpellingVendor,attr,omitempty"`    // Proveedor de corrección ortográfica
}

// ResourceRef representa un elemento de referencia a recurso idPkg:*.
// Apuntan a archivos XML externos dentro del paquete IDML.
//
// Ejemplos:
//
//	<idPkg:Graphic src="Resources/Graphic.xml" />
//	<idPkg:Fonts src="Resources/Fonts.xml" />
//	<idPkg:Styles src="Resources/Styles.xml" />
type ResourceRef struct {
	XMLName xml.Name // Se establece con el nombre de elemento con namespace
	Src     string   `xml:"src,attr"`
}

// Layer representa una capa del documento para organizar contenido.
// Las capas controlan la visibilidad, bloqueo e impresión del contenido.
type Layer struct {
	XMLName xml.Name `xml:"Layer"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "uba")
	Name string `xml:"Name,attr"` // Nombre de visualización (ej: "Editorial")

	// Visibilidad e interacción
	Visible    string `xml:"Visible,attr,omitempty"`    // Visibilidad de la capa ("true"/"false")
	Locked     string `xml:"Locked,attr,omitempty"`     // Bloquear capa ("true"/"false")
	IgnoreWrap string `xml:"IgnoreWrap,attr,omitempty"` // Ignorar contorno de texto ("true"/"false")

	// Guías
	ShowGuides string `xml:"ShowGuides,attr,omitempty"` // Mostrar guías ("true"/"false")
	LockGuides string `xml:"LockGuides,attr,omitempty"` // Bloquear guías ("true"/"false")

	// Comportamiento de la capa
	UI         string `xml:"UI,attr,omitempty"`         // Mostrar en UI ("true"/"false")
	Expendable string `xml:"Expendable,attr,omitempty"` // Puede eliminarse ("true"/"false")
	Printable  string `xml:"Printable,attr,omitempty"`  // Imprimir capa ("true"/"false")

	// Properties puede contener LayerColor y otras configuraciones
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Comodín para otros hijos de Layer
	OtherElements []common.RawXMLElement `xml:",any"`
}

// NumberingList representa una definición de lista numerada.
// Controla la numeración automática de párrafos entre stories y documentos.
type NumberingList struct {
	XMLName xml.Name `xml:"NumberingList"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "NumberingList/$ID/[Default]")
	Name string `xml:"Name,attr"` // Nombre de visualización (ej: "$ID/[Default]")

	// Comportamiento de numeración
	ContinueNumbersAcrossStories   string `xml:"ContinueNumbersAcrossStories,attr,omitempty"`   // Continuar entre stories ("true"/"false")
	ContinueNumbersAcrossDocuments string `xml:"ContinueNumbersAcrossDocuments,attr,omitempty"` // Continuar entre documentos ("true"/"false")

	// Comodín para atributos o elementos hijo futuros
	OtherElements []common.RawXMLElement `xml:",any"`
}

// NamedGrid representa una definición de grilla de layout con nombre.
// Se usa principalmente para tipografía CJK y sistemas de layout basados en grilla.
type NamedGrid struct {
	XMLName xml.Name `xml:"NamedGrid"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "NamedGrid/$ID/[Page Grid]")
	Name string `xml:"Name,attr"` // Nombre de visualización (ej: "$ID/[Page Grid]")

	// Configuración de grilla
	GridDataInformation *common.GridDataInformation `xml:"GridDataInformation,omitempty"`

	// Comodín para otros hijos de NamedGrid
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Section representa una sección del documento con su propia numeración de páginas.
// Las secciones permiten diferentes esquemas de numeración, prefijos y estilos dentro de un documento.
type Section struct {
	XMLName xml.Name `xml:"Section"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "ub4")
	Name string `xml:"Name,attr"` // Nombre de la sección (ej: "A")

	// Rango de páginas
	Length    string `xml:"Length,attr,omitempty"`    // Número de páginas en la sección
	PageStart string `xml:"PageStart,attr,omitempty"` // ID de la página inicial

	// Configuración de numeración
	ContinueNumbering    string `xml:"ContinueNumbering,attr,omitempty"`    // Continuar desde la sección anterior ("true"/"false")
	IncludeSectionPrefix string `xml:"IncludeSectionPrefix,attr,omitempty"` // Incluir prefijo en números de página ("true"/"false")
	PageNumberStart      string `xml:"PageNumberStart,attr,omitempty"`      // Número de página inicial
	SectionPrefix        string `xml:"SectionPrefix,attr,omitempty"`        // Prefijo para números de página (ej: "A")
	Marker               string `xml:"Marker,attr,omitempty"`               // Marcador de sección

	// Layout alternativo (para publicación digital)
	AlternateLayout       string `xml:"AlternateLayout,attr,omitempty"`       // Nombre del layout alternativo
	AlternateLayoutLength string `xml:"AlternateLayoutLength,attr,omitempty"` // Longitud en el layout alternativo

	// Properties puede contener PageNumberStyle y Label
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Comodín para otros hijos de Section
	OtherElements []common.RawXMLElement `xml:",any"`
}

// DocumentUser representa un usuario que ha trabajado en el documento.
// Registra la colaboración y la autoría de los cambios rastreados.
type DocumentUser struct {
	XMLName xml.Name `xml:"DocumentUser"`

	// Identificación
	Self     string `xml:"Self,attr"`     // Identificador único (ej: "dDocumentUser0")
	UserName string `xml:"UserName,attr"` // Nombre del usuario

	// Properties puede contener UserColor y otras configuraciones
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Comodín para otros hijos de DocumentUser
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ColorGroup representa un grupo de muestras de color para organización.
// Ayuda a organizar y gestionar las muestras de color en el documento.
type ColorGroup struct {
	XMLName xml.Name `xml:"ColorGroup"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "ColorGroup/[Root Color Group]")
	Name string `xml:"Name,attr"` // Nombre del grupo (ej: "[Root Color Group]")

	// Indicador de grupo raíz
	IsRootColorGroup string `xml:"IsRootColorGroup,attr,omitempty"` // Es el grupo raíz ("true"/"false")

	// Muestras de color en este grupo
	ColorGroupSwatches []ColorGroupSwatch `xml:"ColorGroupSwatch,omitempty"`

	// Comodín para otros hijos de ColorGroup
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ColorGroupSwatch representa una referencia a una muestra de color dentro de un grupo de colores.
type ColorGroupSwatch struct {
	XMLName xml.Name `xml:"ColorGroupSwatch"`

	// Identificación
	Self          string `xml:"Self,attr"`          // Identificador único
	SwatchItemRef string `xml:"SwatchItemRef,attr"` // Referencia a la muestra (ej: "Color/Black")

	// Comodín para otros hijos de ColorGroupSwatch
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ABullet representa una definición de carácter de viñeta.
// Define el carácter y la fuente usados para viñetas en listas.
type ABullet struct {
	XMLName xml.Name `xml:"ABullet"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "dABullet0")

	// Definición del carácter
	CharacterType  string `xml:"CharacterType,attr,omitempty"`  // Tipo de carácter ("UnicodeOnly", "UnicodeWithFont")
	CharacterValue string `xml:"CharacterValue,attr,omitempty"` // Valor Unicode (ej: "8226" para punto de viñeta)

	// Properties puede contener BulletsFont y BulletsFontStyle
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Comodín para otros hijos de ABullet
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Assignment representa una asignación InCopy para edición colaborativa.
// Define qué contenido está asignado a usuarios específicos para editar.
type Assignment struct {
	XMLName xml.Name `xml:"Assignment"`

	// Identificación
	Self     string `xml:"Self,attr"`     // Identificador único (ej: "uc9")
	Name     string `xml:"Name,attr"`     // Nombre de la asignación
	UserName string `xml:"UserName,attr"` // Usuario asignado a este contenido

	// Configuración de exportación y empaquetado
	ExportOptions           string `xml:"ExportOptions,attr,omitempty"`           // Qué exportar (ej: "AssignedSpreads")
	IncludeLinksWhenPackage string `xml:"IncludeLinksWhenPackage,attr,omitempty"` // Incluir archivos vinculados ("true"/"false")
	FilePath                string `xml:"FilePath,attr,omitempty"`                // Ruta del archivo de asignación

	// Properties puede contener FrameColor y otras configuraciones
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Comodín para otros hijos de Assignment
	OtherElements []common.RawXMLElement `xml:",any"`
}

// TextVariable representa una variable de texto dinámico en el documento.
// Se actualiza automáticamente según las propiedades o el contexto del documento
// (ej: números de página, fechas, nombres de archivo, encabezados corrientes).
type TextVariable struct {
	XMLName xml.Name `xml:"TextVariable"`

	// Identificación
	Self string `xml:"Self,attr"` // Identificador único (ej: "dTextVariablenChapter Number")
	Name string `xml:"Name,attr"` // Nombre de visualización (ej: "Chapter Number")

	// Tipo de variable
	VariableType string `xml:"VariableType,attr,omitempty"` // Tipo de variable (ej: "ChapterNumberType", "CreationDateType")

	// Preferencias de variable (distintos tipos según el tipo de variable)
	// Nota: Solo uno de estos estará presente según VariableType
	ChapterNumberPreference       *ChapterNumberVariablePreference   `xml:"ChapterNumberVariablePreference,omitempty"`
	DatePreference                *DateVariablePreference            `xml:"DateVariablePreference,omitempty"`
	FileNamePreference            *FileNameVariablePreference        `xml:"FileNameVariablePreference,omitempty"`
	CaptionMetadataPreference     *CaptionMetadataVariablePreference `xml:"CaptionMetadataVariablePreference,omitempty"`
	PageNumberPreference          *PageNumberVariablePreference      `xml:"PageNumberVariablePreference,omitempty"`
	MatchParagraphStylePreference *MatchParagraphStylePreference     `xml:"MatchParagraphStylePreference,omitempty"`

	// Comodín para otros hijos de TextVariable o tipos de preferencia desconocidos
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ChapterNumberVariablePreference contiene configuraciones para variables de número de capítulo.
type ChapterNumberVariablePreference struct {
	XMLName    xml.Name `xml:"ChapterNumberVariablePreference"`
	TextBefore string   `xml:"TextBefore,attr,omitempty"` // Texto antes del número
	Format     string   `xml:"Format,attr,omitempty"`     // Formato del número (ej: "Current")
	TextAfter  string   `xml:"TextAfter,attr,omitempty"`  // Texto después del número
}

// DateVariablePreference contiene configuraciones para variables de fecha.
type DateVariablePreference struct {
	XMLName    xml.Name `xml:"DateVariablePreference"`
	TextBefore string   `xml:"TextBefore,attr,omitempty"` // Texto antes de la fecha
	Format     string   `xml:"Format,attr,omitempty"`     // Formato de fecha (ej: "dd/MM/yy", "d MMMM yyyy h:mm aa")
	TextAfter  string   `xml:"TextAfter,attr,omitempty"`  // Texto después de la fecha
}

// FileNameVariablePreference contiene configuraciones para variables de nombre de archivo.
type FileNameVariablePreference struct {
	XMLName          xml.Name `xml:"FileNameVariablePreference"`
	TextBefore       string   `xml:"TextBefore,attr,omitempty"`       // Texto antes del nombre de archivo
	IncludePath      string   `xml:"IncludePath,attr,omitempty"`      // Incluir ruta del archivo ("true"/"false")
	IncludeExtension string   `xml:"IncludeExtension,attr,omitempty"` // Incluir extensión del archivo ("true"/"false")
	TextAfter        string   `xml:"TextAfter,attr,omitempty"`        // Texto después del nombre de archivo
}

// CaptionMetadataVariablePreference contiene configuraciones para variables de metadatos de caption.
type CaptionMetadataVariablePreference struct {
	XMLName              xml.Name `xml:"CaptionMetadataVariablePreference"`
	TextBefore           string   `xml:"TextBefore,attr,omitempty"`           // Texto antes de los metadatos
	MetadataProviderName string   `xml:"MetadataProviderName,attr,omitempty"` // Fuente de metadatos (ej: "$ID/#LinkInfoNameStr")
	TextAfter            string   `xml:"TextAfter,attr,omitempty"`            // Texto después de los metadatos
}

// PageNumberVariablePreference contiene configuraciones para variables de número de página.
type PageNumberVariablePreference struct {
	XMLName    xml.Name `xml:"PageNumberVariablePreference"`
	TextBefore string   `xml:"TextBefore,attr,omitempty"` // Texto antes del número de página
	Format     string   `xml:"Format,attr,omitempty"`     // Formato del número (ej: "Current")
	TextAfter  string   `xml:"TextAfter,attr,omitempty"`  // Texto después del número de página
	Scope      string   `xml:"Scope,attr,omitempty"`      // Alcance (ej: "SectionScope")
}

// MatchParagraphStylePreference contiene configuraciones para variables de encabezado corriente.
type MatchParagraphStylePreference struct {
	XMLName               xml.Name `xml:"MatchParagraphStylePreference"`
	TextBefore            string   `xml:"TextBefore,attr,omitempty"`            // Texto antes del texto coincidente
	TextAfter             string   `xml:"TextAfter,attr,omitempty"`             // Texto después del texto coincidente
	AppliedParagraphStyle string   `xml:"AppliedParagraphStyle,attr,omitempty"` // Estilo de párrafo a coincidir
	SearchStrategy        string   `xml:"SearchStrategy,attr,omitempty"`        // Estrategia de búsqueda (ej: "FirstOnPage")
	ChangeCase            string   `xml:"ChangeCase,attr,omitempty"`            // Transformación de mayúsculas (ej: "None")
	DeleteEndPunctuation  string   `xml:"DeleteEndPunctuation,attr,omitempty"`  // Eliminar puntuación final ("true"/"false")
}

// ============================================================================
// Fase 4.3: Tipos específicos para IDMS
// ============================================================================

// TinDocumentDataObject representa datos internos del documento de InDesign.
// Este elemento es requerido para compatibilidad con InDesign en archivos IDMS.
// Típicamente contiene configuraciones de espacio de color e intención de renderizado.
type TinDocumentDataObject struct {
	XMLName xml.Name `xml:"TinDocumentDataObject,omitempty"`
	// Contiene datos internos de InDesign - preservado tal cual para compatibilidad
	OtherElements []common.RawXMLElement `xml:",any"`
}

// TransparencyDefaultContainerObject contiene configuraciones de transparencia por defecto.
// Este elemento es requerido para compatibilidad con InDesign en archivos IDMS.
// Define la opacidad y el modo de mezcla por defecto para el documento.
type TransparencyDefaultContainerObject struct {
	XMLName xml.Name `xml:"TransparencyDefaultContainerObject,omitempty"`
	// Contiene valores por defecto de transparencia - preservado tal cual para compatibilidad
	OtherElements []common.RawXMLElement `xml:",any"`
}
