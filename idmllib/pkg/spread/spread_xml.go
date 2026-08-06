package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Lectura y escritura propias de SpreadElement.
//
// El defecto que arreglan: encoding/xml emite los hijos en el orden en que están
// declarados los campos del struct, así que un <Spread> que entra con sus elementos de
// página intercalados —rectángulo, rectángulo, grupo, marco, marco...— sale con todos los
// marcos juntos, luego todos los rectángulos, luego los grupos. Mismos elementos, otro
// documento. Medido en el corpus: 5 spreads con la secuencia reagrupada.
//
// El orden se reproduce con xmlutil.ChildOrder, el mismo registro que ya se usa en
// designmap.xml, Styles.xml y Graphic.xml. No se inventa un mecanismo nuevo.
//
// Consecuencia de tener UnmarshalXML propio, y es una trampa real: la etiqueta
// `xml:",any,attr"` de OtherAttrs **deja de aplicarse**, porque encoding/xml ya no
// reparte los atributos. De ahí las llamadas a xmlutil.UnmarshalAttrs y MarshalAttrs, que
// hacen ese reparto leyendo del propio struct qué campos declara.

// UnmarshalXML implementa la deserialización propia de SpreadElement, conservando el
// Orden_Documental de los hijos.
func (se *SpreadElement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("spread", "unmarshal spread element", "", "decoder is nil")
	}

	// Se limpia antes de rellenar. Sin esto, decodificar dos veces sobre el mismo valor
	// acumularía hijos y duplicaría el orden registrado.
	se.reset()
	se.XMLName = start.Name

	if err := xmlutil.UnmarshalAttrs(start.Attr, se); err != nil {
		return common.WrapError("spread", "unmarshal spread element", err)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if err := se.decodeChild(d, t); err != nil {
				return err
			}
		case xml.EndElement:
			if t.Name == start.Name {
				se.rebuildItems()
				return nil
			}
		}
	}
}

// decodeChild deserializa un hijo en su campo por tipo y anota su clase en el registro
// de orden.
//
// Los dos pasos van juntos a propósito: si se anota una clase y no se guarda el hijo, o
// al contrario, la reproducción del orden queda desalineada.
func (se *SpreadElement) decodeChild(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case tagFlattenerPreference:
		var fp FlattenerPreference
		if err := d.DecodeElement(&fp, &start); err != nil {
			return err
		}
		se.FlattenerPreference = &fp

	case tagPage:
		var p Page
		if err := d.DecodeElement(&p, &start); err != nil {
			return err
		}
		se.Pages = append(se.Pages, p)

	case TagTextFrame:
		var x SpreadTextFrame
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.textFrames = append(se.textFrames, x)

	case TagRectangle:
		var x Rectangle
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.rectangles = append(se.rectangles, x)

	case TagImage:
		var x Image
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.images = append(se.images, x)

	case TagOval:
		var x Oval
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.ovals = append(se.ovals, x)

	case TagPolygon:
		var x Polygon
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.polygons = append(se.polygons, x)

	case TagGraphicLine:
		var x GraphicLine
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.graphicLines = append(se.graphicLines, x)

	case TagGroup:
		var x Group
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		se.groups = append(se.groups, x)

	default:
		// Todo lo que el modelo no declara se guarda como XML crudo, incluido <PDF>:
		// SpreadElement no tiene campo para PDF suelto a este nivel, y el corpus no
		// contiene ninguno. Va aquí para no perderlo, y su posición se conserva porque
		// el registro de orden anota su nombre de elemento como cualquier otra clase.
		var raw common.RawXMLElement
		if err := d.DecodeElement(&raw, &start); err != nil {
			return err
		}
		se.OtherElements = append(se.OtherElements, raw)
	}

	se.childOrder.Record(start.Name.Local)
	return nil
}

// MarshalXML implementa la serialización propia de SpreadElement, emitiendo los hijos en
// el Orden_Documental registrado al parsear.
func (se *SpreadElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Trampa de encoding/xml, y costó un fallo en pkg/idms: cuando un tipo implementa
	// Marshaler, defaultStart **no consulta el campo XMLName**. Si el llamador no da un
	// nombre explícito —`encoder.Encode(&spreadElem)`, que es lo que hace pkg/document
	// para el spread inline de un IDMS— el nombre que llega es el del tipo de Go, y se
	// emitiría <SpreadElement> en lugar de <Spread>. Antes de tener Marshaler propio esto
	// no pasaba, porque encoding/xml sí leía la etiqueta `xml:"Spread"` del XMLName.
	name := start.Name
	if name.Local == "" || name.Local == goTypeName {
		name = se.XMLName
	}
	if name.Local == "" || name.Local == goTypeName {
		name = xml.Name{Local: "Spread"}
	}

	attrs, err := xmlutil.MarshalAttrs(se)
	if err != nil {
		return common.WrapError("spread", "marshal spread element", err)
	}

	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return common.WrapError("spread", "marshal spread element", err)
	}
	if err := se.emitChildren(e); err != nil {
		return common.WrapError("spread", "marshal spread element", err)
	}
	return e.EncodeToken(el.End())
}

