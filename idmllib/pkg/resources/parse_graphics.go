package resources

import (
	"encoding/xml"
	"errors"
	"io"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseGraphicFile parsea un archivo Graphic.xml en un struct GraphicFile.
func ParseGraphicFile(data []byte) (*GraphicFile, error) {
	// Verificar que los datos de entrada no sean nil
	if data == nil {
		return nil, common.Errorf("resources", "parse graphic", "", "input data is nil")
	}

	// Verificar que los datos de entrada no estén vacíos
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse graphic", "", "input data is empty")
	}

	var graphic GraphicFile
	if err := xml.Unmarshal(data, &graphic); err != nil {
		return nil, common.WrapError("resources", "parse graphic", err)
	}
	return &graphic, nil
}

// MarshalGraphicFile serializa un struct GraphicFile de vuelta a XML con formato adecuado.
func MarshalGraphicFile(graphic *GraphicFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(graphic, "", "\t")
}

// Clases de hijo de <idPkg:Graphic>. Coinciden con el nombre de la etiqueta, salvo
// la de los elementos sin modelar, que van todos al mismo campo.
const (
	graphicChildColor             = "Color"
	graphicChildInk               = "Ink"
	graphicChildGradient          = "Gradient"
	graphicChildSwatch            = "Swatch"
	graphicChildPastedSmoothShade = "PastedSmoothShade"
	graphicChildStrokeStyle       = "StrokeStyle"
	graphicChildOther             = "OtherElement"
)

// graphicChildOrder es el orden en que están declarados los campos del struct, que
// es el que se usa para los hijos que el registro de orden no menciona. Tiene que
// nombrar todas las clases de arriba; lo comprueba un test.
var graphicChildOrder = []string{
	graphicChildColor,
	graphicChildInk,
	graphicChildGradient,
	graphicChildSwatch,
	graphicChildPastedSmoothShade,
	graphicChildStrokeStyle,
	graphicChildOther,
}

// UnmarshalXML implementa la deserialización XML personalizada para GraphicFile.
func (g *GraphicFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if d == nil {
		return common.Errorf("resources", "unmarshal graphic", "", "decoder is nil")
	}

	// Verificar que estamos parseando un elemento idPkg:Graphic
	if start.Name.Local != "Graphic" {
		return common.WrapError("resources", "unmarshal graphic", common.ErrInvalidFormat)
	}

	// Extraer DOMVersion de los atributos de idPkg:Graphic
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			g.DOMVersion = attr.Value
			break
		}
	}

	// Los hijos se recorren de uno en uno, y no con un struct temporal como antes,
	// porque hay que registrar en qué orden vienen: encoding/xml los reparte por
	// campos y pierde la secuencia.
	for {
		token, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return common.WrapError("resources", "unmarshal graphic content", err)
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if err := g.unmarshalChild(d, elem); err != nil {
				return err
			}
		case xml.EndElement:
			return nil
		}
	}

	return nil
}

// unmarshalChild decodifica un hijo de <idPkg:Graphic> en su campo y anota su clase
// en el registro de orden.
func (g *GraphicFile) unmarshalChild(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case graphicChildColor:
		var color Color
		if err := d.DecodeElement(&color, &start); err != nil {
			return common.WrapError("resources", "unmarshal graphic content", err)
		}
		g.Colors = append(g.Colors, color)

	case graphicChildInk:
		var ink Ink
		if err := d.DecodeElement(&ink, &start); err != nil {
			return common.WrapError("resources", "unmarshal graphic content", err)
		}
		g.Inks = append(g.Inks, ink)

	case graphicChildGradient:
		var gradient Gradient
		if err := d.DecodeElement(&gradient, &start); err != nil {
			return common.WrapError("resources", "unmarshal graphic content", err)
		}
		g.Gradients = append(g.Gradients, gradient)

	case graphicChildSwatch:
		var swatch Swatch
		if err := d.DecodeElement(&swatch, &start); err != nil {
			return common.WrapError("resources", "unmarshal graphic content", err)
		}
		g.Swatches = append(g.Swatches, swatch)

	case graphicChildPastedSmoothShade:
		var shade PastedSmoothShade
		if err := d.DecodeElement(&shade, &start); err != nil {
			return common.WrapError("resources", "unmarshal graphic content", err)
		}
		g.PastedSmoothShades = append(g.PastedSmoothShades, shade)

	case graphicChildStrokeStyle:
		var style StrokeStyle
		if err := d.DecodeElement(&style, &start); err != nil {
			return common.WrapError("resources", "unmarshal graphic content", err)
		}
		g.StrokeStyles = append(g.StrokeStyles, style)

	default:
		var raw common.RawXMLElement
		if err := d.DecodeElement(&raw, &start); err != nil {
			return common.WrapErrorWithPath("resources", "unmarshal graphic content", start.Name.Local, err)
		}
		g.OtherElements = append(g.OtherElements, raw)
		g.childOrder.Record(graphicChildOther)
		return nil
	}

	g.childOrder.Record(start.Name.Local)
	return nil
}

// MarshalXML implementa la serialización XML personalizada para GraphicFile.
func (g *GraphicFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Crear el elemento contenedor idPkg:Graphic
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Graphic"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: g.DOMVersion},
		},
	}

	// Abrir el elemento contenedor
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Emitir los hijos en el orden en que venían al parsear, o en el orden de los
	// campos si el GraphicFile se construyó desde cero.
	if err := g.childOrder.Replay(graphicChildOrder, g.childrenByKind(e)); err != nil {
		return err
	}

	// Cerrar el elemento contenedor
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return err
	}

	return nil
}

// childrenByKind agrupa los hijos que el archivo tiene ahora, por clase y en el
// orden de su campo, cada uno con la función que lo emite.
func (g *GraphicFile) childrenByKind(e *xml.Encoder) map[string][]xmlutil.ChildEmitter {
	children := make(map[string][]xmlutil.ChildEmitter, len(graphicChildOrder))

	named := func(kind string, child any) xmlutil.ChildEmitter {
		return func() error {
			return e.EncodeElement(child, xml.StartElement{Name: xml.Name{Local: kind}})
		}
	}

	for i := range g.Colors {
		children[graphicChildColor] = append(children[graphicChildColor], named(graphicChildColor, &g.Colors[i]))
	}
	for i := range g.Inks {
		children[graphicChildInk] = append(children[graphicChildInk], named(graphicChildInk, &g.Inks[i]))
	}
	for i := range g.Gradients {
		children[graphicChildGradient] = append(children[graphicChildGradient], named(graphicChildGradient, &g.Gradients[i]))
	}
	for i := range g.Swatches {
		children[graphicChildSwatch] = append(children[graphicChildSwatch], named(graphicChildSwatch, &g.Swatches[i]))
	}
	for i := range g.PastedSmoothShades {
		children[graphicChildPastedSmoothShade] = append(children[graphicChildPastedSmoothShade], named(graphicChildPastedSmoothShade, &g.PastedSmoothShades[i]))
	}
	for i := range g.StrokeStyles {
		children[graphicChildStrokeStyle] = append(children[graphicChildStrokeStyle], named(graphicChildStrokeStyle, &g.StrokeStyles[i]))
	}

	// Los elementos sin modelar se emiten con el nombre que traían.
	for i := range g.OtherElements {
		elem := &g.OtherElements[i]
		children[graphicChildOther] = append(children[graphicChildOther], func() error {
			return e.EncodeElement(elem, xml.StartElement{Name: elem.XMLName})
		})
	}

	return children
}
