package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Lectura y escritura propias de MasterSpreadElement.
//
// Sigue exactamente el mismo patrón que SpreadElement (spread_xml.go): contenedor
// ordenado con xmlutil.ChildOrder, UnmarshalAttrs/MarshalAttrs para los atributos, y
// Replay para emitir los hijos en Orden_Documental.
//
// La única diferencia estructural es el nombre del elemento raíz del wrapper
// (idPkg:MasterSpread) y del interno (MasterSpread).

const (
	// goTypeMasterSpread es el nombre del tipo de Go. encoding/xml lo usa como nombre
	// de elemento por defecto en un tipo con Marshaler propio, así que hay que
	// detectarlo y corregirlo.
	goTypeMasterSpread = "MasterSpreadElement"
)

// masterSpreadChildFieldOrder es el orden de emisión por defecto para un
// MasterSpreadElement construido desde cero.
var masterSpreadChildFieldOrder = []string{
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

// UnmarshalXML implementa la deserialización propia de MasterSpreadElement, conservando
// el Orden_Documental de los hijos.
func (mse *MasterSpreadElement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("spread", "unmarshal master spread element", "", "decoder is nil")
	}

	mse.reset()
	mse.XMLName = start.Name

	if err := xmlutil.UnmarshalAttrs(start.Attr, mse); err != nil {
		return common.WrapError("spread", "unmarshal master spread element", err)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if err := mse.decodeChild(d, t); err != nil {
				return err
			}
		case xml.EndElement:
			if t.Name == start.Name {
				mse.rebuildItems()
				return nil
			}
		}
	}
}

// decodeChild deserializa un hijo en su campo por tipo y anota su clase en el registro
// de orden. Idéntico al de SpreadElement.
func (mse *MasterSpreadElement) decodeChild(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case tagFlattenerPreference:
		var fp FlattenerPreference
		if err := d.DecodeElement(&fp, &start); err != nil {
			return err
		}
		mse.FlattenerPreference = &fp

	case tagPage:
		var p Page
		if err := d.DecodeElement(&p, &start); err != nil {
			return err
		}
		mse.Pages = append(mse.Pages, p)

	case TagTextFrame:
		var x SpreadTextFrame
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.textFrames = append(mse.textFrames, x)

	case TagRectangle:
		var x Rectangle
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.rectangles = append(mse.rectangles, x)

	case TagImage:
		var x Image
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.images = append(mse.images, x)

	case TagOval:
		var x Oval
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.ovals = append(mse.ovals, x)

	case TagPolygon:
		var x Polygon
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.polygons = append(mse.polygons, x)

	case TagGraphicLine:
		var x GraphicLine
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.graphicLines = append(mse.graphicLines, x)

	case TagGroup:
		var x Group
		if err := d.DecodeElement(&x, &start); err != nil {
			return err
		}
		mse.groups = append(mse.groups, x)

	default:
		var raw common.RawXMLElement
		if err := d.DecodeElement(&raw, &start); err != nil {
			return err
		}
		mse.OtherElements = append(mse.OtherElements, raw)
	}

	mse.childOrder.Record(start.Name.Local)
	return nil
}

// MarshalXML implementa la serialización propia de MasterSpreadElement, emitiendo los
// hijos en el Orden_Documental registrado al parsear.
func (mse *MasterSpreadElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Misma trampa que SpreadElement: encoding/xml ignora XMLName en un Marshaler.
	name := start.Name
	if name.Local == "" || name.Local == goTypeMasterSpread {
		name = mse.XMLName
	}
	if name.Local == "" || name.Local == goTypeMasterSpread {
		name = xml.Name{Local: "MasterSpread"}
	}

	attrs, err := xmlutil.MarshalAttrs(mse)
	if err != nil {
		return common.WrapError("spread", "marshal master spread element", err)
	}

	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return common.WrapError("spread", "marshal master spread element", err)
	}
	if err := mse.emitChildren(e); err != nil {
		return common.WrapError("spread", "marshal master spread element", err)
	}
	return e.EncodeToken(el.End())
}

