package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// FontsFile representa el archivo Resources/Fonts.xml que contiene las definiciones de fuentes.
//
// El elemento raíz es <idPkg:Fonts> con el namespace idPkg.
type FontsFile struct {
	// XMLName no se establece directamente - se maneja manualmente en MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion es la versión del DOM de InDesign (ej., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// FontFamilies contiene grupos de familias tipográficas con sus definiciones de fuentes individuales
	FontFamilies []FontFamily `xml:"FontFamily,omitempty"`

	// CompositeFonts define configuraciones de fuentes compuestas (principalmente para tipografía asiática)
	CompositeFonts []CompositeFont `xml:"CompositeFont,omitempty"`

	// Captura todos los demás elementos no modelados explícitamente
	OtherElements []common.RawXMLElement `xml:",any"`
}

// FontFamily representa un grupo de familia tipográfica (ej., "Minion Pro", "Myriad Pro").
// Contiene múltiples entradas Font para distintos pesos y estilos.
type FontFamily struct {
	Self  string `xml:"Self,attr"`
	Name  string `xml:"Name,attr"`
	Fonts []Font `xml:"Font,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// Font representa una definición de fuente individual dentro de una familia tipográfica.
// Contiene métricas detalladas, información de estilo y estado de instalación.
type Font struct {
	Self                string `xml:"Self,attr"`
	FontFamily          string `xml:"FontFamily,attr"`
	Name                string `xml:"Name,attr"`
	PostScriptName      string `xml:"PostScriptName,attr"`
	Status              string `xml:"Status,attr"`        // "Installed", "Substituted", "NotAvailable"
	FontStyleName       string `xml:"FontStyleName,attr"` // "Regular", "Bold", "Italic", etc.
	FontType            string `xml:"FontType,attr"`      // "OpenTypeCFF", "OpenTypeCID", "TrueType", etc.
	WritingScript       string `xml:"WritingScript,attr"` // "0" para latín, "1" para CJK, etc.
	FullName            string `xml:"FullName,attr"`
	FullNameNative      string `xml:"FullNameNative,attr"`
	FontStyleNameNative string `xml:"FontStyleNameNative,attr"`
	PlatformName        string `xml:"PlatformName,attr"`
	Version             string `xml:"Version,attr"`
	TypekitID           string `xml:"TypekitID,attr,omitempty"` // ID de Adobe Typekit/Fonts

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// CompositeFont representa una definición de fuente compuesta (principalmente para tipografía CJK).
// Permite mezclar distintas fuentes para diferentes rangos de caracteres.
type CompositeFont struct {
	Self                 string                 `xml:"Self,attr"`
	Name                 string                 `xml:"Name,attr"`
	CompositeFontEntries []CompositeFontEntry   `xml:"CompositeFontEntry,omitempty"`
	OtherElements        []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// CompositeFontEntry representa una entrada individual en una fuente compuesta.
// Define qué fuente usar para un rango de caracteres o sistema de escritura específico.
type CompositeFontEntry struct {
	Self             string             `xml:"Self,attr"`
	Name             string             `xml:"Name,attr"`                       // Nombre del rango de caracteres (ej., "$ID/Kanji", "$ID/Kana")
	FontStyle        string             `xml:"FontStyle,attr"`                  // Referencia al estilo de fuente
	RelativeSize     string             `xml:"RelativeSize,attr,omitempty"`     // Porcentaje de tamaño relativo
	HorizontalScale  string             `xml:"HorizontalScale,attr,omitempty"`  // Porcentaje de escala horizontal
	VerticalScale    string             `xml:"VerticalScale,attr,omitempty"`    // Porcentaje de escala vertical
	CustomCharacters string             `xml:"CustomCharacters,attr,omitempty"` // Lista de caracteres personalizados
	Locked           string             `xml:"Locked,attr,omitempty"`           // "true" o "false"
	ScaleOption      string             `xml:"ScaleOption,attr,omitempty"`      // "true" o "false"
	BaselineShift    string             `xml:"BaselineShift,attr,omitempty"`    // Valor de desplazamiento de línea base
	Properties       *common.Properties `xml:"Properties,omitempty"`            // Contiene <AppliedFont>

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}
