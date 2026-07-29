package xmlutil

import (
	"bytes"
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ProcessingInstruction representa una instrucción de procesamiento XML como <?aid ...?>
type ProcessingInstruction struct {
	Target string // ej: "aid"
	Inst   string // ej: 'style="50" type="document" ...'
}

// Metadata contiene los metadatos XML que deben preservarse durante el parseo/serialización.
type Metadata struct {
	XMLDeclaration         string                  // ej: 'version="1.0" encoding="UTF-8" standalone="yes"'
	ProcessingInstructions []ProcessingInstruction // Instrucciones de procesamiento como <?aid ...?>
	NamespaceDeclarations  map[string]string       // Mapeo de prefijo de namespace a URI
}

// ParseWithMetadata parsea datos XML en la interfaz proporcionada preservando los metadatos.
// Retorna los datos parseados y los metadatos extraídos (instrucciones de procesamiento, etc.).
func ParseWithMetadata(data []byte, v interface{}) (*Metadata, error) {
	// Add nil checks for input parameters
	if data == nil {
		return nil, common.Errorf("xmlutil", "parse with metadata", "", "input data is nil")
	}

	if len(data) == 0 {
		return nil, common.Errorf("xmlutil", "parse with metadata", "", "input data is empty")
	}

	if v == nil {
		return nil, common.Errorf("xmlutil", "parse with metadata", "", "target interface is nil")
	}

	// Parsear los datos normalmente primero
	if err := xml.Unmarshal(data, v); err != nil {
		return nil, common.WrapError("xmlutil", "parse with metadata", err)
	}

	// Extraer metadatos
	metadata := &Metadata{
		NamespaceDeclarations: make(map[string]string),
	}

	// Extraer la declaración XML
	xmlDeclRegex := regexp.MustCompile(`<\?xml\s+([^?]*)\?>`)
	if match := xmlDeclRegex.FindSubmatch(data); match != nil {
		metadata.XMLDeclaration = string(match[1])
	}

	// Extraer instrucciones de procesamiento (excluyendo la declaración xml)
	piRegex := regexp.MustCompile(`<\?(\w+)\s+([^?]+)\?>`)
	matches := piRegex.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			target := string(match[1])
			// Omitir la declaración xml (ya fue capturada)
			if target == "xml" {
				continue
			}
			// Eliminar espacios en blanco al final de la instrucción
			inst := strings.TrimRight(string(match[2]), " \t")
			metadata.ProcessingInstructions = append(metadata.ProcessingInstructions, ProcessingInstruction{
				Target: target,
				Inst:   inst,
			})
		}
	}

	// Extraer declaraciones de namespace del elemento raíz
	rootElemRegex := regexp.MustCompile(`<[^>]*xmlns:([^=]+)="([^"]*)"[^>]*>`)
	nsMatches := rootElemRegex.FindAllSubmatch(data, -1)
	for _, match := range nsMatches {
		if len(match) >= 3 {
			prefix := string(match[1])
			uri := string(match[2])
			metadata.NamespaceDeclarations[prefix] = uri
		}
	}

	return metadata, nil
}

// MarshalWithMetadata serializa la interfaz proporcionada a XML preservando los metadatos.
// Los metadatos incluyen instrucciones de procesamiento y declaraciones de namespace.
func MarshalWithMetadata(v interface{}, metadata *Metadata) ([]byte, error) {
	var buf bytes.Buffer

	// Agregar declaración XML (usar la preservada si está disponible, si no usar la predeterminada)
	if metadata != nil && metadata.XMLDeclaration != "" {
		buf.WriteString(`<?xml `)
		buf.WriteString(metadata.XMLDeclaration)
		buf.WriteString(`?>`)
	} else {
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	}
	buf.WriteByte('\n')

	// Agregar instrucciones de procesamiento
	if metadata != nil {
		for _, pi := range metadata.ProcessingInstructions {
			buf.WriteString(`<?`)
			buf.WriteString(pi.Target)
			buf.WriteByte(' ')
			buf.WriteString(pi.Inst)
			buf.WriteString(` ?>`)
			buf.WriteByte('\n')
		}
	}

	// Serializar el contenido (SIN el prefijo xml.Header ya que agregamos el nuestro)
	contentXML, err := xml.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, common.WrapError("xmlutil", "marshal with metadata", err)
	}

	// Aplicar correcciones de prefijos de namespace si es necesario
	if metadata != nil && len(metadata.NamespaceDeclarations) > 0 {
		contentXML = fixNamespacePrefixes(contentXML, metadata.NamespaceDeclarations)
	}

	// Convertir elementos vacíos a etiquetas auto-cerradas (formato IDML)
	contentXML = CompactEmptyElements(contentXML)

	buf.Write(contentXML)
	return buf.Bytes(), nil
}

// MarshalWithHeader serializa la interfaz proporcionada a XML con el encabezado XML estándar.
// Es una función de conveniencia para casos donde no se necesitan metadatos especiales.
func MarshalWithHeader(v interface{}) ([]byte, error) {
	return MarshalWithMetadata(v, nil)
}

