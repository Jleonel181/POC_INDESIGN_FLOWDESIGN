package idml

// Rutas y prefijos de archivos del paquete IDML.
// Estas constantes definen las rutas estándar usadas dentro de los paquetes IDML.

const (
	// Archivos raíz
	PathMimetype  = "mimetype"
	PathDesignmap = "designmap.xml"

	// Archivos de recursos
	PathFonts       = "Resources/Fonts.xml"
	PathStyles      = "Resources/Styles.xml"
	PathGraphic     = "Resources/Graphic.xml"
	PathPreferences = "Resources/Preferences.xml"

	// Archivos de metadatos
	PathContainer    = "META-INF/container.xml"
	PathTags         = "XML/Tags.xml"
	PathBackingStory = "XML/BackingStory.xml"

	// Plantilla de master spread
	PathMasterSpread = "MasterSpreads/MasterSpread_ub4.xml"

	// Prefijos de directorios (con barra al final)
	PrefixStories       = "Stories/"
	PrefixSpreads       = "Spreads/"
	PrefixMasterSpreads = "MasterSpreads/"
	PrefixResources     = "Resources/"
	PrefixMetaInf       = "META-INF/"
	PrefixXML           = "XML/"

	// Extensión de archivo
	ExtXML = ".xml"
)

// StoryPath retorna la ruta estándar para un archivo de story.
// Ejemplo: StoryPath("u1d8") retorna "Stories/Story_u1d8.xml"
func StoryPath(id string) string {
	return PrefixStories + "Story_" + id + ExtXML
}

// SpreadPath retorna la ruta estándar para un archivo de spread.
// Ejemplo: SpreadPath("u210") retorna "Spreads/Spread_u210.xml"
func SpreadPath(id string) string {
	return PrefixSpreads + "Spread_" + id + ExtXML
}

// MasterSpreadPath retorna la ruta estándar para un archivo de master spread.
// Ejemplo: MasterSpreadPath("ub4") retorna "MasterSpreads/MasterSpread_ub4.xml"
func MasterSpreadPath(id string) string {
	return PrefixMasterSpreads + "MasterSpread_" + id + ExtXML
}

// IsStoryPath verifica si una ruta pertenece al directorio Stories.
func IsStoryPath(path string) bool {
	return len(path) > len(PrefixStories) &&
		path[:len(PrefixStories)] == PrefixStories &&
		len(path) > 4 && path[len(path)-4:] == ExtXML
}

// IsSpreadPath verifica si una ruta pertenece al directorio Spreads.
func IsSpreadPath(path string) bool {
	return len(path) > len(PrefixSpreads) &&
		path[:len(PrefixSpreads)] == PrefixSpreads &&
		len(path) > 4 && path[len(path)-4:] == ExtXML
}

// IsResourcePath verifica si una ruta pertenece al directorio Resources.
func IsResourcePath(path string) bool {
	return len(path) > len(PrefixResources) &&
		path[:len(PrefixResources)] == PrefixResources &&
		len(path) > 4 && path[len(path)-4:] == ExtXML
}

// IsMetaInfPath verifica si una ruta pertenece al directorio META-INF.
func IsMetaInfPath(path string) bool {
	return len(path) > len(PrefixMetaInf) &&
		path[:len(PrefixMetaInf)] == PrefixMetaInf
}

// IsXMLPath verifica si una ruta pertenece al directorio XML.
func IsXMLPath(path string) bool {
	return len(path) > len(PrefixXML) &&
		path[:len(PrefixXML)] == PrefixXML
}
