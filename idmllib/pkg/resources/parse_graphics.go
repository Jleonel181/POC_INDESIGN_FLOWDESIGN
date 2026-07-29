package resources

import (
	"encoding/xml"

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

	// Definir un struct temporal para deserializar el contenido interno
	type graphicContent struct {
		Colors             []Color                `xml:"Color,omitempty"`
		Inks               []Ink                  `xml:"Ink,omitempty"`
		Gradients          []Gradient             `xml:"Gradient,omitempty"`
		Swatches           []Swatch               `xml:"Swatch,omitempty"`
		PastedSmoothShades []PastedSmoothShade    `xml:"PastedSmoothShade,omitempty"`
		StrokeStyles       []StrokeStyle          `xml:"StrokeStyle,omitempty"`
		OtherElements      []common.RawXMLElement `xml:",any"`
	}

	var content graphicContent
	if err := d.DecodeElement(&content, &start); err != nil {
		return common.WrapError("resources", "unmarshal graphic content", err)
	}

	// Copiar el contenido parseado al GraphicFile
	g.Colors = content.Colors
	g.Inks = content.Inks
	g.Gradients = content.Gradients
	g.Swatches = content.Swatches
	g.PastedSmoothShades = content.PastedSmoothShades
	g.StrokeStyles = content.StrokeStyles
	g.OtherElements = content.OtherElements

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

	// Codificar todos los elementos hijo
	for _, color := range g.Colors {
		if err := e.EncodeElement(&color, xml.StartElement{Name: xml.Name{Local: "Color"}}); err != nil {
			return err
		}
	}

	for _, ink := range g.Inks {
		if err := e.EncodeElement(&ink, xml.StartElement{Name: xml.Name{Local: "Ink"}}); err != nil {
			return err
		}
	}

	for _, gradient := range g.Gradients {
		if err := e.EncodeElement(&gradient, xml.StartElement{Name: xml.Name{Local: "Gradient"}}); err != nil {
			return err
		}
	}

	for _, swatch := range g.Swatches {
		if err := e.EncodeElement(&swatch, xml.StartElement{Name: xml.Name{Local: "Swatch"}}); err != nil {
			return err
		}
	}

	for _, shade := range g.PastedSmoothShades {
		if err := e.EncodeElement(&shade, xml.StartElement{Name: xml.Name{Local: "PastedSmoothShade"}}); err != nil {
			return err
		}
	}

	for _, style := range g.StrokeStyles {
		if err := e.EncodeElement(&style, xml.StartElement{Name: xml.Name{Local: "StrokeStyle"}}); err != nil {
			return err
		}
	}

	// Codificar los demás elementos
	for _, elem := range g.OtherElements {
		if err := e.EncodeElement(&elem, xml.StartElement{Name: elem.XMLName}); err != nil {
			return err
		}
	}

	// Cerrar el elemento contenedor
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return err
	}

	return nil
}
