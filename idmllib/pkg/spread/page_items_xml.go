package spread

// Serialización custom de los elementos de página que tienen hijos tipados.
//
// Sin esto, encoding/xml emite los hijos en orden de campos del struct, que no
// coincide con el orden del documento de InDesign. El ChildOrder registra la
// secuencia leída y la reproduce al emitir.
//
// Los tipos cubiertos son: SpreadTextFrame, Rectangle, Oval, Polygon, GraphicLine.
// Image y PDF no necesitan custom marshal: sus hijos no presentan desórdenes en
// el corpus.

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ---------- helpers compartidos ----------

// pageItemChild describe un hijo tipado de un elemento de página: su nombre XML y
// una función que lo emite. Se usa para construir el mapa de emitters de ChildOrder.Replay.
type pageItemChildSpec struct {
	tag string
	fn  xmlutil.ChildEmitter
}

// emitPageItemChildren es el patrón compartido por todos los tipos: recorre los hijos
// tipados y los OtherElements, y los emite con ChildOrder.Replay.
func emitPageItemChildren(e *xml.Encoder, order *xmlutil.ChildOrder, specs []pageItemChildSpec, others []common.RawXMLElement) error {
	children := make(map[string][]xmlutil.ChildEmitter)
	fieldOrder := make([]string, 0, len(specs)+len(others))
	seen := make(map[string]bool, len(specs))

	for _, s := range specs {
		children[s.tag] = append(children[s.tag], s.fn)
		if !seen[s.tag] {
			fieldOrder = append(fieldOrder, s.tag)
			seen[s.tag] = true
		}
	}
	for i := range others {
		local := others[i].XMLName.Local
		idx := i
		children[local] = append(children[local], func() error {
			return e.Encode(&others[idx])
		})
		if !seen[local] {
			fieldOrder = append(fieldOrder, local)
			seen[local] = true
		}
	}

	return order.Replay(fieldOrder, children)
}

// encodeElem es un helper que crea un emitter para un puntero a struct.
func encodeElem(e *xml.Encoder, v any, tag string) xmlutil.ChildEmitter {
	return func() error {
		return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: tag}})
	}
}

// decodeOther decodifica un hijo no reconocido como RawXMLElement.
func decodeOther(d *xml.Decoder, start xml.StartElement) (common.RawXMLElement, error) {
	var raw common.RawXMLElement
	if err := d.DecodeElement(&raw, &start); err != nil {
		return raw, err
	}
	return raw, nil
}

// ---------- SpreadTextFrame ----------

func (f *SpreadTextFrame) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	f.Properties = nil
	f.ObjectExportOption = nil
	f.TextFramePreference = nil
	f.TextWrapPreference = nil
	f.TransparencySetting = nil
	f.OtherElements = nil
	f.OtherAttrs = nil
	f.childOrder.Reset()

	if err := xmlutil.UnmarshalAttrs(start.Attr, f); err != nil {
		return err
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Properties":
				var v common.Properties
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				f.Properties = &v
			case "ObjectExportOption":
				var v ObjectExportOption
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				f.ObjectExportOption = &v
			case "TextFramePreference":
				var v TextFramePreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				f.TextFramePreference = &v
			case "TextWrapPreference":
				var v TextWrapPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				f.TextWrapPreference = &v
			case "TransparencySetting":
				var v TransparencySetting
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				f.TransparencySetting = &v
			default:
				raw, err := decodeOther(d, t)
				if err != nil {
					return err
				}
				f.OtherElements = append(f.OtherElements, raw)
			}
			f.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

func (f SpreadTextFrame) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	name := start.Name
	if name.Local == "" || name.Local == "SpreadTextFrame" {
		name = xml.Name{Local: TagTextFrame}
	}

	attrs, err := xmlutil.MarshalAttrs(&f)
	if err != nil {
		return err
	}
	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	var specs []pageItemChildSpec
	if f.Properties != nil {
		specs = append(specs, pageItemChildSpec{"Properties", encodeElem(e, f.Properties, "Properties")})
	}
	if f.ObjectExportOption != nil {
		specs = append(specs, pageItemChildSpec{"ObjectExportOption", encodeElem(e, f.ObjectExportOption, "ObjectExportOption")})
	}
	if f.TextFramePreference != nil {
		specs = append(specs, pageItemChildSpec{"TextFramePreference", encodeElem(e, f.TextFramePreference, "TextFramePreference")})
	}
	if f.TextWrapPreference != nil {
		specs = append(specs, pageItemChildSpec{"TextWrapPreference", encodeElem(e, f.TextWrapPreference, "TextWrapPreference")})
	}
	if f.TransparencySetting != nil {
		specs = append(specs, pageItemChildSpec{"TransparencySetting", encodeElem(e, f.TransparencySetting, "TransparencySetting")})
	}

	if err := emitPageItemChildren(e, &f.childOrder, specs, f.OtherElements); err != nil {
		return err
	}
	return e.EncodeToken(el.End())
}

