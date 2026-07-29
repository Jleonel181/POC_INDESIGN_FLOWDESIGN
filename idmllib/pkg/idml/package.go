package idml

import (
	"archive/zip"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
	"github.com/dimelords/idmllib/v2/pkg/xmp"
)

// fileEntry almacena el contenido y los metadatos ZIP de un archivo.
type fileEntry struct {
	data   []byte
	header *zip.FileHeader
}

// Package representa un documento IDML cargado en memoria.
// En la Fase 1, los archivos se almacenan como bytes crudos con sus metadatos ZIP preservados.
// En la Fase 2, se agrega acceso estructurado al Document de designmap.xml.
//
// DECISIÓN DE DISEÑO: Parseo lazy con caché
// Package usa una estrategia de parseo lazy donde los archivos solo se parsean al primer acceso,
// y luego se cachean para llamadas posteriores. Esto provee varios beneficios:
// 1. Tiempos de carga iniciales rápidos - solo lee la estructura ZIP, no parsea todo el XML
// 2. Eficiencia de memoria - solo parsea los archivos que realmente se usan
// 3. Rendimiento - evita re-parsear el mismo archivo múltiples veces
// 4. Flexibilidad - soporta tanto acceso a bytes crudos como acceso a objetos estructurados
// El trade-off es código levemente más complejo, pero los beneficios de rendimiento son significativos
// para archivos IDML grandes donde típicamente solo se accede a un subconjunto del contenido.
type Package struct {
	// files mapea nombres de archivo a su contenido y metadatos.
	// Intencionalmente no se exporta para mantener la encapsulación.
	files map[string]*fileEntry

	// fileOrder preserva el orden original de los archivos en el ZIP.
	// DECISIÓN DE DISEÑO: Preservar el orden de archivos ZIP para roundtrip byte-perfecto
	// InDesign es sensible al orden de archivos, particularmente el archivo mimetype que
	// debe ir primero y sin comprimir. Mantener el orden original garantiza compatibilidad.
	fileOrder []string

	// DECISIÓN DE DISEÑO: Estrategia de caché de dos niveles
	// Se cachean tanto la versión genérica (ResourceFile) como la tipada (FontsFile, StylesFile)
	// de los archivos de recursos. Esto permite acceder al mismo archivo a través de distintas
	// APIs sin re-parsear, manteniendo la seguridad de tipos donde se necesita.

	// document es el Document parseado de designmap.xml (Fase 2).
	// Se parsea bajo demanda y se cachea.
	document *document.Document

	// documentMetadata almacena instrucciones de procesamiento y otros metadatos
	// que deben preservarse durante el marshal/unmarshal.
	documentMetadata *document.DocumentWithMetadata

	// stories cachea los archivos Story parseados del directorio Stories/.
	// La clave del mapa es el nombre del archivo (ej., "Stories/Story_u1d8.xml").
	stories map[string]*story.Story

	// spreads cachea los archivos Spread parseados del directorio Spreads/.
	// La clave del mapa es el nombre del archivo (ej., "Spreads/Spread_u210.xml").
	spreads map[string]*spread.Spread

	// resources cachea los archivos Resource parseados del directorio Resources/.
	// La clave del mapa es el nombre del archivo (ej., "Resources/Graphic.xml").
	// Este es el parser genérico basado en preservación.
	resources map[string]*ResourceFile

	// fonts cachea el archivo Fonts.xml tipado (si fue parseado)
	fonts *resources.FontsFile

	// graphics cachea el archivo Graphic.xml tipado (si fue parseado)
	graphics *resources.GraphicFile

	// styles cachea el archivo Styles.xml tipado (si fue parseado)
	styles *resources.StylesFile

	// metadata cachea archivos de metadatos opcionales (META-INF/*, XML/*).
	// La clave del mapa es la ruta del archivo (ej., "META-INF/container.xml").
	metadata map[string]*MetadataFile

	// XMPMetadata contiene el paquete XMP extraído de designmap.xml.
	// XMP es el estándar de Adobe para embeber metadatos en documentos.
	// Ejemplo: <?xpacket begin="" id="..."?>...<x:xmpmeta>...</x:xmpmeta><?xpacket end="r"?>
	XMPMetadata string

	// indexState almacena el índice de elementos para búsquedas O(1) de page items.
	// Se construye de forma lazy en la primera llamada a SelectXxxByID.
	indexState itemIndexState
}

// New crea un nuevo paquete IDML vacío.
func New() *Package {
	return &Package{
		files:     make(map[string]*fileEntry),
		stories:   make(map[string]*story.Story),
		spreads:   make(map[string]*spread.Spread),
		resources: make(map[string]*ResourceFile),
		metadata:  make(map[string]*MetadataFile),
	}
}

// Files retorna una copia de todos los nombres de archivo del paquete.
// Útil para inspección y depuración.
func (p *Package) Files() []string {
	names := make([]string, 0, len(p.files))
	for name := range p.files {
		names = append(names, name)
	}
	return names
}