// emitChildren emite los hijos reproduciendo el orden registrado.
//
// El contenido se lee de los campos por tipo, no de Items. Es lo que permite que las 18
// escrituras existentes a esos campos sigan surtiendo efecto sin tocarlas, que es el
// criterio 3 de la Tarea 7. Ver el comentario del campo Items.
func (se *SpreadElement) emitChildren(e *xml.Encoder) error {
	children := make(map[string][]xmlutil.ChildEmitter, len(childFieldOrder))
	add := func(kind string, fn xmlutil.ChildEmitter) {
		children[kind] = append(children[kind], fn)
	}
	emit := func(v any, local string) xmlutil.ChildEmitter {
		return func() error {
			return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: local}})
		}
	}

	if se.FlattenerPreference != nil {
		add(tagFlattenerPreference, emit(se.FlattenerPreference, tagFlattenerPreference))
	}
	for i := range se.Pages {
		add(tagPage, emit(&se.Pages[i], tagPage))
	}
	for i := range se.textFrames {
		add(TagTextFrame, emit(&se.textFrames[i], TagTextFrame))
	}
	for i := range se.rectangles {
		add(TagRectangle, emit(&se.rectangles[i], TagRectangle))
	}
	for i := range se.images {
		add(TagImage, emit(&se.images[i], TagImage))
	}
	for i := range se.ovals {
		add(TagOval, emit(&se.ovals[i], TagOval))
	}
	for i := range se.polygons {
		add(TagPolygon, emit(&se.polygons[i], TagPolygon))
	}
	for i := range se.graphicLines {
		add(TagGraphicLine, emit(&se.graphicLines[i], TagGraphicLine))
	}
	for i := range se.groups {
		add(TagGroup, emit(&se.groups[i], TagGroup))
	}

	// Las clases de los hijos no modelados son dinámicas, así que se añaden al orden de
	// campos: Replay tiene que conocer **todas** las clases que puede recibir, o una que
	// falte no se emitiría cuando el registro no la menciona.
	fieldOrder := append([]string(nil), childFieldOrder...)
	for i := range se.OtherElements {
		local := se.OtherElements[i].XMLName.Local
		add(local, emit(&se.OtherElements[i], local))
		fieldOrder = append(fieldOrder, local)
	}

	return se.childOrder.Replay(fieldOrder, children)
}

// childFieldOrder es el orden en que SpreadElement declara sus campos de hijos. Es el
// orden en que se emite un modelo construido desde cero, que nunca se parseó y por tanto
// no tiene orden registrado: exactamente el comportamiento anterior a esta tarea.
//
// Se copia al usarlo porque emitChildren le añade las clases no modeladas.
var childFieldOrder = []string{
	tagFlattenerPreference,
	tagPage,
	TagTextFrame,
	TagRectangle,
	TagImage,
	TagOval,
	TagPolygon,
	TagGraphicLine,
	TagGroup,
}

const (
	tagFlattenerPreference = "FlattenerPreference"
	tagPage                = "Page"

	// goTypeName es el nombre del tipo de Go, que es con el que encoding/xml llama a
	// MarshalXML cuando el llamador no da un nombre de elemento explícito. Ver el
	// comentario de MarshalXML.
	goTypeName = "SpreadElement"
)

// rebuildItems reconstruye Items a partir del orden registrado y de los campos por tipo,
// dejando en cada posición un puntero al elemento correspondiente.
//
// Se llama una sola vez, al terminar de parsear, y no durante el parseo: tomar la
// dirección de un elemento de un slice al que todavía se le va a hacer append daría
// punteros al array viejo cuando el slice crezca.
func (se *SpreadElement) rebuildItems() {
	se.Items = nil
	usados := make(map[string]int, len(childFieldOrder))

	for _, kind := range se.childOrder.Kinds() {
		i := usados[kind]
		usados[kind]++

		var item PageItem
		switch kind {
		case TagTextFrame:
			if i < len(se.textFrames) {
				item = &se.textFrames[i]
			}
		case TagRectangle:
			if i < len(se.rectangles) {
				item = &se.rectangles[i]
			}
		case TagImage:
			if i < len(se.images) {
				item = &se.images[i]
			}
		case TagOval:
			if i < len(se.ovals) {
				item = &se.ovals[i]
			}
		case TagPolygon:
			if i < len(se.polygons) {
				item = &se.polygons[i]
			}
		case TagGraphicLine:
			if i < len(se.graphicLines) {
				item = &se.graphicLines[i]
			}
		case TagGroup:
			if i < len(se.groups) {
				item = &se.groups[i]
			}
		default:
			// FlattenerPreference, Page y los hijos no modelados no son elementos de
			// página y no entran en Items.
		}
		if item != nil {
			se.Items = append(se.Items, item)
		}
	}
}

