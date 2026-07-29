package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Métodos de gestión de caché para el struct Package.
// Estos métodos proveen operaciones centralizadas de caché e invalidación.

// clearCache limpia todos los objetos parseados en caché.
// Esto fuerza el re-parseo desde los datos crudos del archivo en el próximo acceso.
// Útil cuando los datos del archivo han sido modificados externamente.
func (p *Package) clearCache() {
	// Limpiar caché del documento
	p.document = nil
	p.documentMetadata = nil

	// Limpiar caché de stories
	p.stories = make(map[string]*story.Story)

	// Limpiar caché de spreads
	p.spreads = make(map[string]*spread.Spread)

	// Limpiar caché de resources
	p.resources = make(map[string]*ResourceFile)

	// Limpiar caché de resources tipados
	p.fonts = nil
	p.graphics = nil
	p.styles = nil

	// Limpiar caché de metadata
	p.metadata = make(map[string]*MetadataFile)

	// Limpiar caché del índice
	p.indexState = itemIndexState{}
}

// invalidateCache invalida los objetos en caché para una ruta de archivo específica.
// Es más eficiente que limpiar todos los cachés cuando solo cambia un archivo.
func (p *Package) invalidateCache(path string) {
	switch path {
	case PathDesignmap:
		p.document = nil
		p.documentMetadata = nil

	case PathFonts:
		p.fonts = nil
		// También limpiar el caché genérico de resources para este archivo
		delete(p.resources, path)

	case PathGraphic:
		p.graphics = nil
		// También limpiar el caché genérico de resources para este archivo
		delete(p.resources, path)

	case PathStyles:
		p.styles = nil
		// También limpiar el caché genérico de resources para este archivo
		delete(p.resources, path)

	default:
		// Manejar archivos de stories
		if IsStoryPath(path) {
			delete(p.stories, path)
			// Invalidar el índice ya que las stories cambiaron
			p.invalidateIndex()
			return
		}

		// Manejar archivos de spreads
		if IsSpreadPath(path) {
			delete(p.spreads, path)
			// Invalidar el índice ya que los spreads cambiaron
			p.invalidateIndex()
			return
		}

		// Manejar archivos de resources
		if IsResourcePath(path) {
			delete(p.resources, path)
			return
		}

		// Manejar archivos de metadata
		if IsMetaInfPath(path) || IsXMLPath(path) {
			delete(p.metadata, path)
			return
		}
	}
}

// invalidateStoryCache limpia todos los objetos de story en caché.
// Útil cuando múltiples stories han sido modificadas.
func (p *Package) invalidateStoryCache() {
	p.stories = make(map[string]*story.Story)
	p.invalidateIndex() // Las stories afectan el índice
}

// invalidateSpreadCache limpia todos los objetos de spread en caché.
// Útil cuando múltiples spreads han sido modificados.
func (p *Package) invalidateSpreadCache() {
	p.spreads = make(map[string]*spread.Spread)
	p.invalidateIndex() // Los spreads afectan el índice
}

// invalidateResourceCache limpia todos los objetos de resource en caché.
// Incluye tanto resources genéricos como resources tipados.
func (p *Package) invalidateResourceCache() {
	p.resources = make(map[string]*ResourceFile)
	p.fonts = nil
	p.graphics = nil
	p.styles = nil
}

// invalidateMetadataCache limpia todos los objetos de metadata en caché.
func (p *Package) invalidateMetadataCache() {
	p.metadata = make(map[string]*MetadataFile)
}

// invalidateIndex limpia el índice de elementos de página.
// El índice se reconstruirá en el próximo acceso a los métodos de selección.
func (p *Package) invalidateIndex() {
	p.indexState = itemIndexState{}
}

// getCacheStats retorna estadísticas sobre los objetos en caché.
// Útil para debugging y monitoreo del uso de caché.
func (p *Package) getCacheStats() CacheStats {
	stats := CacheStats{}

	// Caché del documento
	if p.document != nil {
		stats.DocumentCached = true
	}

	// Caché de stories
	stats.StoriesCached = len(p.stories)

	// Caché de spreads
	stats.SpreadsCached = len(p.spreads)

	// Caché de resources
	stats.ResourcesCached = len(p.resources)

	// Caché de resources tipados
	if p.fonts != nil {
		stats.FontsCached = true
	}
	if p.graphics != nil {
		stats.GraphicsCached = true
	}
	if p.styles != nil {
		stats.StylesCached = true
	}

	// Caché de metadata
	stats.MetadataCached = len(p.metadata)

	// Caché del índice
	if p.indexState.index != nil {
		stats.IndexCached = true
		stats.IndexedItems = p.ItemCount()
	}

	return stats
}

