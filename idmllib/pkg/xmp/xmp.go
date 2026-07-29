// Package xmp provee soporte de metadatos XMP (Extensible Metadata Platform)
// para archivos IDML e IDMS. XMP es el estándar de Adobe para incrustar metadatos
// en documentos.
package xmp

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Metadata representa metadatos XMP que pueden ser incrustados en archivos IDML o IDMS.
// XMP es el estándar de Adobe para incrustar metadatos en documentos.
type Metadata struct {
	raw string // El paquete XMP completo incluyendo las instrucciones de procesamiento
}

// Parse crea una instancia de Metadata XMP a partir de un string XMP crudo.
// El string XMP típicamente incluye instrucciones de procesamiento <?xpacket begin...?>
// y <?xpacket end...?> envolviendo el elemento <x:xmpmeta>.
//
// Parse acepta cualquier string, incluyendo strings vacíos, y retorna una instancia
// *Metadata no nula en todos los casos.
func Parse(xmpString string) *Metadata {
	return &Metadata{raw: xmpString}
}

// String retorna los metadatos XMP como string, incluyendo las instrucciones de procesamiento.
// Permite recuperar el XMP modificado luego de realizar operaciones.
func (x *Metadata) String() string {
	return x.raw
}

// IsEmpty retorna true si no hay metadatos XMP.
// Se usa para verificar si existen metadatos XMP antes de realizar operaciones.
func (x *Metadata) IsEmpty() bool {
	return x.raw == ""
}

var (
	modifyDatePattern   = regexp.MustCompile(`<xmp:ModifyDate>[^<]+</xmp:ModifyDate>`)
	metadataDatePattern = regexp.MustCompile(`<xmp:MetadataDate>[^<]+</xmp:MetadataDate>`)
	thumbnailPattern    = regexp.MustCompile(`(?s)\s*<xmp:Thumbnails>.*?</xmp:Thumbnails>`)
)

// UpdateTimestamps actualiza xmp:ModifyDate y xmp:MetadataDate a la hora actual.
// El campo xmp:CreateDate se preserva sin cambios.
// Retorna un error si no existen metadatos XMP.
func (x *Metadata) UpdateTimestamps() error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Obtener timestamp actual en formato ISO 8601 con zona horaria
	currentTime := time.Now().Format("2006-01-02T15:04:05-07:00")

	// Actualizar xmp:ModifyDate
	x.raw = modifyDatePattern.ReplaceAllString(x.raw, "<xmp:ModifyDate>"+currentTime+"</xmp:ModifyDate>")

	// Actualizar xmp:MetadataDate
	x.raw = metadataDatePattern.ReplaceAllString(x.raw, "<xmp:MetadataDate>"+currentTime+"</xmp:MetadataDate>")

	return nil
}

// RemoveThumbnails elimina la sección xmp:Thumbnails de los metadatos.
// Es útil cuando el contenido fue modificado y los thumbnails están desactualizados.
// Eliminar thumbnails puede reducir el tamaño del archivo entre 20-30KB.
// Retorna un error si no existen metadatos XMP.
func (x *Metadata) RemoveThumbnails() error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Eliminar la sección xmp:Thumbnails completa incluyendo contenido anidado y espacios circundantes
	x.raw = thumbnailPattern.ReplaceAllString(x.raw, "")

	return nil
}

// AddThumbnail agrega un thumbnail a los metadatos XMP.
// thumbnailData debe ser una imagen JPEG codificada en base64.
// El thumbnail se agregará con dimensiones predeterminadas de 512x512.
// Retorna un error si no existen metadatos XMP.
func (x *Metadata) AddThumbnail(thumbnailData string, width, height int) error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Verificar si ya existen thumbnails y eliminarlos primero
	if thumbnailPattern.MatchString(x.raw) {
		x.raw = thumbnailPattern.ReplaceAllString(x.raw, "")
	}

	// Construir la estructura XML del thumbnail
	thumbnailXML := fmt.Sprintf(`<xmp:Thumbnails>
            <rdf:Alt>
               <rdf:li rdf:parseType="Resource">
                  <xmpGImg:format>JPEG</xmpGImg:format>
                  <xmpGImg:width>%d</xmpGImg:width>
                  <xmpGImg:height>%d</xmpGImg:height>
                  <xmpGImg:image>%s</xmpGImg:image>
               </rdf:li>
            </rdf:Alt>
         </xmp:Thumbnails>`, width, height, thumbnailData)

	// Encontrar la posición para insertar el thumbnail (después de xmp:ModifyDate o xmp:MetadataDate)
	insertPattern := regexp.MustCompile(`(</xmp:ModifyDate>|</xmp:MetadataDate>)`)
	matches := insertPattern.FindStringIndex(x.raw)
	
	if matches == nil {
		return errors.New("could not find insertion point for thumbnail")
	}

	// Insertar el thumbnail después del tag encontrado
	insertPos := matches[1]
	x.raw = x.raw[:insertPos] + "\n         " + thumbnailXML + x.raw[insertPos:]

	return nil
}

// GetField obtiene el valor de un campo XMP específico.
// fieldName debe incluir el prefijo de namespace (ej. "xmp:CreateDate").
// Retorna un error si el campo no se encuentra o si no existen metadatos XMP.
func (x *Metadata) GetField(fieldName string) (string, error) {
	if x.raw == "" {
		return "", errors.New("no XMP metadata")
	}

	// Construir patrón regex para extraer el valor del campo
	// Usar regexp.QuoteMeta para escapar caracteres especiales en el nombre del campo
	pattern := regexp.MustCompile(fmt.Sprintf(`<%s>([^<]+)</%s>`, 
		regexp.QuoteMeta(fieldName), regexp.QuoteMeta(fieldName)))
	
	matches := pattern.FindStringSubmatch(x.raw)
	if matches == nil {
		return "", fmt.Errorf("field %s not found", fieldName)
	}

	// Retornar el grupo capturado (valor del campo)
	return matches[1], nil
}

// SetField establece el valor de un campo XMP específico.
// fieldName debe incluir el prefijo de namespace (ej. "xmp:CreatorTool").
// Retorna un error si el campo no se encuentra o si no existen metadatos XMP.
func (x *Metadata) SetField(fieldName, value string) error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Construir patrón regex para encontrar y reemplazar el valor del campo
	// Usar regexp.QuoteMeta para escapar caracteres especiales en el nombre del campo
	pattern := regexp.MustCompile(fmt.Sprintf(`<%s>[^<]+</%s>`, 
		regexp.QuoteMeta(fieldName), regexp.QuoteMeta(fieldName)))
	
	// Verificar si el campo existe
	if !pattern.MatchString(x.raw) {
		return fmt.Errorf("field %s not found", fieldName)
	}

	// Reemplazar el valor del campo
	replacement := fmt.Sprintf("<%s>%s</%s>", fieldName, value, fieldName)
	x.raw = pattern.ReplaceAllString(x.raw, replacement)

	return nil
}
