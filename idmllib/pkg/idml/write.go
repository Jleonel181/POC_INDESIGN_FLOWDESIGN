package idml

import (
	"archive/zip"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// marshalCachedObjects serializa todos los objetos en caché de vuelta a datos XML.
// Esto asegura que cualquier modificación a los structs parseados sea guardada.
func (p *Package) marshalCachedObjects() error {
	// Actualizar metadatos XMP en META-INF/metadata.xml si existe y ha sido modificado
	if err := p.updateXMPInMetadataFile(); err != nil {
		return err
	}

	// Si el documento fue parseado, serializarlo de vuelta a XML con metadatos preservados
	if p.documentMetadata != nil {
		xmlData, err := document.MarshalDocumentWithMetadata(p.documentMetadata)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal document", PathDesignmap, err)
		}
		p.setFileData(PathDesignmap, xmlData)
	}

	// Si las stories fueron parseadas, serializarlas de vuelta a XML
	for filename, st := range p.stories {
		xmlData, err := story.MarshalStory(st)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal story", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// Si los spreads fueron parseados, serializarlos de vuelta a XML
	for filename, sp := range p.spreads {
		xmlData, err := spread.MarshalSpread(sp)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal spread", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// Si los resources fueron parseados, serializarlos de vuelta a XML
	for filename, resource := range p.resources {
		xmlData, err := MarshalResourceFile(resource)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal resource", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// Si las fuentes tipadas fueron parseadas, serializarlas de vuelta a XML
	if p.fonts != nil {
		xmlData, err := resources.MarshalFontsFile(p.fonts)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal fonts", PathFonts, err)
		}
		p.setFileData(PathFonts, xmlData)
	}

	// Si los gráficos tipados fueron parseados, serializarlos de vuelta a XML
	if p.graphics != nil {
		xmlData, err := resources.MarshalGraphicFile(p.graphics)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal graphics", PathGraphic, err)
		}
		p.setFileData(PathGraphic, xmlData)
	}

	// Si los estilos tipados fueron parseados, serializarlos de vuelta a XML
	if p.styles != nil {
		xmlData, err := resources.MarshalStylesFile(p.styles)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal styles", PathStyles, err)
		}
		p.setFileData(PathStyles, xmlData)
	}

	// Si los archivos de metadata fueron parseados, serializarlos de vuelta
	for filename, metadata := range p.metadata {
		data, err := MarshalMetadataFile(metadata)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal metadata", filename, err)
		}
		p.setFileData(filename, data)
	}

	return nil
}

// updateXMPInMetadataFile actualiza el paquete XMP en META-INF/metadata.xml.
// Esto asegura que las modificaciones XMP sean persistidas al escribir el package.
func (p *Package) updateXMPInMetadataFile() error {
	// Verificar si metadata.xml existe
	entry, err := p.getFileEntry("META-INF/metadata.xml")
	if err != nil {
		// Si metadata.xml no existe, no hay nada que actualizar
		return nil
	}

	// Obtener el contenido actual
	content := string(entry.data)

	// Reemplazar el paquete XMP con el actualizado
	xmpPattern := regexp.MustCompile(`(?s)<\?xpacket begin.*?<\?xpacket end[^>]*\?>`)

	if p.XMPMetadata != "" {
		// Reemplazar XMP existente o agregar si no está presente
		if xmpPattern.MatchString(content) {
			content = xmpPattern.ReplaceAllString(content, p.XMPMetadata)
		} else {
			// Si no existe XMP, agregarlo antes del tag de cierre
			// Encontrar un punto de inserción adecuado (antes de </rdf:RDF> o al final)
			if idx := strings.Index(content, "</rdf:RDF>"); idx != -1 {
				content = content[:idx] + p.XMPMetadata + "\n" + content[idx:]
			} else {
				content = content + "\n" + p.XMPMetadata
			}
		}
	} else {
		// Si XMPMetadata está vacío, no tocamos el archivo: el contenido original de la
		// plantilla se preserva. Solo se borra si se asigna explícitamente una cadena
		// vacía después de haber tenido contenido (caso de uso: RemoveXMP).
		return nil
	}

	// Actualizar los datos del archivo
	p.setFileData("META-INF/metadata.xml", []byte(content))
	return nil
}

