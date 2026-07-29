package idml

import (
	"bytes"
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ResourceFile representa un archivo XML de recursos genérico (Graphic, Fonts, Styles, Preferences).
// Estos archivos contienen colecciones de definiciones de recursos (colores, fuentes, estilos, etc.).
//
// El elemento raíz es <idPkg:TipoRecurso> (ej. <idPkg:Graphic>) con el namespace idPkg.
type ResourceFile struct {
	// XMLName no se define directamente - se maneja manualmente en MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// ResourceType es el tipo de recurso (Graphic, Fonts, Styles, Preferences)
	ResourceType string `xml:"-"`

	// DOMVersion es la versión del DOM de InDesign (ej. "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// RawContent almacena todos los elementos hijos como XML crudo
	// Preserva la estructura completa sin necesidad de modelar cada elemento
	RawContent []byte `xml:",innerxml"`
}

// UnmarshalXML implementa la deserialización XML personalizada para ResourceFile.
// Maneja el elemento wrapper idPkg:TipoRecurso y preserva el contenido.
func (r *ResourceFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Guardar el tipo de recurso desde el nombre del elemento
	r.ResourceType = start.Name.Local

	// Extraer DOMVersion de los atributos
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			r.DOMVersion = attr.Value
			break
		}
	}

	// Leer todo el contenido como XML crudo
	var buf bytes.Buffer
	depth := 0

	for {
		tok, err := d.Token()
		if err != nil {
			return common.WrapError("idml", "unmarshal resource", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			buf.WriteString("<")
			buf.WriteString(t.Name.Local)
			for _, attr := range t.Attr {
				buf.WriteString(" ")
				buf.WriteString(attr.Name.Local)
				buf.WriteString(`="`)
				xml.EscapeText(&buf, []byte(attr.Value)) // nolint:errcheck
				buf.WriteString(`"`)
			}
			buf.WriteString(">")

		case xml.EndElement:
			depth--
			if depth < 0 {
				// Fin del elemento raíz
				r.RawContent = buf.Bytes()
				return nil
			}
			buf.WriteString("</")
			buf.WriteString(t.Name.Local)
			buf.WriteString(">")

		case xml.CharData:
			xml.EscapeText(&buf, t) // nolint:errcheck

		case xml.Comment:
			buf.WriteString("<!--")
			buf.Write(t)
			buf.WriteString("-->")

		case xml.ProcInst:
			buf.WriteString("<?")
			buf.WriteString(t.Target)
			buf.WriteString(" ")
			buf.Write(t.Inst)
			buf.WriteString("?>")
		}
	}
}

// MarshalXML implementa la serialización XML personalizada para ResourceFile.
func (r *ResourceFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Crear el elemento wrapper idPkg:TipoRecurso
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:" + r.ResourceType},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: r.DOMVersion},
		},
	}

	// Abrir el elemento wrapper
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Escribir el contenido crudo directamente
	if len(r.RawContent) > 0 {
		// Parsear y re-encodear para asegurar el formato correcto
		decoder := xml.NewDecoder(bytes.NewReader(r.RawContent))
		for {
			tok, err := decoder.Token()
			if err != nil {
				break
			}
			if err := e.EncodeToken(xml.CopyToken(tok)); err != nil {
				return err
			}
		}
	}

	// Cerrar el elemento wrapper
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return err
	}

	return nil
}

// ParseResourceFile parsea un archivo XML de recursos en un struct ResourceFile.
func ParseResourceFile(data []byte) (*ResourceFile, error) {
	// Verificar que los datos no sean nil
	if data == nil {
		return nil, common.Errorf("idml", "parse resource", "", "input data is nil")
	}

	// Verificar que los datos no estén vacíos
	if len(data) == 0 {
		return nil, common.Errorf("idml", "parse resource", "", "input data is empty")
	}

	var resource ResourceFile
	if err := xml.Unmarshal(data, &resource); err != nil {
		return nil, common.WrapError("idml", "parse resource", err)
	}
	return &resource, nil
}

// MarshalResourceFile serializa un struct ResourceFile de vuelta a XML con el formato correcto.
func MarshalResourceFile(resource *ResourceFile) ([]byte, error) {
	var buf bytes.Buffer

	// Agregar declaración XML
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')

	// Serializar el recurso
	resourceXML, err := xml.MarshalIndent(resource, "", "\t")
	if err != nil {
		return nil, common.WrapError("idml", "marshal resource", err)
	}

	buf.Write(resourceXML)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}
