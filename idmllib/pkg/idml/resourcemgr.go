// resourcemgr.go define el tipo ResourceManager y los tipos relacionados para
// gestionar los recursos de un documento IDML (Epic 2: API de Gestión de Recursos).
package idml

import "fmt"

// ResourceManager analiza y gestiona los recursos de un paquete IDML.
// Provee funcionalidad para:
//   - Detectar recursos huérfanos (fuentes, estilos, colores)
//   - Limpiar recursos no utilizados
//   - Validar referencias a recursos
//   - Resolución automática de recursos faltantes
//
// El ResourceManager analiza el contenido del documento directamente para
// determinar qué recursos están siendo utilizados.
type ResourceManager struct {
	pkg *Package
}

// NewResourceManager crea un nuevo ResourceManager para el paquete dado.
// El manager puede reutilizarse para múltiples operaciones sobre el mismo paquete.
func NewResourceManager(pkg *Package) *ResourceManager {
	return &ResourceManager{
		pkg: pkg,
	}
}

// dependencySet registra todas las dependencias encontradas durante el análisis.
// Es un tipo interno usado por el ResourceManager.
type dependencySet struct {
	fonts           map[string]bool
	paragraphStyles map[string]bool
	characterStyles map[string]bool
	objectStyles    map[string]bool
	colors          map[string]bool
	swatches        map[string]bool
	layers          map[string]bool
}

// newDependencySet crea un nuevo conjunto de dependencias vacío.
func newDependencySet() *dependencySet {
	return &dependencySet{
		fonts:           make(map[string]bool),
		paragraphStyles: make(map[string]bool),
		characterStyles: make(map[string]bool),
		objectStyles:    make(map[string]bool),
		colors:          make(map[string]bool),
		swatches:        make(map[string]bool),
		layers:          make(map[string]bool),
	}
}

// CleanupOptions configura qué limpiar al eliminar elementos.
// Por defecto, la mayoría de las operaciones de limpieza están habilitadas por seguridad.
// Usar DefaultCleanupOptions() para obtener una configuración predeterminada segura.
type CleanupOptions struct {
	// RemoveOrphanedFonts elimina fuentes que ya no se usan en ninguna story
	RemoveOrphanedFonts bool

	// RemoveOrphanedParagraphStyles elimina estilos de párrafo no usados en ninguna story
	RemoveOrphanedParagraphStyles bool

	// RemoveOrphanedCharacterStyles elimina estilos de carácter no usados en ninguna story
	RemoveOrphanedCharacterStyles bool

	// RemoveOrphanedObjectStyles elimina estilos de objeto no usados en ningún page item
	// PREDETERMINADO: false - los estilos de objeto frecuentemente sirven como plantillas
	RemoveOrphanedObjectStyles bool

	// RemoveOrphanedColors elimina colores no usados en ningún elemento
	// PREDETERMINADO: false - los colores son parte de la biblioteca de colores
	RemoveOrphanedColors bool

	// RemoveOrphanedSwatches elimina swatches no usados en ningún elemento
	// PREDETERMINADO: false - los swatches son parte de la biblioteca de swatches
	RemoveOrphanedSwatches bool

	// RemoveOrphanedLayers elimina capas vacías sin page items
	// PREDETERMINADO: false - las capas son estructurales y los usuarios suelen querer conservarlas
	RemoveOrphanedLayers bool

	// DryRun no elimina nada, solo reporta qué se eliminaría
	// Útil para previsualizar operaciones de limpieza antes de confirmarlas
	DryRun bool
}

// DefaultCleanupOptions retorna opciones de limpieza predeterminadas y seguras.
// Estos valores son conservadores y solo eliminan recursos obviamente no utilizados.
func DefaultCleanupOptions() CleanupOptions {
	return CleanupOptions{
		RemoveOrphanedFonts:           true,
		RemoveOrphanedParagraphStyles: true,
		RemoveOrphanedCharacterStyles: true,
		RemoveOrphanedObjectStyles:    false, // Conservar - frecuentemente usados como plantillas
		RemoveOrphanedColors:          false, // Conservar - parte de la biblioteca de colores
		RemoveOrphanedSwatches:        false, // Conservar - parte de la biblioteca de swatches
		RemoveOrphanedLayers:          false, // Conservar - las capas son estructurales
		DryRun:                        false,
	}
}

