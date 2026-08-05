package idmlgen

import (
	"fmt"

	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
)

// templateRefs contiene los IDs de referencia leídos del paquete generado por NewFromTemplate.
// Evita hardcodear IDs de la plantilla minimal; si la plantilla cambia, estos valores se
// adaptan automáticamente.
type templateRefs struct {
	layerID      string // ID de la primera capa (ej: "uba")
	masterSpread string // Ruta del primer MasterSpread (ej: "MasterSpreads/MasterSpread_ub4.xml")
	sectionID    string // ID de la primera Section (usada como AppliedAlternateLayout)
	spreadPath   string // Ruta del spread de la plantilla
	textFrameID  string // Self del TextFrame placeholder a eliminar
	storyPath    string // Path de la Story placeholder a eliminar
	storyID      string // ID de la Story placeholder (para limpiar StoryList)
}

// readTemplateRefs extrae dinámicamente los IDs de referencia del paquete recién generado.
func readTemplateRefs(pkg *idmlpkg.Package) (*templateRefs, error) {
	doc, err := pkg.Document()
	if err != nil {
		return nil, fmt.Errorf("error al leer designmap: %w", err)
	}

	refs := &templateRefs{}

	// Layer: tomamos el primer layer disponible
	if len(doc.Layers) > 0 {
		refs.layerID = doc.Layers[0].Self
	}

	// MasterSpread: ruta del primer master spread
	if len(doc.MasterSpreads) > 0 {
		refs.masterSpread = doc.MasterSpreads[0].Src
	}

	// Section: ID de la primera section (se usa como AppliedAlternateLayout)
	if len(doc.Sections) > 0 {
		refs.sectionID = doc.Sections[0].Self
	}

	// Spread de la plantilla: el primero
	if len(doc.Spreads) > 0 {
		refs.spreadPath = doc.Spreads[0].Src
	}

	// TextFrame y Story placeholder: se leen del spread
	if refs.spreadPath != "" {
		sp, err := pkg.Spread(refs.spreadPath)
		if err == nil && len(sp.TextFrames()) > 0 {
			tf := sp.TextFrames()[0]
			refs.textFrameID = tf.Self
			refs.storyID = tf.ParentStory
			refs.storyPath = "Stories/Story_" + tf.ParentStory + ".xml"
		}
	}

	return refs, nil
}

// masterSpreadSelf extrae el Self del MasterSpread a partir de su ruta.
// "MasterSpreads/MasterSpread_ub4.xml" → "ub4"
func masterSpreadSelf(path string) string {
	const prefix = "MasterSpreads/MasterSpread_"
	const suffix = ".xml"
	if len(path) > len(prefix)+len(suffix) {
		return path[len(prefix) : len(path)-len(suffix)]
	}
	return path
}
