package story

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseStory parsea un archivo XML de Story en un struct Story.
// Maneja el namespace IDML y preserva elementos desconocidos.
func ParseStory(data []byte) (*Story, error) {
	// Verificar que el input no sea nil
	if data == nil {
		return nil, common.Errorf("story", "parse story", "", "input data is nil")
	}

	// Verificar que el input no esté vacío
	if len(data) == 0 {
		return nil, common.Errorf("story", "parse story", "", "input data is empty")
	}

	var story Story
	if err := xml.Unmarshal(data, &story); err != nil {
		return nil, common.WrapError("story", "parse story", err)
	}
	return &story, nil
}

// MarshalStory serializa un struct Story de vuelta a XML con formato correcto.
// Incluye la declaración XML y el namespace correcto con prefijo idPkg.
func MarshalStory(story *Story) ([]byte, error) {
	// Serializar la story con indentación correcta
	data, err := xml.MarshalIndent(story, "", "\t")
	if err != nil {
		return nil, common.WrapError("story", "marshal story", err)
	}

	// Agregar declaración XML (usar el formato original para compatibilidad hacia atrás)
	result := []byte(xml.Header)
	result = append(result, data...)
	result = append(result, '\n')

	return result, nil
}

// UnmarshalXML implementa deserialización custom para Story para manejar el prefijo de namespace idPkg.
func (s *Story) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if d == nil {
		return common.Errorf("story", "unmarshal story", "", "decoder is nil")
	}

	// Establecer el XMLName basado en el elemento de inicio
	s.XMLName = start.Name

	// Parsear atributos
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			s.DOMVersion = attr.Value
		}
	}

	// Parsear elementos hijo
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "Story" {
				if err := d.DecodeElement(&s.StoryElement, &t); err != nil {
					return err
				}
			}
		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML implementa serialización custom para Story para usar el prefijo de namespace idPkg.
func (s Story) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Establecer el nombre del elemento con prefijo idPkg
	start.Name = xml.Name{Local: "idPkg:Story"}

	// Agregar declaración de namespace
	start.Attr = append(start.Attr, xml.Attr{
		Name:  xml.Name{Local: "xmlns:idPkg"},
		Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
	})

	// Agregar atributo DOMVersion
	start.Attr = append(start.Attr, xml.Attr{
		Name:  xml.Name{Local: "DOMVersion"},
		Value: s.DOMVersion,
	})

	// Escribir tag de apertura
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// Serializar el elemento story
	if err := e.Encode(s.StoryElement); err != nil {
		return err
	}

	// Escribir tag de cierre
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}

	return nil
}

// UnmarshalXML implementa deserialización custom para CharacterStyleRange para preservar el orden de los elementos.
func (c *CharacterStyleRange) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if d == nil {
		return common.Errorf("story", "unmarshal character style range", "", "decoder is nil")
	}

	// Establecer XMLName
	c.XMLName = start.Name

	// Parsear atributos
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "AppliedCharacterStyle":
			c.AppliedCharacterStyle = attr.Value
		case "HorizontalScale":
			c.HorizontalScale = attr.Value
		case "Tracking":
			c.Tracking = attr.Value
		default:
			// Guardar atributos desconocidos
			c.OtherAttrs = append(c.OtherAttrs, attr)
		}
	}

	// Parsear elementos hijo en orden
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Content":
				var content Content
				if err := d.DecodeElement(&content, &t); err != nil {
					return err
				}
				c.Children = append(c.Children, CharacterChild{Content: &content})

			case "Br":
				// Br es self-closing, solo consumir el elemento
				if err := d.Skip(); err != nil {
					return err
				}
				c.Children = append(c.Children, CharacterChild{Br: &Br{XMLName: t.Name}})

			default:
				// Elemento desconocido - guardar como RawXMLElement
				var raw common.RawXMLElement
				raw.XMLName = t.Name
				raw.Attrs = t.Attr

				// Leer contenido interno
				if err := d.DecodeElement(&raw, &t); err != nil {
					return err
				}
				c.Children = append(c.Children, CharacterChild{Other: &raw})
			}

		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML implementa serialización custom para CharacterStyleRange para preservar el orden de los elementos.
