package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Lectura y escritura propias de ObjectStyleGroup, para conservar el orden de sus hijos.
//
// El defecto que arreglan: el struct declara `ObjectStyles` antes que `NestedGroups`, así
// que al serializar salen primero todos los estilos y después todos los grupos anidados.
// En el corpus aparecen **las dos** disposiciones, medidas sobre los cinco documentos:
// `[ObjectStyleGroup, ObjectStyle ×4]` en 2 casos, con el grupo delante, y
// `[ObjectStyle, ObjectStyleGroup]` en otros 2, con el estilo delante. Ninguna disposición
// fija de los campos sirve para las dos, hace falta recordar el orden de cada elemento.
//
// El orden se reproduce con xmlutil.ChildOrder, el mismo registro que ya usan el designmap,
// los spreads y el elemento Properties.
//
// Cuidado con el nombre del elemento: este tipo se emite como `RootObjectStyleGroup` en la
// raíz y como `ObjectStyleGroup` cuando está anidado, y su campo `XMLName` no lleva
// etiqueta, así que el nombre sale de lo que se leyó. Eso importa porque encoding/xml
// **ignora el campo XMLName en un tipo que implementa Marshaler** y cae en el nombre del
// tipo de Go, que aquí es «ObjectStyleGroup»: si se dejara decidir a la biblioteca, la raíz
// se emitiría con el nombre del anidado. Por eso XMLName tiene prioridad al resolverlo.

const (
	tagObjectStyle      = "ObjectStyle"
	tagObjectStyleGroup = "ObjectStyleGroup"
)

// objectStyleGroupFieldOrder es el orden en que ObjectStyleGroup declara sus campos de
// hijos. Es el orden en que se emite un grupo construido desde cero, que nunca se parseó y
// por tanto no tiene orden registrado: el comportamiento anterior a esta tarea.
//
// Se copia al usarlo, porque las clases del comodín se añaden en tiempo de ejecución.
var objectStyleGroupFieldOrder = []string{
	tagObjectStyle,
	tagObjectStyleGroup,
}

// UnmarshalXML implementa la deserialización propia de ObjectStyleGroup, recordando el
// orden de sus hijos. Los grupos anidados se deserializan por recursión, así que el orden
// se conserva en cada nivel.
func (g *ObjectStyleGroup) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("resources", "unmarshal object style group", "", "decoder is nil")
	}

	// Se limpia antes de rellenar, para que deserializar dos veces sobre el mismo valor no
	// acumule hijos ni duplique el orden registrado.
	g.reset()
	g.XMLName = start.Name

	if err := xmlutil.UnmarshalAttrs(start.Attr, g); err != nil {
		return common.WrapError("resources", "unmarshal object style group", err)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case tagObjectStyle:
				var s ObjectStyle
				if err := d.DecodeElement(&s, &t); err != nil {
					return err
				}
				g.ObjectStyles = append(g.ObjectStyles, s)
			case tagObjectStyleGroup:
				var nested ObjectStyleGroup
				if err := d.DecodeElement(&nested, &t); err != nil {
					return err
				}
				g.NestedGroups = append(g.NestedGroups, nested)
			default:
				var raw common.RawXMLElement
				if err := d.DecodeElement(&raw, &t); err != nil {
					return err
				}
				g.OtherElements = append(g.OtherElements, raw)
			}
			g.childOrder.Record(t.Name.Local)

		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

// MarshalXML implementa la serialización propia de ObjectStyleGroup, emitiendo los hijos en
// el orden registrado al parsear.
func (g *ObjectStyleGroup) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// XMLName manda: es lo que distingue la raíz «RootObjectStyleGroup» del anidado
	// «ObjectStyleGroup», y el nombre que llega en start puede ser el del tipo de Go.
	name := g.XMLName
	if name.Local == "" {
		name = start.Name
	}
	if name.Local == "" {
		name = xml.Name{Local: tagObjectStyleGroup}
	}

	attrs, err := xmlutil.MarshalAttrs(g)
	if err != nil {
		return common.WrapError("resources", "marshal object style group", err)
	}

	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return common.WrapError("resources", "marshal object style group", err)
	}
	if err := g.emitChildren(e); err != nil {
		return common.WrapError("resources", "marshal object style group", err)
	}
	return e.EncodeToken(el.End())
}

// emitChildren emite los hijos reproduciendo el orden registrado.
func (g *ObjectStyleGroup) emitChildren(e *xml.Encoder) error {
	children := make(map[string][]xmlutil.ChildEmitter, len(objectStyleGroupFieldOrder))
	add := func(kind string, fn xmlutil.ChildEmitter) {
		children[kind] = append(children[kind], fn)
	}
	emit := func(v any, local string) xmlutil.ChildEmitter {
		return func() error {
			return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: local}})
		}
	}

	for i := range g.ObjectStyles {
		add(tagObjectStyle, emit(&g.ObjectStyles[i], tagObjectStyle))
	}
	for i := range g.NestedGroups {
		// El nombre se pasa explícito para que un grupo anidado no arrastre el XMLName de
		// la raíz si el modelo se construyó a mano.
		add(tagObjectStyleGroup, emit(&g.NestedGroups[i], tagObjectStyleGroup))
	}

	fieldOrder := append([]string(nil), objectStyleGroupFieldOrder...)
	for i := range g.OtherElements {
		local := g.OtherElements[i].XMLName.Local
		add(local, emit(&g.OtherElements[i], local))
		fieldOrder = append(fieldOrder, local)
	}

	return g.childOrder.Replay(fieldOrder, children)
}

// reset deja el grupo como recién creado.
func (g *ObjectStyleGroup) reset() {
	g.childOrder.Reset()
	g.ObjectStyles = nil
	g.NestedGroups = nil
	g.OtherElements = nil
	// OtherAttrs lo vacía UnmarshalAttrs antes de rellenarlo.
}

// ChildTags devuelve la secuencia de nombres de elemento de los hijos, en el orden
// registrado al parsear. Es para los tests y el diagnóstico.
func (g *ObjectStyleGroup) ChildTags() []string {
	return g.childOrder.Kinds()
}
