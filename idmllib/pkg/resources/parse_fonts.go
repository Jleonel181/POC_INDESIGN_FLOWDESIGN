package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseFontsFile parsea un archivo Fonts.xml en un struct FontsFile.
func ParseFontsFile(data []byte) (*FontsFile, error) {
	// Verificar que los datos de entrada no sean nil
	if data == nil {
		return nil, common.Errorf("resources", "parse fonts", "", "input data is nil")
	}

	// Verificar que los datos de entrada no estén vacíos
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse fonts", "", "input data is empty")
	}

	var fonts FontsFile
	if err := xml.Unmarshal(data, &fonts); err != nil {
		return nil, common.WrapError("resources", "parse fonts", err)
	}
	return &fonts, nil
}

// MarshalFontsFile serializa un struct FontsFile de vuelta a XML con formato adecuado.
func MarshalFontsFile(fonts *FontsFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(fonts, "", "\t")
}

// UnmarshalXML implementa la deserialización XML personalizada para FontsFile.
func (f *FontsFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if d == nil {
		return common.Errorf("resources", "unmarshal fonts", "", "decoder is nil")
	}

	// Verificar que estamos parseando un elemento idPkg:Fonts
	if start.Name.Local != "Fonts" {
		return common.WrapError("resources", "unmarshal fonts", common.ErrInvalidFormat)
	}

	// Extraer DOMVersion de los atributos de idPkg:Fonts
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			f.DOMVersion = attr.Value
			break
		}
	}

	// Definir un struct temporal para deserializar el contenido interno
	type fontsContent struct {
		FontFamilies   []FontFamily           `xml:"FontFamily,omitempty"`
		CompositeFonts []CompositeFont        `xml:"CompositeFont,omitempty"`
		OtherElements  []common.RawXMLElement `xml:",any"`
	}

	var content fontsContent
	if err := d.DecodeElement(&content, &start); err != nil {
		return common.WrapError("resources", "unmarshal fonts content", err)
	}

	// Copiar el contenido parseado al FontsFile
	f.FontFamilies = content.FontFamilies
	f.CompositeFonts = content.CompositeFonts
	f.OtherElements = content.OtherElements

	return nil
}

// MarshalXML implementa la serialización XML personalizada para FontsFile.
func (f *FontsFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Crear el elemento contenedor idPkg:Fonts
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Fonts"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: f.DOMVersion},
		},
	}

	// Abrir el elemento contenedor
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Codificar todas las familias tipográficas
	for _, family := range f.FontFamilies {
		if err := e.EncodeElement(&family, xml.StartElement{Name: xml.Name{Local: "FontFamily"}}); err != nil {
			return err
		}
	}

	// Codificar todas las fuentes compuestas
	for _, composite := range f.CompositeFonts {
		if err := e.EncodeElement(&composite, xml.StartElement{Name: xml.Name{Local: "CompositeFont"}}); err != nil {
			return err
		}
	}

	// Codificar los demás elementos
	for _, elem := range f.OtherElements {
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
