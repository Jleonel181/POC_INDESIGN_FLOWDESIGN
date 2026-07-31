package resources

import (
	"encoding/xml"
	"errors"
	"io"

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

// Clases de hijo de <idPkg:Styles>. Coinciden con el nombre de la etiqueta, salvo
// la de los elementos sin modelar, que van todos al mismo campo.
const (
	styleChildRootCharacterGroup = "RootCharacterStyleGroup"
	styleChildRootParagraphGroup = "RootParagraphStyleGroup"
	styleChildRootCellGroup      = "RootCellStyleGroup"
	styleChildRootTableGroup     = "RootTableStyleGroup"
	styleChildRootObjectGroup    = "RootObjectStyleGroup"
	styleChildTOCStyle           = "TOCStyle"
	styleChildOther              = "OtherElement"
)

// stylesChildOrder es el orden en que están declarados los campos del struct, que
// es el que se usa para los hijos que el registro de orden no menciona. Tiene que
// nombrar todas las clases de arriba; lo comprueba un test.
var stylesChildOrder = []string{
	styleChildRootCharacterGroup,
	styleChildRootParagraphGroup,
	styleChildRootCellGroup,
	styleChildRootTableGroup,
	styleChildRootObjectGroup,
	styleChildTOCStyle,
	styleChildOther,
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

	// Los hijos se recorren de uno en uno, y no con un struct temporal como antes,
	// porque hay que registrar en qué orden vienen: encoding/xml los reparte por
	// campos y pierde la secuencia.
	for {
		token, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return common.WrapError("resources", "unmarshal styles content", err)
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if err := s.unmarshalChild(d, elem); err != nil {
				return err
			}
		case xml.EndElement:
			return nil
		}
	}

	return nil
}

// unmarshalChild decodifica un hijo de <idPkg:Styles> en su campo y anota su clase
// en el registro de orden.
func (s *StylesFile) unmarshalChild(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case styleChildRootCharacterGroup:
		var group CharacterStyleGroup
		if err := d.DecodeElement(&group, &start); err != nil {
			return common.WrapError("resources", "unmarshal styles content", err)
		}
		s.RootCharacterStyleGroup = &group

	case styleChildRootParagraphGroup:
		var group ParagraphStyleGroup
		if err := d.DecodeElement(&group, &start); err != nil {
			return common.WrapError("resources", "unmarshal styles content", err)
		}
		s.RootParagraphStyleGroup = &group

	case styleChildRootCellGroup:
		var group CellStyleGroup
		if err := d.DecodeElement(&group, &start); err != nil {
			return common.WrapError("resources", "unmarshal styles content", err)
		}
		s.RootCellStyleGroup = &group

	case styleChildRootTableGroup:
		var group TableStyleGroup
		if err := d.DecodeElement(&group, &start); err != nil {
			return common.WrapError("resources", "unmarshal styles content", err)
		}
		s.RootTableStyleGroup = &group

	case styleChildRootObjectGroup:
		var group ObjectStyleGroup
		if err := d.DecodeElement(&group, &start); err != nil {
			return common.WrapError("resources", "unmarshal styles content", err)
		}
		s.RootObjectStyleGroup = &group

	case styleChildTOCStyle:
		var toc TOCStyle
		if err := d.DecodeElement(&toc, &start); err != nil {
			return common.WrapError("resources", "unmarshal styles content", err)
		}
		s.TOCStyles = append(s.TOCStyles, toc)

	default:
		var raw common.RawXMLElement
		if err := d.DecodeElement(&raw, &start); err != nil {
			return common.WrapErrorWithPath("resources", "unmarshal styles content", start.Name.Local, err)
		}
		s.OtherElements = append(s.OtherElements, raw)
		s.childOrder.Record(styleChildOther)
		return nil
	}

	s.childOrder.Record(start.Name.Local)
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

	// Emitir los hijos en el orden en que venían al parsear, o en el orden de los
	// campos si el StylesFile se construyó desde cero.
	if err := s.childOrder.Replay(stylesChildOrder, s.childrenByKind(e)); err != nil {
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
func (s *StylesFile) childrenByKind(e *xml.Encoder) map[string][]xmlutil.ChildEmitter {
	children := make(map[string][]xmlutil.ChildEmitter, len(stylesChildOrder))

	named := func(kind string, child any) xmlutil.ChildEmitter {
		return func() error {
			return e.EncodeElement(child, xml.StartElement{Name: xml.Name{Local: kind}})
		}
	}
	one := func(kind string, child any, present bool) {
		if present {
			children[kind] = []xmlutil.ChildEmitter{named(kind, child)}
		}
	}

	one(styleChildRootCharacterGroup, s.RootCharacterStyleGroup, s.RootCharacterStyleGroup != nil)
	one(styleChildRootParagraphGroup, s.RootParagraphStyleGroup, s.RootParagraphStyleGroup != nil)
	one(styleChildRootCellGroup, s.RootCellStyleGroup, s.RootCellStyleGroup != nil)
	one(styleChildRootTableGroup, s.RootTableStyleGroup, s.RootTableStyleGroup != nil)
	one(styleChildRootObjectGroup, s.RootObjectStyleGroup, s.RootObjectStyleGroup != nil)

	for i := range s.TOCStyles {
		children[styleChildTOCStyle] = append(children[styleChildTOCStyle], named(styleChildTOCStyle, &s.TOCStyles[i]))
	}

	// Los elementos sin modelar se emiten con el nombre que traían.
	for i := range s.OtherElements {
		elem := &s.OtherElements[i]
		children[styleChildOther] = append(children[styleChildOther], func() error {
			return e.EncodeElement(elem, xml.StartElement{Name: elem.XMLName})
		})
	}

	return children
}