// reset deja el SpreadElement como recién creado, para que deserializar dos veces sobre
// el mismo valor no acumule hijos.
func (se *SpreadElement) reset() {
	se.childOrder.Reset()
	se.Items = nil
	se.FlattenerPreference = nil
	se.Pages = nil
	se.textFrames = nil
	se.rectangles = nil
	se.images = nil
	se.ovals = nil
	se.polygons = nil
	se.graphicLines = nil
	se.groups = nil
	se.OtherElements = nil
	// OtherAttrs lo vacía UnmarshalAttrs antes de rellenarlo.
}

// ItemTags devuelve la secuencia de nombres de elemento de Items, en Orden_Documental.
// Es para los tests y el diagnóstico: permite comprobar el orden sin serializar.
func (se *SpreadElement) ItemTags() []string {
	tags := make([]string, 0, len(se.Items))
	for _, it := range se.Items {
		tags = append(tags, it.xmlTag())
	}
	return tags
}

// Append agrega un elemento de página al final del spread, en Orden_Documental.
//
// Es la pieza que permite construir un spread desde cero con los elementos en el orden en
// que se agregan. Sin ella, un `SpreadElement` que nunca se parseó emite sus hijos en el
// orden en que están declarados los campos del struct —todos los marcos de texto, luego
// todos los rectángulos— y no en el orden de adición. Para InDesign ese orden es el
// apilamiento, así que un documento generado con elementos superpuestos saldría con las
// capas cambiadas.
//
// Devuelve el puntero al elemento **guardado**, que no es el que se le pasó. El contenido
// vive en el campo por tipo, así que Append guarda una copia del valor apuntado; para
// seguir modificando el elemento después de agregarlo hay que usar el puntero devuelto. Lo
// natural es configurarlo antes de agregarlo, y entonces el asunto no aparece.
//
// Error si el tipo no tiene campo donde guardarse. Hoy solo pasa con *PDF: un <PDF> suelto
// como hijo directo de <Spread> no lo modela SpreadElement, y el corpus no contiene
// ninguno. Un PDF va dentro de un marco, en Rectangle.PDF.
//
// ponytail: rebuildItems recorre el registro completo en cada llamada, así que construir
// un spread de n elementos es O(n²). Con las decenas de elementos que tiene un spread real
// es irrelevante, y es lo que mantiene válidos **todos** los punteros de Items cuando el
// append a un campo por tipo reubica el array de ese slice. La vía de mejora, si algún día
// importa, es la fase 3: Items pasa a ser el contenedor y los campos por tipo la vista
// derivada, y entonces no hay punteros que revalidar.
func (se *SpreadElement) Append(item PageItem) (PageItem, error) {
	if item == nil {
		return nil, common.Errorf("spread", "append page item", "", "el elemento de página es nil")
	}

	switch v := item.(type) {
	case *SpreadTextFrame:
		se.textFrames = append(se.textFrames, *v)
	case *Rectangle:
		se.rectangles = append(se.rectangles, *v)
	case *Image:
		se.images = append(se.images, *v)
	case *Oval:
		se.ovals = append(se.ovals, *v)
	case *Polygon:
		se.polygons = append(se.polygons, *v)
	case *GraphicLine:
		se.graphicLines = append(se.graphicLines, *v)
	case *Group:
		se.groups = append(se.groups, *v)
	default:
		return nil, common.Errorf("spread", "append page item", item.GetSelf(),
			"SpreadElement no tiene campo para un <"+item.xmlTag()+"> a este nivel; "+
				"un PDF va dentro de un marco, en Rectangle.PDF")
	}

	se.childOrder.Record(item.xmlTag())
	se.rebuildItems()

	// El registro que se acaba de anotar es el último, así que el elemento guardado es el
	// último de Items. Se devuelve desde ahí para que sea el puntero canónico, el mismo
	// que se usa al emitir.
	if len(se.Items) == 0 {
		return nil, common.Errorf("spread", "append page item", item.GetSelf(),
			"el elemento no quedó registrado en Items")
	}
	return se.Items[len(se.Items)-1], nil
}