// FileCount retorna la cantidad de archivos en el paquete.
func (p *Package) FileCount() int {
	return len(p.files)
}

// Document retorna el Document parseado de designmap.xml.
// El documento se parsea en el primer acceso y se cachea.
// Retorna un error si designmap.xml no existe o no puede parsearse.
func (p *Package) Document() (*document.Document, error) {
	// Retornar el documento cacheado si está disponible
	if doc, cached := p.getCachedDocument(); cached {
		return doc, nil
	}

	// Obtener el archivo designmap.xml
	entry, err := p.getFileEntry(PathDesignmap)
	if err != nil {
		return nil, err
	}

	// Parsear el documento con metadatos (instrucciones de procesamiento, etc.)
	docMeta, err := document.ParseDocumentWithMetadata(entry.data)
	if err != nil {
		return nil, err
	}

	// Cachear para llamadas futuras
	p.cacheDocument(docMeta.Document, docMeta)
	return p.document, nil
}

// Story retorna un Story parseado del directorio Stories/.
// El story se parsea en el primer acceso y se cachea.
// Retorna un error si el archivo no existe o no puede parsearse.
func (p *Package) Story(filename string) (*story.Story, error) {
	// Retornar el story cacheado si está disponible
	if st, cached := p.getCachedStory(filename); cached {
		return st, nil
	}

	// Obtener el archivo de story
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}

	// Parsear el story
	st, err := story.ParseStory(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse story", filename, err)
	}

	// Cachear para llamadas futuras
	p.cacheStory(filename, st)
	return st, nil
}

// Stories retorna todos los archivos Story parseados del directorio Stories/.
// Los stories se parsean en el primer acceso y se cachean.
func (p *Package) Stories() (map[string]*story.Story, error) {
	stories := make(map[string]*story.Story)

	// Primero, agregar los stories ya cacheados
	for filename, st := range p.stories {
		stories[filename] = st
	}

	// Luego buscar todos los archivos de story en p.files que aún no están cacheados
	for filename := range p.files {
		if IsStoryPath(filename) {
			// Saltar si ya está en caché
			if _, cached := stories[filename]; cached {
				continue
			}

			st, err := p.Story(filename)
			if err != nil {
				return nil, err
			}
			stories[filename] = st
		}
	}

	return stories, nil
}

// Spread retorna un Spread parseado del directorio Spreads/.
// El spread se parsea en el primer acceso y se cachea.
// Retorna un error si el archivo no existe o no puede parsearse.
func (p *Package) Spread(filename string) (*spread.Spread, error) {
	// Retornar el spread cacheado si está disponible
	if sp, cached := p.getCachedSpread(filename); cached {
		return sp, nil
	}

	// Obtener el archivo de spread
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}

	// Parsear el spread
	sp, err := spread.ParseSpread(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse spread", filename, err)
	}

	// Cachear para llamadas futuras
	p.cacheSpread(filename, sp)
	return sp, nil
}

// Spreads retorna todos los archivos Spread parseados del directorio Spreads/.
// Los spreads se parsean en el primer acceso y se cachean.
func (p *Package) Spreads() (map[string]*spread.Spread, error) {
	spreads := make(map[string]*spread.Spread)

	// Buscar todos los archivos de spread
	for filename := range p.files {
		if IsSpreadPath(filename) {
			sp, err := p.Spread(filename)
			if err != nil {
				return nil, err
			}
			spreads[filename] = sp
		}
	}

	return spreads, nil
}

// Resource retorna un archivo Resource parseado del directorio Resources/.
// El recurso se parsea en el primer acceso y se cachea.
// Retorna un error si el archivo no existe o no puede parsearse.
func (p *Package) Resource(filename string) (*ResourceFile, error) {
	// Retornar el recurso cacheado si está disponible
	if resource, cached := p.getCachedResource(filename); cached {
		return resource, nil
	}

	// Obtener el archivo de recurso
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}

	// Parsear el recurso
	resource, err := ParseResourceFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse resource", filename, err)
	}

	// Cachear para llamadas futuras
	p.cacheResource(filename, resource)
	return resource, nil
}

// Resources retorna todos los archivos Resource parseados del directorio Resources/.
// Los recursos se parsean en el primer acceso y se cachean.
func (p *Package) Resources() (map[string]*ResourceFile, error) {
	resources := make(map[string]*ResourceFile)

	// Buscar todos los archivos de recurso
	for filename := range p.files {
		if IsResourcePath(filename) {
			resource, err := p.Resource(filename)
			if err != nil {
				return nil, err
			}
			resources[filename] = resource
		}
	}

	return resources, nil
}

