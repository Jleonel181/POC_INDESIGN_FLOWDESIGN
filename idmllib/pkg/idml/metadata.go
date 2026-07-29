package idml

// MetadataFile representa archivos de metadatos opcionales en un paquete IDML.
// Estos archivos incluyen:
//   - META-INF/container.xml: Metadatos del paquete (especificación de contenedor OASIS)
//   - META-INF/metadata.xml: Metadatos XMP (Dublin Core, fechas de creación, etc.)
//   - XML/Tags.xml: Definiciones de etiquetas XML para contenido estructurado
//   - XML/BackingStory.xml: Story predeterminada para contenido XML
//
// Se usa una estrategia de preservación (almacenando el XML crudo) para garantizar
// roundtrips perfectos sin necesidad de modelar cada estructura XMP/RDF compleja.
type MetadataFile struct {
	// Filename es la ruta dentro del paquete IDML (ej. "META-INF/container.xml")
	Filename string

	// RawContent almacena el contenido completo del archivo tal como está
	// Preserva toda la estructura XML, namespaces y formato
	RawContent []byte
}

// ParseMetadataFile parsea un archivo de metadatos usando la estrategia de preservación.
// El contenido del archivo se almacena tal como está para fidelidad perfecta de roundtrip.
func ParseMetadataFile(filename string, data []byte) (*MetadataFile, error) {
	return &MetadataFile{
		Filename:   filename,
		RawContent: data,
	}, nil
}

// MarshalMetadataFile serializa un archivo de metadatos de vuelta a su forma original.
// Dado que se usa la estrategia de preservación, simplemente retorna el contenido crudo.
func MarshalMetadataFile(mf *MetadataFile) ([]byte, error) {
	return mf.RawContent, nil
}