// CacheStats provee información sobre los objetos en caché de un Package.
type CacheStats struct {
	DocumentCached  bool // Si el documento está en caché
	StoriesCached   int  // Número de stories en caché
	SpreadsCached   int  // Número de spreads en caché
	ResourcesCached int  // Número de resources genéricos en caché
	FontsCached     bool // Si las fuentes tipadas están en caché
	GraphicsCached  bool // Si los gráficos tipados están en caché
	StylesCached    bool // Si los estilos tipados están en caché
	MetadataCached  int  // Número de archivos de metadata en caché
	IndexCached     bool // Si el índice de elementos de página está en caché
	IndexedItems    int  // Número de elementos en el índice
}

// ensureCacheInitialized asegura que todos los mapas de caché estén inicializados.
// Es llamado por los métodos que necesitan escribir en los mapas de caché.
func (p *Package) ensureCacheInitialized() {
	if p.stories == nil {
		p.stories = make(map[string]*story.Story)
	}
	if p.spreads == nil {
		p.spreads = make(map[string]*spread.Spread)
	}
	if p.resources == nil {
		p.resources = make(map[string]*ResourceFile)
	}
	if p.metadata == nil {
		p.metadata = make(map[string]*MetadataFile)
	}
}

// cacheStory almacena una story parseada en la caché.
func (p *Package) cacheStory(filename string, st *story.Story) {
	p.ensureCacheInitialized()
	p.stories[filename] = st
}

// cacheSpread almacena un spread parseado en la caché.
func (p *Package) cacheSpread(filename string, sp *spread.Spread) {
	p.ensureCacheInitialized()
	p.spreads[filename] = sp
}

// cacheResource almacena un resource parseado en la caché.
func (p *Package) cacheResource(filename string, resource *ResourceFile) {
	p.ensureCacheInitialized()
	p.resources[filename] = resource
}

// cacheMetadata almacena un archivo de metadata parseado en la caché.
func (p *Package) cacheMetadata(path string, metadata *MetadataFile) {
	p.ensureCacheInitialized()
	p.metadata[path] = metadata
}

// cacheDocument almacena un documento parseado en la caché.
func (p *Package) cacheDocument(doc *document.Document, docMeta *document.DocumentWithMetadata) {
	p.document = doc
	p.documentMetadata = docMeta
}

// cacheFonts almacena las fuentes parseadas en la caché.
func (p *Package) cacheFonts(fonts *resources.FontsFile) {
	p.fonts = fonts
}

// cacheGraphics almacena los gráficos parseados en la caché.
func (p *Package) cacheGraphics(graphics *resources.GraphicFile) {
	p.graphics = graphics
}

// cacheStyles almacena los estilos parseados en la caché.
func (p *Package) cacheStyles(styles *resources.StylesFile) {
	p.styles = styles
}

// getCachedStory recupera una story en caché si existe.
func (p *Package) getCachedStory(filename string) (*story.Story, bool) {
	if p.stories == nil {
		return nil, false
	}
	st, exists := p.stories[filename]
	return st, exists
}

// getCachedSpread recupera un spread en caché si existe.
func (p *Package) getCachedSpread(filename string) (*spread.Spread, bool) {
	if p.spreads == nil {
		return nil, false
	}
	sp, exists := p.spreads[filename]
	return sp, exists
}

// getCachedResource recupera un resource en caché si existe.
func (p *Package) getCachedResource(filename string) (*ResourceFile, bool) {
	if p.resources == nil {
		return nil, false
	}
	resource, exists := p.resources[filename]
	return resource, exists
}

// getCachedMetadata recupera un archivo de metadata en caché si existe.
func (p *Package) getCachedMetadata(path string) (*MetadataFile, bool) {
	if p.metadata == nil {
		return nil, false
	}
	metadata, exists := p.metadata[path]
	return metadata, exists
}

// getCachedDocument recupera el documento en caché si existe.
func (p *Package) getCachedDocument() (*document.Document, bool) {
	return p.document, p.document != nil
}

// getCachedFonts recupera las fuentes en caché si existen.
func (p *Package) getCachedFonts() (*resources.FontsFile, bool) {
	return p.fonts, p.fonts != nil
}

// getCachedGraphics recupera los gráficos en caché si existen.
func (p *Package) getCachedGraphics() (*resources.GraphicFile, bool) {
	return p.graphics, p.graphics != nil
}

// getCachedStyles recupera los estilos en caché si existen.
func (p *Package) getCachedStyles() (*resources.StylesFile, bool) {
	return p.styles, p.styles != nil
}