// writeZipFiles escribe todos los archivos al archivo ZIP en el orden correcto.
func writeZipFiles(w *zip.Writer, pkg *Package) error {
	// Agregar validación de parámetros
	if w == nil {
		return common.Errorf("idml", "write zip files", "", "zip writer is nil")
	}

	if pkg == nil {
		return common.Errorf("idml", "write zip files", "", "package is nil")
	}

	// CRÍTICO: Escribir mimetype primero y sin compresión
	if entry, err := pkg.getFileEntry(PathMimetype); err == nil {
		// Validar la entrada antes de usarla
		if entry == nil || entry.header == nil {
			return common.WrapErrorWithPath("idml", "write", PathMimetype, common.Errorf("idml", "write", PathMimetype, "invalid mimetype entry"))
		}

		// Siempre usar el método Store para mimetype (requisito CRÍTICO)
		// Crear una copia del header para evitar modificar el original
		header := *entry.header
		header.Method = zip.Store // Forzar sin compresión

		mimeWriter, err := w.CreateHeader(&header)
		if err != nil {
			return common.WrapErrorWithPath("idml", "write", PathMimetype, err)
		}

		// Escribir el contenido del mimetype
		if _, err := mimeWriter.Write(entry.data); err != nil {
			return common.WrapErrorWithPath("idml", "write", PathMimetype, err)
		}
	}

	// Escribir todos los demás archivos en el orden original
	for _, name := range pkg.fileOrder {
		if name == PathMimetype {
			continue // Ya fue escrito
		}

		entry, err := pkg.getFileEntry(name)
		if err != nil {
			continue // Omitir archivos faltantes
		}

		// Crear header si no existe (por ejemplo, para archivos recién agregados)
		if entry.header == nil {
			entry.header = &zip.FileHeader{
				Name:     name,
				Method:   zip.Deflate, // Usar compresión para todos los archivos excepto mimetype
				Modified: time.Now(),
			}
		}

		// Usar el FileHeader original para preservar la compresión y los metadatos
		fileWriter, err := w.CreateHeader(entry.header)
		if err != nil {
			return common.WrapErrorWithPath("idml", "write", name, err)
		}

		if _, err := fileWriter.Write(entry.data); err != nil {
			return common.WrapErrorWithPath("idml", "write", name, err)
		}
	}

	return nil
}

// WriteTo escribe el paquete IDML en un io.Writer arbitrario (stdout, un buffer,
// una conexión HTTP, etc.).
//
// Serializa los objetos en caché, escribe mimetype primero sin compresión, y emite
// el resto en el orden original. El caller es responsable de cerrar el writer si
// aplica.
func WriteTo(pkg *Package, w io.Writer) error {
	if err := pkg.marshalCachedObjects(); err != nil {
		return err
	}

	zw := zip.NewWriter(w)

	if err := writeZipFiles(zw, pkg); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return common.WrapError("idml", "write to stream", err)
	}

	return nil
}

// Write escribe un package IDML en un archivo.
//
// La función:
// 1. Serializa el struct Document de vuelta a designmap.xml (si fue modificado)
// 2. Escribe mimetype primero y sin compresión (requisito CRÍTICO de IDML)
// 3. Escribe todos los demás archivos en el orden original
//
// CRÍTICO: El archivo mimetype DEBE escribirse primero y DEBE estar sin compresión.
// Esto es requerido por la especificación IDML. InDesign rechazará archivos
// que no cumplan este requisito.
func Write(pkg *Package, path string) error {
	// #nosec G304 - Esta es una función de librería; la ruta del archivo es proporcionada intencionalmente por el llamador
	f, err := os.Create(path)
	if err != nil {
		return common.WrapErrorWithPath("idml", "write", path, err)
	}
	defer f.Close()

	if err := WriteTo(pkg, f); err != nil {
		return err
	}

	return nil
}