// ---------- Rectangle ----------

func (r *Rectangle) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	r.Properties = nil
	r.FrameFittingOption = nil
	r.ObjectExportOption = nil
	r.TextWrapPreference = nil
	r.TransparencySetting = nil
	r.InCopyExportOption = nil
	r.Image = nil
	r.PDF = nil
	r.OtherElements = nil
	r.OtherAttrs = nil
	r.childOrder.Reset()

	if err := xmlutil.UnmarshalAttrs(start.Attr, r); err != nil {
		return err
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Properties":
				var v common.Properties
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.Properties = &v
			case "FrameFittingOption":
				var v FrameFittingOption
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.FrameFittingOption = &v
			case "ObjectExportOption":
				var v ObjectExportOption
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.ObjectExportOption = &v
			case "TextWrapPreference":
				var v TextWrapPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.TextWrapPreference = &v
			case "TransparencySetting":
				var v TransparencySetting
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.TransparencySetting = &v
			case "InCopyExportOption":
				var v InCopyExportOption
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.InCopyExportOption = &v
			case "Image":
				var v Image
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.Image = &v
			case "PDF":
				var v PDF
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				r.PDF = &v
			default:
				raw, err := decodeOther(d, t)
				if err != nil {
					return err
				}
				r.OtherElements = append(r.OtherElements, raw)
			}
			r.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

func (r Rectangle) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	name := start.Name
	if name.Local == "" || name.Local == "Rectangle" {
		name = xml.Name{Local: TagRectangle}
	}

	attrs, err := xmlutil.MarshalAttrs(&r)
	if err != nil {
		return err
	}
	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	var specs []pageItemChildSpec
	if r.Properties != nil {
		specs = append(specs, pageItemChildSpec{"Properties", encodeElem(e, r.Properties, "Properties")})
	}
	if r.FrameFittingOption != nil {
		specs = append(specs, pageItemChildSpec{"FrameFittingOption", encodeElem(e, r.FrameFittingOption, "FrameFittingOption")})
	}
	if r.ObjectExportOption != nil {
		specs = append(specs, pageItemChildSpec{"ObjectExportOption", encodeElem(e, r.ObjectExportOption, "ObjectExportOption")})
	}
	if r.TextWrapPreference != nil {
		specs = append(specs, pageItemChildSpec{"TextWrapPreference", encodeElem(e, r.TextWrapPreference, "TextWrapPreference")})
	}
	if r.TransparencySetting != nil {
		specs = append(specs, pageItemChildSpec{"TransparencySetting", encodeElem(e, r.TransparencySetting, "TransparencySetting")})
	}
	if r.InCopyExportOption != nil {
		specs = append(specs, pageItemChildSpec{"InCopyExportOption", encodeElem(e, r.InCopyExportOption, "InCopyExportOption")})
	}
	if r.Image != nil {
		specs = append(specs, pageItemChildSpec{"Image", encodeElem(e, r.Image, "Image")})
	}
	if r.PDF != nil {
		specs = append(specs, pageItemChildSpec{"PDF", encodeElem(e, r.PDF, "PDF")})
	}

	if err := emitPageItemChildren(e, &r.childOrder, specs, r.OtherElements); err != nil {
		return err
	}
	return e.EncodeToken(el.End())
}

// ---------- Oval ----------

func (o *Oval) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	o.Properties = nil
	o.TextWrapPreference = nil
	o.Image = nil
	o.OtherElements = nil
	o.OtherAttrs = nil
	o.childOrder.Reset()

	if err := xmlutil.UnmarshalAttrs(start.Attr, o); err != nil {
		return err
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Properties":
				var v common.Properties
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				o.Properties = &v
			case "TextWrapPreference":
				var v TextWrapPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				o.TextWrapPreference = &v
			case "Image":
				var v Image
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				o.Image = &v
			default:
				raw, err := decodeOther(d, t)
				if err != nil {
					return err
				}
				o.OtherElements = append(o.OtherElements, raw)
			}
			o.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

