package story

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Story representa un archivo XML de Story de InDesign.
// Las stories contienen el contenido de texto con información de formato.
type Story struct {
	XMLName    xml.Name `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Story"`
	DOMVersion string   `xml:"DOMVersion,attr"`

	// El contenido real de la story
	StoryElement StoryElement `xml:"Story"`
}

// ExtractText retorna todo el contenido de texto de la story concatenado como un único string.
// Los saltos de línea (elementos <Br>) se convierten en caracteres de nueva línea.
// Es un método de conveniencia que navega la estructura de la story automáticamente.
func (s *Story) ExtractText() string {
	var buf strings.Builder
	for _, psr := range s.StoryElement.ParagraphStyleRanges {
		for _, csr := range psr.CharacterStyleRanges {
			for _, child := range csr.Children {
				if child.Content != nil {
					buf.WriteString(child.Content.Text)
				} else if child.Br != nil {
					buf.WriteString("\n")
				}
			}
		}
	}
	return buf.String()
}

// StoryElement representa el elemento principal Story que contiene todo el contenido.
type StoryElement struct {
	XMLName xml.Name `xml:"Story"`

	// Identidad
	Self string `xml:"Self,attr"`

	// Metadatos de la story
	UserText         string `xml:"UserText,attr,omitempty"`         // "true"/"false"
	IsEndnoteStory   string `xml:"IsEndnoteStory,attr,omitempty"`   // "true"/"false"
	AppliedTOCStyle  string `xml:"AppliedTOCStyle,attr,omitempty"`  // Referencia al estilo de TOC
	TrackChanges     string `xml:"TrackChanges,attr,omitempty"`     // "true"/"false"
	StoryTitle       string `xml:"StoryTitle,attr,omitempty"`       // Título de la story
	AppliedNamedGrid string `xml:"AppliedNamedGrid,attr,omitempty"` // Referencia a la grilla con nombre

	// Preferencias de la story
	StoryPreference *StoryPreference `xml:"StoryPreference,omitempty"`

	// Opciones de exportación InCopy
	InCopyExportOption *InCopyExportOption `xml:"InCopyExportOption,omitempty"`

	// Contenido - rangos de estilo de párrafo
	ParagraphStyleRanges []ParagraphStyleRange `xml:"ParagraphStyleRange"`

	// Comodín para elementos desconocidos
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// StoryPreference representa las preferencias a nivel de story.
type StoryPreference struct {
	XMLName xml.Name `xml:"StoryPreference"`

	// Alineación óptica de márgenes
	OpticalMarginAlignment string `xml:"OpticalMarginAlignment,attr,omitempty"` // "true"/"false"
	OpticalMarginSize      string `xml:"OpticalMarginSize,attr,omitempty"`      // Tamaño en puntos

	// Tipo y orientación del frame
	FrameType        string `xml:"FrameType,attr,omitempty"`        // "TextFrameType", etc.
	StoryOrientation string `xml:"StoryOrientation,attr,omitempty"` // "Horizontal"/"Vertical"
	StoryDirection   string `xml:"StoryDirection,attr,omitempty"`   // "LeftToRightDirection", etc.

	// Comodín para atributos o elementos desconocidos
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// InCopyExportOption representa la configuración de exportación InCopy para la story.
type InCopyExportOption struct {
	XMLName xml.Name `xml:"InCopyExportOption"`

	IncludeGraphicProxies string `xml:"IncludeGraphicProxies,attr,omitempty"` // "true"/"false"
	IncludeAllResources   string `xml:"IncludeAllResources,attr,omitempty"`   // "true"/"false"

	// Comodín para atributos o elementos desconocidos
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// ParagraphStyleRange representa un rango de párrafos con el mismo estilo de párrafo.
type ParagraphStyleRange struct {
	XMLName xml.Name `xml:"ParagraphStyleRange"`

	// Referencia al estilo de párrafo aplicado
	AppliedParagraphStyle string `xml:"AppliedParagraphStyle,attr"`

	// Rangos de estilo de carácter dentro de este párrafo
	CharacterStyleRanges []CharacterStyleRange `xml:"CharacterStyleRange"`

	// Comodín para elementos desconocidos
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Hoy son los
	// 10 atributos de formato de párrafo que el Documento_Referencia trae y el modelo
	// no declara: Justification, Hyphenation, HyphenationZone, FirstLineIndent,
	// LeftIndent, GridAlignment, BulletsAndNumberingListType, RuleAboveLineWeight,
	// RuleBelowLineWeight y SplitColumnInsideGutter.
	//
	// La Tarea 15 los declarará como campos tipados, que es lo que el constructor de
	// documentos necesita para **generarlos**. Para **preservarlos** basta este
	// comodín, y los que sigan sin declararse seguirán cayendo aquí.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// CharacterStyleRange representa un rango de caracteres con el mismo estilo de carácter.
// IMPORTANTE: Este struct usa marshaling custom para preservar el orden de los elementos Content y Br.
type CharacterStyleRange struct {
	XMLName xml.Name `xml:"CharacterStyleRange"`

	// Referencia al estilo de carácter aplicado
	AppliedCharacterStyle string `xml:"AppliedCharacterStyle,attr"`

	// Atributos comunes de formato de carácter
	// Son opcionales y preservan el formato de texto de InDesign
	HorizontalScale string `xml:"HorizontalScale,attr,omitempty"` // Porcentaje de escala horizontal
	Tracking        string `xml:"Tracking,attr,omitempty"`        // Valor de tracking/espaciado entre letras

	// Atributos de formato adicionales (comodín)
	OtherAttrs []xml.Attr `xml:"-"` // No usado por encoding/xml, manejado manualmente

	// Contenido mixto: elementos Content y Br en orden
	// Este campo almacena ambos tipos de elementos en el orden en que aparecen
	Children []CharacterChild `xml:"-"` // Serializado manualmente para preservar el orden
}

// CharacterChild representa ya sea un elemento Content o un elemento Br
type CharacterChild struct {
	Content *Content              // Si no es nil, es un elemento Content
	Br      *Br                   // Si no es nil, es un elemento Br
	Other   *common.RawXMLElement // Si no es nil, es un elemento desconocido
}

// Content representa el contenido de texto real.
type Content struct {
	XMLName xml.Name `xml:"Content"`

	// Text es el texto legible, ya sin escapes. Es lo que leen los consumidores y lo
	// que escriben los métodos de la API, y sigue siendo la fuente de verdad del
	// contenido.
	Text string `xml:",chardata"`

	// Raw es el contenido interno literal, tal como venía en el archivo. Existe porque
	// un Content puede llevar **instrucciones de proceso** intercaladas con el texto, y
	// esas no son datos de carácter: `Text` no puede representarlas.
	//
	// InDesign las usa para los marcadores de carácter especial. En el
	// Documento_Referencia hay 4 apariciones de `<?ACE 18?>`, y las cuatro con una
	// forma distinta: la instrucción sola, antes del texto, antes de un texto con
	// espacios finales, y después de un espacio.
	//
	// Solo se usa para re-emitir lo que se parseó. Ver MarshalXML.
	Raw string `xml:",innerxml"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// MarshalXML emite el Content conservando las instrucciones de proceso que traía, sin
// perder la seguridad del escapado.
//
// La decisión, y el motivo de que no sea «emitir siempre Raw»: si un llamador cambió
// `Text` después de parsear, `Raw` quedó obsoleto, y emitirlo descartaría ese cambio
// **en silencio**. Perder un marcador de InDesign es malo; escribir un texto que no es
// el que pidió el llamador es peor. Así que `Text` manda, y `Raw` solo se usa mientras
// siga siendo coherente con él.
func (c Content) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "Content"}
	start.Attr = c.OtherAttrs

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	tokens, texto, err := tokensDeRaw(c.Raw)
	if err == nil && texto == c.Text && tieneInstruccionDeProceso(tokens) {
		// El contenido parseado traía marcado que Text no puede representar, y Text no
		// se ha modificado desde entonces: se reemiten los tokens originales. El
		// encoder escapa los datos de carácter y deja pasar las instrucciones, así que
		// el escapado sigue siendo correcto.
		for _, tok := range tokens {
			if err := e.EncodeToken(tok); err != nil {
				return err
			}
		}
	} else if c.Text != "" {
		// Camino normal: el texto se emite escapado por el encoder.
		if err := e.EncodeToken(xml.CharData(c.Text)); err != nil {
			return err
		}
	}

	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}
	return e.Flush()
}

// tokensDeRaw decodifica el contenido literal en tokens y devuelve además el texto que
// resulta de concatenar sus datos de carácter, que es lo que debería coincidir con
// Text si nadie lo ha modificado.
func tokensDeRaw(raw string) (tokens []xml.Token, texto string, err error) {
	if raw == "" {
		return nil, "", nil
	}

	var sb strings.Builder
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, "", err
		}
		// xml.CopyToken es necesario: el decoder reutiliza el buffer de CharData entre
		// llamadas, así que guardar el token sin copiar deja basura.
		tok = xml.CopyToken(tok)
		if cd, ok := tok.(xml.CharData); ok {
			sb.Write(cd)
		}
		tokens = append(tokens, tok)
	}
	return tokens, sb.String(), nil
}

// tieneInstruccionDeProceso indica si los tokens llevan algo que `Text` no puede
// representar. Si no lo llevan, no hace falta el camino de re-emisión literal.
func tieneInstruccionDeProceso(tokens []xml.Token) bool {
	for _, tok := range tokens {
		if _, ok := tok.(xml.ProcInst); ok {
			return true
		}
	}
	return false
}

// Br representa un elemento de salto de línea.
type Br struct {
	XMLName xml.Name `xml:"Br"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// NewCharacterStyleRange crea un nuevo CharacterStyleRange con el estilo y contenido dados.
// Es un constructor de conveniencia para compatibilidad hacia atrás con código que usaba literales de struct.
// Si appliedStyle está vacío, usa el marcador de sin estilo por defecto.
func NewCharacterStyleRange(appliedStyle string, contents []Content) CharacterStyleRange {
	if appliedStyle == "" {
		appliedStyle = "CharacterStyle/$ID/[No character style]"
	}

	csr := CharacterStyleRange{
		XMLName:               xml.Name{Local: "CharacterStyleRange"},
		AppliedCharacterStyle: appliedStyle,
	}

	// Convertir el arreglo de Content a Children con elementos Br intercalados
	for _, content := range contents {
		csr.Children = append(csr.Children, CharacterChild{Content: &Content{
			XMLName: xml.Name{Local: "Content"},
			Text:    content.Text,
		}})
		csr.Children = append(csr.Children, CharacterChild{Br: &Br{XMLName: xml.Name{Local: "Br"}}})
	}

	return csr
}

// GetContent retorna todos los elementos Content en orden (para compatibilidad hacia atrás).
// Permite que el código existente acceda al contenido sin conocer la nueva estructura Children.
func (c *CharacterStyleRange) GetContent() []Content {
	var contents []Content
	for _, child := range c.Children {
		if child.Content != nil {
			contents = append(contents, *child.Content)
		}
	}
	return contents
}

// SetContent establece los elementos de contenido, reemplazando los hijos existentes con la nueva estructura Content + Br.
// Provee compatibilidad hacia atrás para código que construye CharacterStyleRanges programáticamente.
// Se agregan saltos de línea después de cada elemento Content.
func (c *CharacterStyleRange) SetContent(contents []Content) {
	c.Children = nil
	for _, content := range contents {
		// Agregar contenido
		c.Children = append(c.Children, CharacterChild{Content: &Content{
			XMLName: xml.Name{Local: "Content"},
			Text:    content.Text,
		}})
		// Agregar salto de línea después del contenido
		c.Children = append(c.Children, CharacterChild{Br: &Br{XMLName: xml.Name{Local: "Br"}}})
	}
}

// AddContent agrega un elemento Content seguido de un Br (para compatibilidad hacia atrás).
func (c *CharacterStyleRange) AddContent(text string) {
	c.Children = append(c.Children, CharacterChild{Content: &Content{
		XMLName: xml.Name{Local: "Content"},
		Text:    text,
	}})
	c.Children = append(c.Children, CharacterChild{Br: &Br{XMLName: xml.Name{Local: "Br"}}})
}