// ValidationOptions configura la validación al agregar elementos.
// Estas opciones controlan qué se valida y cómo se manejan los recursos faltantes.
type ValidationOptions struct {
	// EnsureStylesExist verifica que existan los estilos de párrafo y de carácter
	EnsureStylesExist bool

	// EnsureFontsExist verifica que existan las fuentes referenciadas en los estilos
	EnsureFontsExist bool

	// EnsureColorsExist verifica que existan los colores referenciados en los elementos
	EnsureColorsExist bool

	// EnsureLayersExist verifica que existan las capas referenciadas por los page items
	EnsureLayersExist bool

	// AutoAddMissing agrega automáticamente los recursos faltantes con valores predeterminados
	// Cuando es false, los recursos faltantes fallarán (si FailOnMissing=true) o serán ignorados
	AutoAddMissing bool

	// FailOnMissing retorna un error si faltan recursos (cuando AutoAdd=false)
	// Cuando es false, las advertencias de validación se registran pero la operación continúa
	FailOnMissing bool
}

// Presets comunes de ValidationOptions para conveniencia.
var (
	// NoValidation deshabilita todas las verificaciones de validación.
	// Usar cuando se tiene certeza de que los recursos son válidos o cuando el rendimiento es crítico.
	NoValidation = ValidationOptions{}

	// FullValidation habilita todas las verificaciones pero no agrega recursos faltantes automáticamente.
	// Usar cuando se quiere asegurar que todos los recursos existan pero fallar si no es así.
	FullValidation = ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  true,
		EnsureColorsExist: true,
		EnsureLayersExist: true,
		FailOnMissing:     true,
	}

	// AutoResolve habilita todas las verificaciones y agrega automáticamente los recursos faltantes.
	// Usar cuando se quiere asegurar que el documento sea válido y corregirlo automáticamente.
	AutoResolve = ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  true,
		EnsureColorsExist: true,
		EnsureLayersExist: true,
		AutoAddMissing:    true,
	}

	// StylesOnly valida únicamente estilos de párrafo y de carácter.
	// Usar cuando se modifica contenido de stories pero los colores y capas no importan.
	StylesOnly = ValidationOptions{
		EnsureStylesExist: true,
		FailOnMissing:     true,
	}
)

// DefaultValidationOptions retorna opciones de validación predeterminadas y seguras.
// Estos valores aseguran la integridad del documento validando todas las referencias a recursos.
func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  true,
		EnsureColorsExist: true,
		EnsureLayersExist: true,
		AutoAddMissing:    false, // Conservador - requiere opt-in explícito
		FailOnMissing:     true,  // Fallar rápido ante dependencias faltantes
	}
}

// OrphanedResources contiene recursos que están definidos pero no se usan.
// Es el resultado de FindOrphans() y representa recursos que potencialmente
// pueden eliminarse del documento.
type OrphanedResources struct {
	// Fonts contiene nombres de familias tipográficas definidas pero no usadas
	Fonts []string

	// ParagraphStyles contiene IDs de estilos de párrafo definidos pero no usados
	ParagraphStyles []string

	// CharacterStyles contiene IDs de estilos de carácter definidos pero no usados
	CharacterStyles []string

	// ObjectStyles contiene IDs de estilos de objeto definidos pero no usados
	ObjectStyles []string

	// Colors contiene IDs de colores definidos pero no usados
	Colors []string

	// Swatches contiene IDs de swatches definidos pero no usados
	Swatches []string

	// Layers contiene IDs de capas que existen pero no tienen page items
	Layers []string
}

// HasOrphans retorna true si hay recursos huérfanos.
func (or *OrphanedResources) HasOrphans() bool {
	return len(or.Fonts) > 0 ||
		len(or.ParagraphStyles) > 0 ||
		len(or.CharacterStyles) > 0 ||
		len(or.ObjectStyles) > 0 ||
		len(or.Colors) > 0 ||
		len(or.Swatches) > 0 ||
		len(or.Layers) > 0
}

