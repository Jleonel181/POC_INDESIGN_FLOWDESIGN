package story

import (
	"encoding/xml"

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
	// Establecer nombre del elemento
	start.Name = c.XMLName

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
