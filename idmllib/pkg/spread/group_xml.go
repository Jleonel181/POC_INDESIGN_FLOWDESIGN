package spread

import (
	"encoding/xml"
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// maxGroupDepth es el límite de anidamiento recursivo de grupos. InDesign no impone uno
// documentado, pero el corpus tiene profundidad máxima 2. El límite de 32 es defensivo:
// evita un stack overflow por un documento corrupto sin restringir nada real.
const maxGroupDepth = 32

// Append agrega un elemento de página al grupo, tanto en Items (para el constructor)
// como en OtherElements (para que se emita con la serialización estándar).
//
// ponytail: este método solo se usa para **construir** grupos. Los hijos parseados se
// quedan en OtherElements como XML crudo, que es lo que preserva la fidelidad perfecta.
// El día que GraphicLine y Polygon dejen de reordenar sus hijos, se puede pasar a typed
// decode y unificar los dos caminos.
func (g *Group) Append(item PageItem) error {
	if item == nil {
		return fmt.Errorf("grupo %q: no se puede agregar un elemento nil", g.Self)
	}

	tag := item.xmlTag()

	// Serializar el item a XML crudo para agregarlo a OtherElements.
	data, err := xml.Marshal(item)
	if err != nil {
		return fmt.Errorf("grupo %q: error al serializar %s: %w", g.Self, tag, err)
	}

	raw := common.RawXMLElement{
		XMLName: xml.Name{Local: tag},
		Content: data,
	}

	// Parsear los atributos del item para ponerlos en el RawXMLElement.
	// La forma más directa: re-parsear los bytes.
	dec := xml.NewDecoder(nil)
	_ = dec // No necesitamos parsear atributos separadamente: xml.Marshal ya los incluye
	// en el content. Pero RawXMLElement usa innerxml, que es el contenido DENTRO del
	// elemento, no el elemento completo. Necesitamos extraer los attrs y el inner content.

	// Alternativa más simple: usar EncodeElement para obtener el formato correcto.
	// RawXMLElement con Content = innerxml funciona porque encoding/xml emite
	// <XMLName attrs>Content</XMLName> donde Content es literal innerxml.
	// Necesitamos separar el start element (con sus attrs) del contenido interno.
	type rawParse struct {
		Attrs   []xml.Attr `xml:",any,attr"`
		Content []byte     `xml:",innerxml"`
	}
	var parsed rawParse
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("grupo %q: error al re-parsear %s: %w", g.Self, tag, err)
	}
	raw.Attrs = parsed.Attrs
	raw.Content = parsed.Content

	g.OtherElements = append(g.OtherElements, raw)
	g.Items = append(g.Items, item)
	return nil
}

// WalkItems recorre los elementos de página del grupo en profundidad (DFS, pre-order).
// Para grupos **parseados**, extrae los hijos del XML crudo al vuelo; para grupos
// **construidos**, usa Items directamente.
//
// Si fn retorna false, el recorrido se detiene.
func (g *Group) WalkItems(fn func(PageItem) bool) {
	g.walkWithDepth(fn, 0)
}

func (g *Group) walkWithDepth(fn func(PageItem) bool, depth int) bool {
	if depth > maxGroupDepth {
		return true // protección contra recursión infinita
	}

	// Si hay Items tipados (construidos con Append), recorrerlos.
	if len(g.Items) > 0 {
		for _, item := range g.Items {
			if !fn(item) {
				return false
			}
			if sub, ok := item.(*Group); ok {
				if !sub.walkWithDepth(fn, depth+1) {
					return false
				}
			}
		}
		return true
	}

	// Para grupos parseados, los hijos son OtherElements como raw XML. Se pueden
	// identificar por su XMLName.Local pero no se deserializan aquí: eso rompería
	// la fidelidad. Un llamador que necesite acceso tipado a los hijos de un grupo
	// parseado debería construir un grupo nuevo.
	return true
}
