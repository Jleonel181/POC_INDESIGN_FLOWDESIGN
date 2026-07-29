package document

import (
	"github.com/dimelords/idmllib/v2/internal/xmlutil"
)

// ProcessingInstruction representa una instrucción de procesamiento XML como <?aid ...?>
type ProcessingInstruction struct {
	Target string // ej: "aid"
	Inst   string // ej: 'style="50" type="document" ...'
}

// DocumentWithMetadata envuelve Document con metadatos adicionales que no
// encajan en el flujo estándar de xml.Unmarshal/Marshal.
type DocumentWithMetadata struct {
	*Document
	XMLDeclaration         string // ej: '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
	ProcessingInstructions []ProcessingInstruction
}

// ParseDocumentWithMetadata parsea designmap.xml y preserva las instrucciones de procesamiento.
func ParseDocumentWithMetadata(data []byte) (*DocumentWithMetadata, error) {
	// Primero, parsear el documento normalmente
	doc, err := ParseDocument(data)
	if err != nil {
		return nil, err
	}

	// Extraer metadatos usando utilidades compartidas
	metadata, err := xmlutil.ParseWithMetadata(data, &struct{}{}) // El doc ya fue parseado, solo necesitamos los metadatos
	if err != nil {
		return nil, err
	}

	result := &DocumentWithMetadata{
		Document:               doc,
		XMLDeclaration:         metadata.XMLDeclaration,
		ProcessingInstructions: make([]ProcessingInstruction, len(metadata.ProcessingInstructions)),
	}

	// Convertir de xmlutil.ProcessingInstruction a document.ProcessingInstruction
	for i, pi := range metadata.ProcessingInstructions {
		result.ProcessingInstructions[i] = ProcessingInstruction{
			Target: pi.Target,
			Inst:   pi.Inst,
		}
	}

	return result, nil
}

// MarshalDocumentWithMetadata serializa un Document a XML preservando los metadatos.
func MarshalDocumentWithMetadata(docMeta *DocumentWithMetadata) ([]byte, error) {
	// Convertir document.ProcessingInstruction a xmlutil.ProcessingInstruction
	metadata := &xmlutil.Metadata{
		XMLDeclaration:         docMeta.XMLDeclaration,
		ProcessingInstructions: make([]xmlutil.ProcessingInstruction, len(docMeta.ProcessingInstructions)),
		NamespaceDeclarations:  map[string]string{"idPkg": "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
	}

	for i, pi := range docMeta.ProcessingInstructions {
		metadata.ProcessingInstructions[i] = xmlutil.ProcessingInstruction{
			Target: pi.Target,
			Inst:   pi.Inst,
		}
	}

	// Usar utilidad compartida para serializar con metadatos
	return xmlutil.MarshalWithMetadata(docMeta.Document, metadata)
}
