package xmlutil

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NamespaceConfig contiene la configuración para el manejo de namespaces.
type NamespaceConfig struct {
	Prefix string // Prefijo del namespace (ej: "idPkg")
	URI    string // URI del namespace (ej: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging")
}

// IDMLNamespace retorna la configuración estándar del namespace de packaging IDML.
func IDMLNamespace() NamespaceConfig {
	return NamespaceConfig{
		Prefix: "idPkg",
		URI:    "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
	}
}

// ParseWithNamespace parsea datos XML en la interfaz proporcionada, manejando envoltorios de namespace.
// Está diseñado para archivos IDML que usan envoltorios de namespace idPkg alrededor del contenido.
//
// La función espera XML en el formato:
//
//	<idPkg:ElementName xmlns:idPkg="..." DOMVersion="...">
//	  <ElementName>...</ElementName>
//	</idPkg:ElementName>
//
// Extrae el DOMVersion y hace unmarshal del contenido interno en la interfaz proporcionada.
func ParseWithNamespace(data []byte, v interface{}, config NamespaceConfig) (string, error) {
	// Validar parámetros de entrada
	if data == nil {
		return "", common.Errorf("xmlutil", "parse with namespace", "", "input data is nil")
	}

	if len(data) == 0 {
		return "", common.Errorf("xmlutil", "parse with namespace", "", "input data is empty")
	}

	if v == nil {
		return "", common.Errorf("xmlutil", "parse with namespace", "", "target interface is nil")
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))

	var domVersion string

	// Buscar el elemento envoltorio
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", fmt.Errorf("failed to find namespace wrapper: %w", err)
		}

		if start, ok := token.(xml.StartElement); ok {
			expectedName := config.Prefix + ":" + getElementName(v)
			if start.Name.Local == expectedName ||
				(start.Name.Local == getElementName(v) && start.Name.Space == config.URI) {

				// Extraer el atributo DOMVersion
				for _, attr := range start.Attr {
					if attr.Name.Local == "DOMVersion" {
						domVersion = attr.Value
						break
					}
				}

				// Parsear el contenido usando el decoder posicionado en el envoltorio
				if err := decoder.DecodeElement(v, &start); err != nil {
					return "", fmt.Errorf("failed to decode element content: %w", err)
				}

				return domVersion, nil
			}
		}
	}
}

// MarshalWithNamespace serializa la interfaz proporcionada a XML con envoltorio de namespace.
// Genera XML en el formato IDML con envoltorio de namespace idPkg.
//
// El formato de salida es:
//
//	<idPkg:ElementName xmlns:idPkg="..." DOMVersion="...">
//	  <ElementName>...</ElementName>
//	</idPkg:ElementName>
func MarshalWithNamespace(v interface{}, config NamespaceConfig, domVersion string) ([]byte, error) {
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	elementName := getElementName(v)
	wrapperName := config.Prefix + ":" + elementName

	// Crear el elemento envoltorio
	wrapper := xml.StartElement{
		Name: xml.Name{Local: wrapperName},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns:" + config.Prefix},
				Value: config.URI,
			},
			{
				Name:  xml.Name{Local: "DOMVersion"},
				Value: domVersion,
			},
		},
	}

	// Iniciar el elemento envoltorio
	if err := encoder.EncodeToken(wrapper); err != nil {
		return nil, fmt.Errorf("failed to encode wrapper start: %w", err)
	}

	// Codificar el contenido interno
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to encode inner content: %w", err)
	}

	// Cerrar el elemento envoltorio
	if err := encoder.EncodeToken(wrapper.End()); err != nil {
		return nil, fmt.Errorf("failed to encode wrapper end: %w", err)
	}

	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// ApplyNamespace aplica la configuración de namespace a datos XML crudos.
// Es útil para transformar XML que no tiene declaraciones de namespace correctas.
func ApplyNamespace(data []byte, config NamespaceConfig) ([]byte, error) {
	// Parsear el XML para verificar que es válido
	var temp interface{}
	if err := xml.Unmarshal(data, &temp); err != nil {
		return nil, fmt.Errorf("invalid XML data: %w", err)
	}

	// Por ahora retornar los datos tal cual, ya que esta es una transformación compleja
	// que requeriría parsear y reconstruir toda la estructura XML.
	// Puede mejorarse más adelante si es necesario.
	return data, nil
}

// getElementName extrae el nombre del elemento XML de un tipo struct.
// Usa reflection para encontrar el tag xml o el nombre del struct.
func getElementName(v interface{}) string {
	// Implementación simplificada que asume que el nombre del elemento
	// coincide con patrones comunes de IDML. Una versión más sofisticada
	// usaría reflection para examinar los tags xml.

	switch v.(type) {
	case *interface{}:
		// Para interfaces genéricas no es posible determinar el nombre
		return "Element"
	default:
		// Por ahora retornar un nombre genérico. Debería mejorarse
		// para usar reflection y obtener el nombre real del struct o tag xml.
		return "Element"
	}
}

// NamespaceWrapper provee una forma genérica de manejar XML con envoltorio de namespace.
type NamespaceWrapper struct {
	XMLName    xml.Name `xml:""`
	DOMVersion string   `xml:"DOMVersion,attr"`
	Content    []byte   `xml:",innerxml"`
}

// ParseNamespaceWrapper parsea XML con un envoltorio de namespace y retorna el contenido interno.
func ParseNamespaceWrapper(data []byte) (*NamespaceWrapper, error) {
	var wrapper NamespaceWrapper
	if err := xml.Unmarshal(data, &wrapper); err != nil {
		return nil, common.WrapError("xmlutil", "parse namespace wrapper", err)
	}
	return &wrapper, nil
}

// MarshalNamespaceWrapper crea XML con un envoltorio de namespace alrededor del contenido proporcionado.
func MarshalNamespaceWrapper(elementName string, config NamespaceConfig, domVersion string, content []byte) ([]byte, error) {
	wrapper := NamespaceWrapper{
		XMLName:    xml.Name{Local: config.Prefix + ":" + elementName},
		DOMVersion: domVersion,
		Content:    content,
	}

	data, err := xml.Marshal(wrapper)
	if err != nil {
		return nil, common.WrapError("xmlutil", "marshal namespace wrapper", err)
	}

	return data, nil
}