// emitChildren emite los hijos reproduciendo el orden registrado.
func (mse *MasterSpreadElement) emitChildren(e *xml.Encoder) error {
	children := make(map[string][]xmlutil.ChildEmitter, len(masterSpreadChildFieldOrder))
	add := func(kind string, fn xmlutil.ChildEmitter) {
		children[kind] = append(children[kind], fn)
	}
	emit := func(v any, local string) xmlutil.ChildEmitter {
		return func() error {
			return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: local}})
		}
	}

	if mse.FlattenerPreference != nil {
		add(tagFlattenerPreference, emit(mse.FlattenerPreference, tagFlattenerPreference))
	}
	for i := range mse.Pages {
		add(tagPage, emit(&mse.Pages[i], tagPage))
	}
	for i := range mse.textFrames {
		add(TagTextFrame, emit(&mse.textFrames[i], TagTextFrame))
	}
	for i := range mse.rectangles {
		add(TagRectangle, emit(&mse.rectangles[i], TagRectangle))
	}
	for i := range mse.images {
		add(TagImage, emit(&mse.images[i], TagImage))
	}
	for i := range mse.ovals {
		add(TagOval, emit(&mse.ovals[i], TagOval))
	}
	for i := range mse.polygons {
		add(TagPolygon, emit(&mse.polygons[i], TagPolygon))
	}
	for i := range mse.graphicLines {
		add(TagGraphicLine, emit(&mse.graphicLines[i], TagGraphicLine))
	}
	for i := range mse.groups {
		add(TagGroup, emit(&mse.groups[i], TagGroup))
	}

	fieldOrder := append([]string(nil), masterSpreadChildFieldOrder...)
	for i := range mse.OtherElements {
		local := mse.OtherElements[i].XMLName.Local
		add(local, emit(&mse.OtherElements[i], local))
		fieldOrder = append(fieldOrder, local)
	}

	return mse.childOrder.Replay(fieldOrder, children)
}

// rebuildItems reconstruye Items a partir del orden registrado y de los campos por tipo.
func (mse *MasterSpreadElement) rebuildItems() {
	mse.Items = nil
	usados := make(map[string]int, len(masterSpreadChildFieldOrder))

	for _, kind := range mse.childOrder.Kinds() {
		i := usados[kind]
		usados[kind]++

		var item PageItem
		switch kind {
		case TagTextFrame:
			if i < len(mse.textFrames) {
				item = &mse.textFrames[i]
			}
		case TagRectangle:
			if i < len(mse.rectangles) {
				item = &mse.rectangles[i]
			}
		case TagImage:
			if i < len(mse.images) {
				item = &mse.images[i]
			}
		case TagOval:
			if i < len(mse.ovals) {
				item = &mse.ovals[i]
			}
		case TagPolygon:
			if i < len(mse.polygons) {
				item = &mse.polygons[i]
			}
		case TagGraphicLine:
			if i < len(mse.graphicLines) {
				item = &mse.graphicLines[i]
			}
		case TagGroup:
			if i < len(mse.groups) {
				item = &mse.groups[i]
			}
		}
		if item != nil {
			mse.Items = append(mse.Items, item)
		}
	}
}

// reset deja el MasterSpreadElement como recién creado.
func (mse *MasterSpreadElement) reset() {
	mse.childOrder.Reset()
	mse.Items = nil
	mse.FlattenerPreference = nil
	mse.Pages = nil
	mse.textFrames = nil
	mse.rectangles = nil
	mse.images = nil
	mse.ovals = nil
	mse.polygons = nil
	mse.graphicLines = nil
	mse.groups = nil
	mse.OtherElements = nil
}

// --- Wrapper MasterSpread: UnmarshalXML y MarshalXML ---

// UnmarshalXML implementa deserialización XML custom para el wrapper MasterSpread.
func (ms *MasterSpread) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("spread", "unmarshal master spread", "", "decoder is nil")
	}

	if start.Name.Local != "MasterSpread" {
		return common.WrapError("spread", "parse master spread", common.ErrInvalidFormat)
	}

	// Extraer DOMVersion desde los atributos del wrapper idPkg:MasterSpread
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			ms.DOMVersion = attr.Value
			break
		}
	}

	// Buscar el elemento <MasterSpread> interno
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "MasterSpread" {
				if err := d.DecodeElement(&ms.InnerMasterSpread, &t); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if t.Name.Local == "MasterSpread" && t.Name.Space == start.Name.Space {
				return nil
			}
		}
	}
}

// MarshalXML implementa serialización XML custom para el wrapper MasterSpread.
func (ms *MasterSpread) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:MasterSpread"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: ms.DOMVersion},
		},
	}

	if err := e.EncodeToken(wrapper); err != nil {
		return common.WrapError("spread", "marshal master spread", err)
	}

	innerStart := xml.StartElement{Name: xml.Name{Local: "MasterSpread"}}
	if err := e.EncodeElement(&ms.InnerMasterSpread, innerStart); err != nil {
		return common.WrapError("spread", "marshal master spread", err)
	}

	if err := e.EncodeToken(wrapper.End()); err != nil {
		return common.WrapError("spread", "marshal master spread", err)
	}

	return nil
}