func (o Oval) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	name := start.Name
	if name.Local == "" || name.Local == "Oval" {
		name = xml.Name{Local: TagOval}
	}

	attrs, err := xmlutil.MarshalAttrs(&o)
	if err != nil {
		return err
	}
	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	var specs []pageItemChildSpec
	if o.Properties != nil {
		specs = append(specs, pageItemChildSpec{"Properties", encodeElem(e, o.Properties, "Properties")})
	}
	if o.TextWrapPreference != nil {
		specs = append(specs, pageItemChildSpec{"TextWrapPreference", encodeElem(e, o.TextWrapPreference, "TextWrapPreference")})
	}
	if o.Image != nil {
		specs = append(specs, pageItemChildSpec{"Image", encodeElem(e, o.Image, "Image")})
	}

	if err := emitPageItemChildren(e, &o.childOrder, specs, o.OtherElements); err != nil {
		return err
	}
	return e.EncodeToken(el.End())
}

// ---------- Polygon ----------

func (p *Polygon) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Properties = nil
	p.TextWrapPreference = nil
	p.Image = nil
	p.OtherElements = nil
	p.OtherAttrs = nil
	p.childOrder.Reset()

	if err := xmlutil.UnmarshalAttrs(start.Attr, p); err != nil {
		return err
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Properties":
				var v common.Properties
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.Properties = &v
			case "TextWrapPreference":
				var v TextWrapPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.TextWrapPreference = &v
			case "Image":
				var v Image
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.Image = &v
			default:
				raw, err := decodeOther(d, t)
				if err != nil {
					return err
				}
				p.OtherElements = append(p.OtherElements, raw)
			}
			p.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

func (p Polygon) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	name := start.Name
	if name.Local == "" || name.Local == "Polygon" {
		name = xml.Name{Local: TagPolygon}
	}

	attrs, err := xmlutil.MarshalAttrs(&p)
	if err != nil {
		return err
	}
	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	var specs []pageItemChildSpec
	if p.Properties != nil {
		specs = append(specs, pageItemChildSpec{"Properties", encodeElem(e, p.Properties, "Properties")})
	}
	if p.TextWrapPreference != nil {
		specs = append(specs, pageItemChildSpec{"TextWrapPreference", encodeElem(e, p.TextWrapPreference, "TextWrapPreference")})
	}
	if p.Image != nil {
		specs = append(specs, pageItemChildSpec{"Image", encodeElem(e, p.Image, "Image")})
	}

	if err := emitPageItemChildren(e, &p.childOrder, specs, p.OtherElements); err != nil {
		return err
	}
	return e.EncodeToken(el.End())
}

// ---------- GraphicLine ----------

func (gl *GraphicLine) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	gl.PathGeometry = nil
	gl.Properties = nil
	gl.TextWrapPreference = nil
	gl.ObjectExportOption = nil
	gl.OtherElements = nil
	gl.OtherAttrs = nil
	gl.childOrder.Reset()

	if err := xmlutil.UnmarshalAttrs(start.Attr, gl); err != nil {
		return err
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "PathGeometry":
				var v common.PathGeometry
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				gl.PathGeometry = &v
			case "Properties":
				var v common.Properties
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				gl.Properties = &v
			case "TextWrapPreference":
				var v TextWrapPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				gl.TextWrapPreference = &v
			case "ObjectExportOption":
				var v ObjectExportOption
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				gl.ObjectExportOption = &v
			default:
				raw, err := decodeOther(d, t)
				if err != nil {
					return err
				}
				gl.OtherElements = append(gl.OtherElements, raw)
			}
			gl.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

func (gl GraphicLine) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	name := start.Name
	if name.Local == "" || name.Local == "GraphicLine" {
		name = xml.Name{Local: TagGraphicLine}
	}

	attrs, err := xmlutil.MarshalAttrs(&gl)
	if err != nil {
		return err
	}
	el := xml.StartElement{Name: name, Attr: attrs}
	if err := e.EncodeToken(el); err != nil {
		return err
	}

	var specs []pageItemChildSpec
	if gl.PathGeometry != nil {
		specs = append(specs, pageItemChildSpec{"PathGeometry", encodeElem(e, gl.PathGeometry, "PathGeometry")})
	}
	if gl.Properties != nil {
		specs = append(specs, pageItemChildSpec{"Properties", encodeElem(e, gl.Properties, "Properties")})
	}
	if gl.TextWrapPreference != nil {
		specs = append(specs, pageItemChildSpec{"TextWrapPreference", encodeElem(e, gl.TextWrapPreference, "TextWrapPreference")})
	}
	if gl.ObjectExportOption != nil {
		specs = append(specs, pageItemChildSpec{"ObjectExportOption", encodeElem(e, gl.ObjectExportOption, "ObjectExportOption")})
	}

	if err := emitPageItemChildren(e, &gl.childOrder, specs, gl.OtherElements); err != nil {
		return err
	}
	return e.EncodeToken(el.End())
}