func (c CharacterStyleRange) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Establecer nombre del elemento. Si el struct fue construido desde cero (no
	// parseado), XMLName está vacío y hay que ponerle el nombre canónico.
	if c.XMLName.Local != "" {
		start.Name = c.XMLName
	} else {
		start.Name = xml.Name{Local: "CharacterStyleRange"}
	}

	// Agregar atributos
	if c.AppliedCharacterStyle != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "AppliedCharacterStyle"},
			Value: c.AppliedCharacterStyle,
		})
	}
	if c.HorizontalScale != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "HorizontalScale"},
			Value: c.HorizontalScale,
		})
	}
	if c.Tracking != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "Tracking"},
			Value: c.Tracking,
		})
	}

	// Agregar otros atributos
	start.Attr = append(start.Attr, c.OtherAttrs...)

	// Escribir tag de apertura
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// Escribir hijos en orden
	for _, child := range c.Children {
		if child.Content != nil {
			if err := e.Encode(child.Content); err != nil {
				return err
			}
		} else if child.Br != nil {
			// Serializar Br como tag self-closing
			brStart := xml.StartElement{Name: xml.Name{Local: "Br"}}
			if err := e.EncodeToken(brStart); err != nil {
				return err
			}
			if err := e.EncodeToken(xml.EndElement{Name: brStart.Name}); err != nil {
				return err
			}
		} else if child.Other != nil {
			if err := e.Encode(child.Other); err != nil {
				return err
			}
		}
	}

	// Escribir tag de cierre
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}

	return nil
}

// ---------- StoryElement: serialización custom para preservar el orden documental ----------

const storyElementGoTypeName = "StoryElement"

// UnmarshalXML deserializa un StoryElement preservando el orden de sus hijos.
// Los hijos modelados (StoryPreference, InCopyExportOption, ParagraphStyleRange) se
// guardan tanto en Children como en su campo por tipo, para compatibilidad con los
// ~25 accesos existentes que iteran ParagraphStyleRanges directamente.
func (se *StoryElement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("story", "unmarshal story element", "", "decoder is nil")
	}

	// Limpiar antes de rellenar: sin esto, decodificar dos veces acumula hijos.
	se.XMLName = start.Name
	se.StoryPreference = nil
	se.InCopyExportOption = nil
	se.ParagraphStyleRanges = nil
	se.OtherElements = nil
	se.Children = nil
	se.OtherAttrs = nil

	// xml:",any,attr" no se aplica a un tipo con UnmarshalXML propio.
	if err := xmlutil.UnmarshalAttrs(start.Attr, se); err != nil {
		return common.WrapError("story", "unmarshal story element", err)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "StoryPreference":
				var sp StoryPreference
				if err := d.DecodeElement(&sp, &t); err != nil {
					return err
				}
				se.StoryPreference = &sp
				se.Children = append(se.Children, StoryChild{StoryPreference: &sp})

			case "InCopyExportOption":
				var ico InCopyExportOption
				if err := d.DecodeElement(&ico, &t); err != nil {
					return err
				}
				se.InCopyExportOption = &ico
				se.Children = append(se.Children, StoryChild{InCopyExportOption: &ico})

			case "ParagraphStyleRange":
				var psr ParagraphStyleRange
				if err := d.DecodeElement(&psr, &t); err != nil {
					return err
				}
				se.ParagraphStyleRanges = append(se.ParagraphStyleRanges, psr)
				// El puntero va al elemento recién agregado al slice.
				se.Children = append(se.Children, StoryChild{
					ParagraphStyleRange: &se.ParagraphStyleRanges[len(se.ParagraphStyleRanges)-1],
				})

			default:
				var raw common.RawXMLElement
				if err := d.DecodeElement(&raw, &t); err != nil {
					return err
				}
				se.OtherElements = append(se.OtherElements, raw)
				se.Children = append(se.Children, StoryChild{
					Other: &se.OtherElements[len(se.OtherElements)-1],
				})
			}

		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML serializa un StoryElement preservando el orden documental.
//
// Si Children está poblado (struct parseado), emite en ese orden.
// Si no (struct construido desde cero), emite en orden de campos:
// StoryPreference, InCopyExportOption, ParagraphStyleRanges, OtherElements.
func (se StoryElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Resolver el nombre del elemento: misma trampa que SpreadElement.
	name := start.Name
	if name.Local == "" || name.Local == storyElementGoTypeName {
		name = se.XMLName
	}
	if name.Local == "" || name.Local == storyElementGoTypeName {
		name = xml.Name{Local: "Story"}
	}

	attrs, err := xmlutil.MarshalAttrs(&se)
	if err != nil {
		return common.WrapError("story", "marshal story element", err)
	}

	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	if len(se.Children) > 0 {
		// Orden documental: emitir desde Children.
		for i := range se.Children {
			if err := se.emitStoryChild(e, &se.Children[i]); err != nil {
				return err
			}
		}
	} else {
		// Modelo construido desde cero: orden de campos.
		if err := se.emitFieldOrder(e); err != nil {
			return err
		}
	}

	return e.EncodeToken(el.End())
}

