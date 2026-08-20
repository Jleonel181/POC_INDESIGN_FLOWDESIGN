package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// UnmarshalXML implementa deserialización XML custom para Spread.
func (s *Spread) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Agregar verificación nil para el decoder
	if d == nil {
		return common.Errorf("spread", "unmarshal spread", "", "decoder is nil")
	}

	// Verificar que estamos parseando un elemento idPkg:Spread
	if start.Name.Local != "Spread" {
		return common.WrapError("spread", "parse", common.ErrInvalidFormat)
	}

	// Extraer DOMVersion desde los atributos de idPkg:Spread
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			s.DOMVersion = attr.Value
			break
		}
	}

	// Leer tokens hasta encontrar el elemento <Spread> interno o llegar al final
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "Spread" {
				// Encontrado el elemento Spread interno - deserializarlo
				if err := d.DecodeElement(&s.InnerSpread, &t); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if t.Name.Local == "Spread" && t.Name.Space == start.Name.Space {
				// Fin del elemento idPkg:Spread
				return nil
			}
		}
	}
}

// MarshalXML implementa serialización XML custom para Spread.
func (s *Spread) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Crear el elemento wrapper idPkg:Spread
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Spread"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: s.DOMVersion},
		},
	}

	// Iniciar el elemento wrapper
	if err := e.EncodeToken(wrapper); err != nil {
		return common.WrapError("spread", "marshal spread", err)
	}

	// Serializar el elemento Spread interno
	innerStart := xml.StartElement{Name: xml.Name{Local: "Spread"}}
	if err := e.EncodeElement(&s.InnerSpread, innerStart); err != nil {
		return common.WrapError("spread", "marshal spread", err)
	}

	// Finalizar el elemento wrapper
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return common.WrapError("spread", "marshal spread", err)
	}

	return nil
}

// ParseSpread parsea un archivo XML de Spread en un struct Spread.
func ParseSpread(data []byte) (*Spread, error) {
	// Agregar verificación nil para los datos de entrada
	if data == nil {
		return nil, common.Errorf("spread", "parse spread", "", "input data is nil")
	}

	// Agregar verificación de datos vacíos
	if len(data) == 0 {
		return nil, common.Errorf("spread", "parse spread", "", "input data is empty")
	}

	var spread Spread
	if err := xml.Unmarshal(data, &spread); err != nil {
		return nil, common.WrapError("spread", "parse spread", err)
	}
	return &spread, nil
}

// MarshalSpread serializa un struct Spread de vuelta a XML con formato correcto.
func MarshalSpread(spread *Spread) ([]byte, error) {
	// Agregar declaración XML y serializar con indentación
	data, err := xmlutil.MarshalIndentWithHeader(spread, "", "\t")
	if err != nil {
		return nil, common.WrapError("spread", "marshal spread", err)
	}

	return data, nil
}

// ParseMasterSpread parsea un archivo XML de MasterSpread en un struct MasterSpread.
func ParseMasterSpread(data []byte) (*MasterSpread, error) {
	if data == nil {
		return nil, common.Errorf("spread", "parse master spread", "", "input data is nil")
	}
	if len(data) == 0 {
		return nil, common.Errorf("spread", "parse master spread", "", "input data is empty")
	}

	var ms MasterSpread
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, common.WrapError("spread", "parse master spread", err)
	}
	return &ms, nil
}

// MarshalMasterSpread serializa un struct MasterSpread de vuelta a XML con formato correcto.
func MarshalMasterSpread(ms *MasterSpread) ([]byte, error) {
	data, err := xmlutil.MarshalIndentWithHeader(ms, "", "\t")
	if err != nil {
		return nil, common.WrapError("spread", "marshal master spread", err)
	}
	return data, nil
}