// MarshalIndentWithHeader serializa la interfaz proporcionada a XML con indentación personalizada y encabezado.
func MarshalIndentWithHeader(v interface{}, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer

	// Agregar declaración XML
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')

	// Serializar con indentación personalizada
	contentXML, err := xml.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, common.WrapError("xmlutil", "marshal indent with header", err)
	}

	// Convertir elementos vacíos a etiquetas auto-cerradas
	contentXML = CompactEmptyElements(contentXML)

	buf.Write(contentXML)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// fixNamespacePrefixes convierte declaraciones de namespace completas de vuelta al formato con prefijo.
// Maneja el patrón común de IDML donde xml.Marshal de Go agrega xmlns="..."
// a cada elemento, pero IDML espera elementos con prefijo como idPkg:ElementName.
func fixNamespacePrefixes(xmlData []byte, namespaceDeclarations map[string]string) []byte {
	xmlStr := string(xmlData)

	// CORRECCIÓN CRÍTICA: Eliminar la declaración rota xmlns:_xmlns="xmlns"
	// y corregir _xmlns:prefix de vuelta a xmlns:prefix
	// Es un bug del serializador XML de Go al usar xml:"xmlns:prefix,attr"
	xmlStr = regexp.MustCompile(`xmlns:_xmlns="xmlns"\s+`).ReplaceAllString(xmlStr, "")

	for prefix := range namespaceDeclarations {
		xmlStr = strings.ReplaceAll(xmlStr, "_xmlns:"+prefix, "xmlns:"+prefix)
	}

	// Para cada declaración de namespace, corregir los elementos que usan ese namespace
	for prefix, nsURL := range namespaceDeclarations {
		// Corregir etiquetas de apertura con namespace
		re1 := regexp.MustCompile(`<(\w+)\s+xmlns="` + regexp.QuoteMeta(nsURL) + `"`)
		xmlStr = re1.ReplaceAllString(xmlStr, `<`+prefix+`:$1`)

		// Corregir etiquetas auto-cerradas con namespace
		re2 := regexp.MustCompile(`<(\w+)\s+xmlns="` + regexp.QuoteMeta(nsURL) + `"\s+`)
		xmlStr = re2.ReplaceAllString(xmlStr, `<`+prefix+`:$1 `)

		// Corregir etiquetas de cierre - es más complejo ya que necesitamos coincidir con las etiquetas de apertura con prefijo
		// Para IDML, manejar los elementos comunes
		if prefix == "idPkg" {
			idmlElements := []string{
				"Graphic", "Fonts", "Styles", "Preferences", "Tags",
				"MasterSpread", "Spread", "Story", "BackingStory",
			}

			for _, elem := range idmlElements {
				xmlStr = strings.ReplaceAll(xmlStr, "</"+elem+">", "</"+prefix+":"+elem+">")
			}
		}
	}

	return []byte(xmlStr)
}

// ExtractProcessingInstructions extrae instrucciones de procesamiento de datos XML.
// Es útil cuando solo se necesitan las instrucciones de procesamiento sin parsear el documento completo.
func ExtractProcessingInstructions(data []byte) ([]ProcessingInstruction, error) {
	var instructions []ProcessingInstruction

	// Extraer instrucciones de procesamiento (excluyendo la declaración xml)
	piRegex := regexp.MustCompile(`<\?(\w+)\s+([^?]+)\?>`)
	matches := piRegex.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			target := string(match[1])
			// Omitir la declaración xml
			if target == "xml" {
				continue
			}
			// Eliminar espacios en blanco al final de la instrucción
			inst := strings.TrimRight(string(match[2]), " \t")
			instructions = append(instructions, ProcessingInstruction{
				Target: target,
				Inst:   inst,
			})
		}
	}

	return instructions, nil
}

// PreserveProcessingInstructions agrega instrucciones de procesamiento a datos XML.
// Es útil para agregar instrucciones de procesamiento a XML ya serializado.
func PreserveProcessingInstructions(xmlData []byte, instructions []ProcessingInstruction) []byte {
	var buf bytes.Buffer

	// Encontrar la declaración XML y preservarla
	xmlDeclRegex := regexp.MustCompile(`<\?xml[^?]*\?>\s*`)
	xmlDecl := xmlDeclRegex.Find(xmlData)
	if xmlDecl != nil {
		buf.Write(xmlDecl)
		// Eliminar la declaración de los datos originales
		xmlData = xmlDeclRegex.ReplaceAll(xmlData, []byte{})
	}

	// Agregar instrucciones de procesamiento
	for _, pi := range instructions {
		buf.WriteString(`<?`)
		buf.WriteString(pi.Target)
		buf.WriteByte(' ')
		buf.WriteString(pi.Inst)
		buf.WriteString(` ?>`)
		buf.WriteByte('\n')
	}

	// Agregar el resto del XML
	buf.Write(bytes.TrimLeft(xmlData, " \t\n\r"))

	return buf.Bytes()
}