// emitStoryChild emite un hijo individual.
func (se *StoryElement) emitStoryChild(e *xml.Encoder, child *StoryChild) error {
	switch {
	case child.StoryPreference != nil:
		return e.Encode(child.StoryPreference)
	case child.InCopyExportOption != nil:
		return e.Encode(child.InCopyExportOption)
	case child.ParagraphStyleRange != nil:
		return e.Encode(child.ParagraphStyleRange)
	case child.Other != nil:
		return e.Encode(child.Other)
	}
	return nil
}

// emitFieldOrder emite los hijos en el orden de los campos del struct, que es el
// comportamiento de siempre para un modelo construido desde cero.
func (se *StoryElement) emitFieldOrder(e *xml.Encoder) error {
	if se.StoryPreference != nil {
		if err := e.Encode(se.StoryPreference); err != nil {
			return err
		}
	}
	if se.InCopyExportOption != nil {
		if err := e.Encode(se.InCopyExportOption); err != nil {
			return err
		}
	}
	for i := range se.ParagraphStyleRanges {
		if err := e.Encode(&se.ParagraphStyleRanges[i]); err != nil {
			return err
		}
	}
	for i := range se.OtherElements {
		if err := e.Encode(&se.OtherElements[i]); err != nil {
			return err
		}
	}
	return nil
}

// ---------- ParagraphStyleRange: serialización custom para preservar el orden documental ----------

const psrGoTypeName = "ParagraphStyleRange"

// UnmarshalXML deserializa un ParagraphStyleRange preservando el orden de sus hijos.
// CharacterStyleRange y Change (u otros elementos) se intercalan en Children.
func (psr *ParagraphStyleRange) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("story", "unmarshal paragraph style range", "", "decoder is nil")
	}

	// Limpiar.
	psr.XMLName = start.Name
	psr.AppliedParagraphStyle = ""
	psr.CharacterStyleRanges = nil
	psr.OtherElements = nil
	psr.Children = nil
	psr.OtherAttrs = nil

	// xml:",any,attr" no se aplica con UnmarshalXML propio.
	if err := xmlutil.UnmarshalAttrs(start.Attr, psr); err != nil {
		return common.WrapError("story", "unmarshal paragraph style range", err)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "CharacterStyleRange":
				var csr CharacterStyleRange
				if err := d.DecodeElement(&csr, &t); err != nil {
					return err
				}
				psr.CharacterStyleRanges = append(psr.CharacterStyleRanges, csr)
				psr.Children = append(psr.Children, ParagraphChild{
					CharacterStyleRange: &psr.CharacterStyleRanges[len(psr.CharacterStyleRanges)-1],
				})

			default:
				var raw common.RawXMLElement
				if err := d.DecodeElement(&raw, &t); err != nil {
					return err
				}
				psr.OtherElements = append(psr.OtherElements, raw)
				psr.Children = append(psr.Children, ParagraphChild{
					Other: &psr.OtherElements[len(psr.OtherElements)-1],
				})
			}

		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML serializa un ParagraphStyleRange preservando el orden documental.
func (psr ParagraphStyleRange) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Resolver nombre del elemento.
	name := start.Name
	if name.Local == "" || name.Local == psrGoTypeName {
		name = psr.XMLName
	}
	if name.Local == "" || name.Local == psrGoTypeName {
		name = xml.Name{Local: "ParagraphStyleRange"}
	}

	attrs, err := xmlutil.MarshalAttrs(&psr)
	if err != nil {
		return common.WrapError("story", "marshal paragraph style range", err)
	}

	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	if len(psr.Children) > 0 {
		// Orden documental.
		for i := range psr.Children {
			child := &psr.Children[i]
			switch {
			case child.CharacterStyleRange != nil:
				if err := e.Encode(child.CharacterStyleRange); err != nil {
					return err
				}
			case child.Other != nil:
				if err := e.Encode(child.Other); err != nil {
					return err
				}
			}
		}
	} else {
		// Modelo construido desde cero: CharacterStyleRanges primero, luego OtherElements.
		for i := range psr.CharacterStyleRanges {
			if err := e.Encode(&psr.CharacterStyleRanges[i]); err != nil {
				return err
			}
		}
		for i := range psr.OtherElements {
			if err := e.Encode(&psr.OtherElements[i]); err != nil {
				return err
			}
		}
	}

	return e.EncodeToken(el.End())
}
