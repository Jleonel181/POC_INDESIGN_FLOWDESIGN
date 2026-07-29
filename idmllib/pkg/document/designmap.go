package document

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Designmap representa el elemento principal Document en designmap.xml.
// Es un struct mínimo que preserva elementos desconocidos para compatibilidad hacia adelante.
//
// DEPRECADO: Implementación de Fase 1. Usar el struct Document en document.go para Fase 2+.
// Este struct se mantiene solo para compatibilidad hacia atrás y propósitos de testing.
type Designmap struct {
	// XMLName captura el nombre y namespace del elemento.
	// Nota: El namespace por defecto NO es el mismo que el namespace idPkg.
	XMLName xml.Name `xml:"Document"`

	// Xmlns define el prefijo de namespace idPkg.
	// Es requerido para el manejo correcto de namespaces XML.
	// Usar el nombre completo del atributo asegura un marshaling correcto.
	Xmlns string `xml:"xmlns:idPkg,attr"`

	// Atributos esenciales que identifican el documento
	DOMVersion string `xml:"DOMVersion,attr,omitempty"`
	Self       string `xml:"Self,attr,omitempty"`
	Version    string `xml:"Version,attr,omitempty"`

	// Comodín para todos los elementos hijo. La Fase 2 usa campos de struct explícitos en su lugar.
	OtherElements []common.RawXMLElement `xml:",any"`
}

// DesignmapMinimal representa un struct aún más mínimo para testing.
// Esta versión modela explícitamente algunos elementos hijo comunes mientras
// preserva los desconocidos.
//
// DEPRECADO: Struct de testing de Fase 1. Usar Document para Fase 2+.
type DesignmapMinimal struct {
	XMLName    xml.Name `xml:"Document"`
	Xmlns      string   `xml:"xmlns:idPkg,attr,omitempty"`
	DOMVersion string   `xml:"DOMVersion,attr,omitempty"`
	Self       string   `xml:"Self,attr,omitempty"`
	Version    string   `xml:"Version,attr,omitempty"`

	// Properties es un elemento común en el XML de InDesign
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Elementos Language (puede haber múltiples)
	Languages []Language `xml:"Language,omitempty"`

	// Referencia Graphic usando el namespace idPkg
	// Nota: el namespace completo en el tag
	Graphic *GraphicRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`

	// Comodín para todos los demás elementos desconocidos
	OtherElements []common.RawXMLElement `xml:",any"`
}

// GraphicRef representa un elemento de referencia idPkg:Graphic.
// DEPRECADO: Struct de Fase 1. Usar ResourceRef en document.go para Fase 2+.
type GraphicRef struct {
	XMLName xml.Name `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic"`
	Src     string   `xml:"src,attr"`
}

// ParseDesignmap parsea un archivo designmap.xml en un struct Designmap.
// Es un parseo mínimo que preserva todo el contenido.
//
// Nota: encoding/xml de Go no popula automáticamente las declaraciones de namespace
// en campos de struct, por eso extraemos xmlns:idPkg manualmente si está presente.
func ParseDesignmap(data []byte) (*Designmap, error) {
	var dm Designmap
	if err := xml.Unmarshal(data, &dm); err != nil {
		return nil, err
	}

	// Extraer manualmente el namespace xmlns:idPkg si está presente en el XML
	// Esto es necesario porque encoding/xml no popula declaraciones de namespace
	dataStr := string(data)
	if idx := findSubstringIndex(dataStr, "xmlns:idPkg="); idx != -1 {
		// Encontrar la comilla después de xmlns:idPkg=
		start := idx + len("xmlns:idPkg=")
		if start < len(dataStr) && (dataStr[start] == '"' || dataStr[start] == '\'') {
			quote := dataStr[start]
			start++
			// Encontrar la comilla de cierre
			end := start
			for end < len(dataStr) && dataStr[end] != byte(quote) {
				end++
			}
			if end < len(dataStr) {
				dm.Xmlns = dataStr[start:end]
			}
		}
	}

	return &dm, nil
}

// findSubstringIndex encuentra el índice de substr en s, retorna -1 si no se encuentra.
func findSubstringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// MarshalDesignmap serializa un struct Designmap de vuelta a bytes XML.
func MarshalDesignmap(dm *Designmap) ([]byte, error) {
	// Agregar encabezado XML
	header := []byte(xml.Header)

	data, err := xml.MarshalIndent(dm, "", "\t")
	if err != nil {
		return nil, err
	}

	// Combinar encabezado y datos
	result := append(header, data...)
	return result, nil
}