// Count retorna el número total de recursos huérfanos en todos los tipos.
func (or *OrphanedResources) Count() int {
	return len(or.Fonts) +
		len(or.ParagraphStyles) +
		len(or.CharacterStyles) +
		len(or.ObjectStyles) +
		len(or.Colors) +
		len(or.Swatches) +
		len(or.Layers)
}

// CleanupResult contiene información sobre lo que fue limpiado.
// Es retornado por CleanupOrphans() para proveer retroalimentación detallada
// sobre la operación de limpieza.
type CleanupResult struct {
	// RemovedFonts lista los nombres de familias tipográficas eliminadas
	RemovedFonts []string

	// RemovedParagraphStyles lista los IDs de estilos de párrafo eliminados
	RemovedParagraphStyles []string

	// RemovedCharacterStyles lista los IDs de estilos de carácter eliminados
	RemovedCharacterStyles []string

	// RemovedObjectStyles lista los IDs de estilos de objeto eliminados
	RemovedObjectStyles []string

	// RemovedColors lista los IDs de colores eliminados
	RemovedColors []string

	// RemovedSwatches lista los IDs de swatches eliminados
	RemovedSwatches []string

	// RemovedLayers lista los IDs de capas eliminadas
	RemovedLayers []string
}

// Count retorna el número total de recursos eliminados en todos los tipos.
func (cr *CleanupResult) Count() int {
	return len(cr.RemovedFonts) +
		len(cr.RemovedParagraphStyles) +
		len(cr.RemovedCharacterStyles) +
		len(cr.RemovedObjectStyles) +
		len(cr.RemovedColors) +
		len(cr.RemovedSwatches) +
		len(cr.RemovedLayers)
}

// MissingResources contiene recursos que son referenciados pero no están definidos.
// Cada clave del mapa es un ID de recurso, y el valor es una lista de IDs de elementos
// o nombres de archivo que lo referencian.
type MissingResources struct {
	// Fonts mapea nombres de familias tipográficas a los elementos que las usan
	Fonts map[string][]string

	// ParagraphStyles mapea IDs de estilos a los nombres de archivo de stories que los usan
	ParagraphStyles map[string][]string

	// CharacterStyles mapea IDs de estilos a los nombres de archivo de stories que los usan
	CharacterStyles map[string][]string

	// ObjectStyles mapea IDs de estilos a los IDs de elementos que los usan
	ObjectStyles map[string][]string

	// Colors mapea IDs de colores a los IDs de elementos que los usan
	Colors map[string][]string

	// Swatches mapea IDs de swatches a los IDs de elementos que los usan
	Swatches map[string][]string

	// Layers mapea IDs de capas a los IDs de elementos que están en ellas
	Layers map[string][]string
}

// HasMissing retorna true si hay recursos faltantes.
func (mr *MissingResources) HasMissing() bool {
	return len(mr.Fonts) > 0 ||
		len(mr.ParagraphStyles) > 0 ||
		len(mr.CharacterStyles) > 0 ||
		len(mr.ObjectStyles) > 0 ||
		len(mr.Colors) > 0 ||
		len(mr.Swatches) > 0 ||
		len(mr.Layers) > 0
}

// ValidationError representa un error de validación individual para un recurso faltante.
type ValidationError struct {
	// ResourceType describe qué tipo de recurso falta (ej. "Font", "ParagraphStyle")
	ResourceType string

	// ResourceID es el identificador del recurso faltante
	ResourceID string

	// UsedBy lista los IDs de elementos o nombres de archivo que referencian este recurso
	UsedBy []string

	// Message es un mensaje de error legible por humanos
	Message string
}

// Error implementa la interfaz error para ValidationError.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (used by %d elements)", ve.ResourceType, ve.ResourceID, len(ve.UsedBy))
}

// FindOrphans identifica todos los recursos huérfanos en el paquete.
// Un recurso huérfano es aquel que está definido en el paquete pero no
// es utilizado por ningún elemento.
//
// Este método escanea el documento completo para construir un panorama
// completo de qué recursos están definidos y cuáles están en uso. La diferencia
// entre estos dos conjuntos representa los recursos huérfanos.
//
// Retorna un struct OrphanedResources con todos los recursos huérfanos,
// o un error si el análisis falla.
