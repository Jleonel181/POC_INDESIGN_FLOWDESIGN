package common

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlorder"
)

// Lectura y escritura propias de Properties, para conservar el orden de sus hijos.
//
// El defecto que arreglan: `Properties` declara `PathGeometry` y `Label` como campos
// tipados y el resto de sus hijos cae en el comodín `OtherElements`. Al serializar,
// encoding/xml emite los campos en el orden en que están declarados, así que los dos
// tipados salen **antes** de todo lo del comodín, aunque en el XML de entrada vinieran
// después.
//
// Por qué no basta con reordenar los campos del struct, que sería mucho más barato: en el
// corpus aparecen **las dos** disposiciones. Medido sobre los 2224 elementos `<Properties>`
// de los cinco documentos, 10 traen un hijo tipado **después** de los del comodín
// —`PageNumberStyle, Label` en el `Properties` de un `Section`, o
// `PageColor, Descriptor, Label` en el de una `Page`— y 5 lo traen **antes**, todos con la
// forma `Label, AppliedMathMLSwatch` en el designmap. Poner el comodín primero arreglaría
// los 10 y rompería los 5. Ninguna disposición fija de los campos sirve, hace falta
// recordar el orden de cada elemento.
//
// Dos consecuencias de tener lectura y escritura propias, y las dos son trampas conocidas:
//
//  1. La etiqueta `xml:",any,attr"` de OtherAttrs **deja de aplicarse**, porque encoding/xml
//     ya no reparte los atributos. Aquí el reparto es trivial y no hace falta el ayudante
//     con reflexión de xmlutil: `Properties` no declara **ningún** atributo con nombre, solo
//     el comodín, así que todos los atributos van y vienen tal cual.
//  2. encoding/xml ignora el campo `XMLName` en un tipo que implementa Marshaler y cae en el
//     nombre del tipo de Go. Aquí no puede dar problema porque el tipo se llama igual que el
//     elemento, `Properties`, pero el nombre se resuelve explícitamente para no depender de
//     esa coincidencia.

// propertiesFieldOrder es el orden en que Properties declara sus campos de hijos. Es el
// orden en que se emite un valor construido desde cero, que nunca se parseó y por tanto no
// tiene orden registrado: exactamente el comportamiento anterior a esta tarea.
//
// Se copia al usarlo, porque las clases del comodín se añaden en tiempo de ejecución.
var propertiesFieldOrder = []string{
	tagPathGeometry,
	tagLabel,
}

const (
	tagPathGeometry = "PathGeometry"
	tagLabel        = "Label"
	tagProperties   = "Properties"
)

// UnmarshalXML implementa la deserialización propia de Properties, recordando el orden de
// sus hijos.
func (p *Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Se limpia antes de rellenar. Sin esto, deserializar dos veces sobre el mismo valor
	// acumularía hijos y duplicaría el orden registrado.
	p.reset()
	p.XMLName = start.Name

	// Properties no declara ningún atributo con nombre, así que todos son «no declarados»
	// y van al comodín tal cual, conservando prefijo, nombre, valor y orden.
	if len(start.Attr) > 0 {
		p.OtherAttrs = append(p.OtherAttrs, start.Attr...)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case tagPathGeometry:
				var pg PathGeometry
				if err := d.DecodeElement(&pg, &t); err != nil {
					return err
				}
				p.PathGeometry = &pg
			case tagLabel:
				var l Label
				if err := d.DecodeElement(&l, &t); err != nil {
					return err
				}
				p.Label = &l
			default:
				var raw RawXMLElement
				if err := d.DecodeElement(&raw, &t); err != nil {
					return err
				}
				p.OtherElements = append(p.OtherElements, raw)
			}
			p.childOrder.Record(t.Name.Local)

		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

// MarshalXML implementa la serialización propia de Properties, emitiendo los hijos en el
// orden registrado al parsear.
func (p *Properties) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	name := start.Name
	if name.Local == "" {
		name = p.XMLName
	}
	if name.Local == "" {
		name = xml.Name{Local: tagProperties}
	}

	el := xml.StartElement{Name: name, Attr: p.OtherAttrs}
	if err := e.EncodeToken(el); err != nil {
		return WrapError("common", "marshal properties", err)
	}
	if err := p.emitChildren(e); err != nil {
		return WrapError("common", "marshal properties", err)
	}
	return e.EncodeToken(el.End())
}

// emitChildren emite los hijos reproduciendo el orden registrado.
func (p *Properties) emitChildren(e *xml.Encoder) error {
	children := make(map[string][]xmlorder.ChildEmitter, len(propertiesFieldOrder)+len(p.OtherElements))
	add := func(kind string, fn xmlorder.ChildEmitter) {
		children[kind] = append(children[kind], fn)
	}
	emit := func(v any, local string) xmlorder.ChildEmitter {
		return func() error {
			return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: local}})
		}
	}

	if p.PathGeometry != nil {
		add(tagPathGeometry, emit(p.PathGeometry, tagPathGeometry))
	}
	if p.Label != nil {
		add(tagLabel, emit(p.Label, tagLabel))
	}

	// Las clases del comodín son dinámicas: en el corpus hay más de cien nombres distintos
	// de hijo de Properties. Replay tiene que conocer **todas** las clases que puede
	// recibir, o una que falte no se emitiría cuando el registro no la menciona.
	fieldOrder := append([]string(nil), propertiesFieldOrder...)
	for i := range p.OtherElements {
		local := p.OtherElements[i].XMLName.Local
		add(local, emit(&p.OtherElements[i], local))
		fieldOrder = append(fieldOrder, local)
	}

	return p.childOrder.Replay(fieldOrder, children)
}

// reset deja el Properties como recién creado.
func (p *Properties) reset() {
	p.childOrder.Reset()
	p.PathGeometry = nil
	p.Label = nil
	p.OtherElements = nil
	p.OtherAttrs = nil
}

// ChildTags devuelve la secuencia de nombres de elemento de los hijos, en el orden
// registrado al parsear. Es para los tests y el diagnóstico: permite comprobar el orden sin
// serializar.
func (p *Properties) ChildTags() []string {
	return p.childOrder.Kinds()
}
