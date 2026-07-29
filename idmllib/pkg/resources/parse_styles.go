package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseStylesFile parsea un archivo Styles.xml en un struct StylesFile.
func ParseStylesFile(data []byte) (*StylesFile, error) {
	// Verificar que los datos de entrada no sean nil
	if data == nil {
		return nil, common.Errorf("resources", "parse styles", "", "input data is nil")
	}

	// Verificar que los datos de entrada no estén vacíos
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse styles", "", "input data is empty")
	}

	var styles StylesFile
	if err := xml.Unmarshal(data, &styles); err != nil {
		return nil, common.WrapError("resources", "parse styles", err)
	}
	return &styles, nil
}

// MarshalStylesFile serializa un struct StylesFile de vuelta a XML con formato adecuado.
func MarshalStylesFile(styles *StylesFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(styles, "", "\t")
}

// UnmarshalXML implementa la deserialización XML personalizada para StylesFile.
func (s *StylesFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if d == nil {
		return common.Errorf("resources", "unmarshal styles", "", "decoder is nil")
	}

	// Verificar que estamos parseando un elemento idPkg:Styles
	if start.Name.Local != "Styles" {
		return common.WrapError("resources", "unmarshal styles", common.ErrInvalidFormat)
	}

	// Extraer DOMVersion de los atributos de idPkg:Styles
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			s.DOMVersion = attr.Value
			break
		}
	}

	// Definir un struct temporal para deserializar el contenido interno
	type stylesContent struct {
		RootCharacterStyleGroup *CharacterStyleGroup   `xml:"RootCharacterStyleGroup,omitempty"`
		RootParagraphStyleGroup *ParagraphStyleGroup   `xml:"RootParagraphStyleGroup,omitempty"`
		RootCellStyleGroup      *CellStyleGroup        `xml:"RootCellStyleGroup,omitempty"`
		RootTableStyleGroup     *TableStyleGroup       `xml:"RootTableStyleGroup,omitempty"`
		RootObjectStyleGroup    *ObjectStyleGroup      `xml:"RootObjectStyleGroup,omitempty"`
		TOCStyles               []TOCStyle             `xml:"TOCStyle,omitempty"`
		OtherElements           []common.RawXMLElement `xml:",any"`
	}

	var content stylesContent
	if err := d.DecodeElement(&content, &start); err != nil {
		return common.WrapError("resources", "unmarshal styles content", err)
	}

	// Copiar el contenido parseado al StylesFile
	s.RootCharacterStyleGroup = content.RootCharacterStyleGroup
	s.RootParagraphStyleGroup = content.RootParagraphStyleGroup
	s.RootCellStyleGroup = content.RootCellStyleGroup
	s.RootTableStyleGroup = content.RootTableStyleGroup
	s.RootObjectStyleGroup = content.RootObjectStyleGroup
	s.TOCStyles = content.TOCStyles
	s.OtherElements = content.OtherElements

	return nil
}

// MarshalXML implementa la serialización XML personalizada para StylesFile.
func (s *StylesFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Crear el elemento contenedor idPkg:Styles
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Styles"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: s.DOMVersion},
		},
	}

	// Abrir el elemento contenedor
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Codificar todos los grupos de estilos
	if s.RootCharacterStyleGroup != nil {
		if err := e.EncodeElement(s.RootCharacterStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootCharacterStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootParagraphStyleGroup != nil {
		if err := e.EncodeElement(s.RootParagraphStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootParagraphStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootCellStyleGroup != nil {
		if err := e.EncodeElement(s.RootCellStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootCellStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootTableStyleGroup != nil {
		if err := e.EncodeElement(s.RootTableStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootTableStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootObjectStyleGroup != nil {
		if err := e.EncodeElement(s.RootObjectStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootObjectStyleGroup"}}); err != nil {
			return err
		}
	}

	// Codificar los estilos de tabla de contenidos
	for _, toc := range s.TOCStyles {
		if err := e.EncodeElement(&toc, xml.StartElement{Name: xml.Name{Local: "TOCStyle"}}); err != nil {
			return err
		}
	}

	// Codificar los demás elementos
	for _, elem := range s.OtherElements {
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