// MetadataFile retorna un archivo de metadatos por ruta.
// Los archivos de metadatos son opcionales e incluyen:
//   - META-INF/container.xml
//   - META-INF/metadata.xml
//   - XML/Tags.xml
//   - XML/BackingStory.xml
//
// El archivo se parsea en el primer acceso y se cachea.
// Retorna ErrNotFound si el archivo no existe.
func (p *Package) MetadataFile(path string) (*MetadataFile, error) {
	// Retornar el cacheado si está disponible
	if mf, cached := p.getCachedMetadata(path); cached {
		return mf, nil
	}

	// Obtener el archivo
	entry, err := p.getFileEntry(path)
	if err != nil {
		return nil, err
	}

	// Parsear el archivo de metadatos
	mf, err := ParseMetadataFile(path, entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse metadata", path, err)
	}

	// Cachear para llamadas futuras
	p.cacheMetadata(path, mf)
	return mf, nil
}

// MetadataFiles retorna todos los archivos de metadatos del paquete.
// Incluye todos los archivos en los directorios META-INF/ y XML/.
func (p *Package) MetadataFiles() (map[string]*MetadataFile, error) {
	metadata := make(map[string]*MetadataFile)

	// Buscar todos los archivos de metadatos (META-INF/* y XML/*)
	for filename := range p.files {
		if IsMetaInfPath(filename) || IsXMLPath(filename) {
			mf, err := p.MetadataFile(filename)
			if err != nil {
				return nil, err
			}
			metadata[filename] = mf
		}
	}

	return metadata, nil
}

// Fonts retorna el archivo Fonts.xml tipado.
// El archivo se parsea en el primer acceso y se cachea.
// Retorna un error si el archivo no existe o no puede parsearse.
func (p *Package) Fonts() (*resources.FontsFile, error) {
	// Retornar el cacheado si está disponible
	if fonts, cached := p.getCachedFonts(); cached {
		return fonts, nil
	}

	// Obtener el archivo Fonts.xml
	entry, err := p.getFileEntry(PathFonts)
	if err != nil {
		return nil, err
	}

	// Parsear el archivo de fuentes
	fonts, err := resources.ParseFontsFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse fonts", PathFonts, err)
	}

	// Cachear para llamadas futuras
	p.cacheFonts(fonts)
	return fonts, nil
}

// Graphics retorna el archivo Graphic.xml tipado.
// El archivo se parsea en el primer acceso y se cachea.
// Retorna un error si el archivo no existe o no puede parsearse.
func (p *Package) Graphics() (*resources.GraphicFile, error) {
	// Retornar el cacheado si está disponible
	if graphics, cached := p.getCachedGraphics(); cached {
		return graphics, nil
	}

	// Obtener el archivo Graphic.xml
	entry, err := p.getFileEntry(PathGraphic)
	if err != nil {
		return nil, err
	}

	// Parsear el archivo de gráficos
	graphics, err := resources.ParseGraphicFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse graphics", PathGraphic, err)
	}

	// Cachear para llamadas futuras
	p.cacheGraphics(graphics)
	return graphics, nil
}

// Styles retorna el archivo Styles.xml tipado.
// El archivo se parsea en el primer acceso y se cachea.
// Retorna un error si el archivo no existe o no puede parsearse.
func (p *Package) Styles() (*resources.StylesFile, error) {
	// Retornar el cacheado si está disponible
	if styles, cached := p.getCachedStyles(); cached {
		return styles, nil
	}

	// Obtener el archivo Styles.xml
	entry, err := p.getFileEntry(PathStyles)
	if err != nil {
		return nil, err
	}

	// Parsear el archivo de estilos
	styles, err := resources.ParseStylesFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse styles", PathStyles, err)
	}

	// Cachear para llamadas futuras
	p.cacheStyles(styles)
	return styles, nil
}

// SetFonts actualiza el archivo de fuentes cacheado.
// El archivo será serializado cuando se llame a Write().
func (p *Package) SetFonts(fonts *resources.FontsFile) {
	p.cacheFonts(fonts)
}

// SetStyles actualiza el archivo de estilos cacheado.
// El archivo será serializado cuando se llame a Write().
func (p *Package) SetStyles(styles *resources.StylesFile) {
	p.cacheStyles(styles)
}

// SetGraphics actualiza el archivo de gráficos cacheado.
// El archivo será serializado cuando se llame a Write().
func (p *Package) SetGraphics(graphics *resources.GraphicFile) {
	p.cacheGraphics(graphics)
}

// XMP retorna un accessor XMP para los metadatos del paquete.
// Permite leer y modificar metadatos XMP de forma segura con tipos.
// Retorna una instancia de xmp.Metadata que puede usarse para actualizar timestamps,
// eliminar thumbnails o modificar campos específicos.
func (p *Package) XMP() *xmp.Metadata {
	return xmp.Parse(p.XMPMetadata)
}

// SetXMP actualiza los metadatos XMP del paquete.
// Debe llamarse después de modificar los metadatos XMP para persistir los cambios.
func (p *Package) SetXMP(x *xmp.Metadata) {
	p.XMPMetadata = x.String()
}

// ============================================================================
