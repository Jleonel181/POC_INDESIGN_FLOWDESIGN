// El paquete analysis proporciona herramientas para analizar documentos IDML y rastrear dependencias.
// Se utiliza en la exportación a IDMS para recopilar todos los recursos necesarios de los elementos de página seleccionados.
package analysis

import (
	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// DependencySet registra todas las dependencias de un conjunto de elementos de página seleccionados.
// Esto incluye Stories, estilos, colores, fuentes y otros recursos necesarios
// para exportar la selección como un fragmento IDMS independiente.
type DependencySet struct {
	// Stories registra los archivos Story referenciados mediante su nombre de archivo
	// Clave: nombre del archivo Story (p. ej., "Stories/Story_u1d8.xml")
	Stories map[string]bool

	// ParagraphStyles registra los IDs de estilos de párrafo referenciados
	// Clave: ID del estilo de párrafo (p. ej., "ParagraphStyle/$ID/[No paragraph style]")
	ParagraphStyles map[string]bool

	// CharacterStyles registra los IDs de estilos de carácter referenciados
	// Clave: ID del estilo de carácter (p. ej., "CharacterStyle/$ID/[No character style]")
	CharacterStyles map[string]bool

	// ObjectStyles registra los IDs de estilos de objeto referenciados
	// Clave: ID del estilo de objeto (p. ej., "ObjectStyle/$ID/[Normal]")
	ObjectStyles map[string]bool

	// Colors registra los IDs de colores referenciados
	// Clave: ID del color (p. ej., "Color/Black")
	Colors map[string]bool

	// Swatches registra los IDs de muestras de color referenciadas
	// Clave: ID de la muestra de color
	Swatches map[string]bool

	// Fonts registra las familias tipográficas referenciadas
	// Clave: nombre de la familia tipográfica (p. ej., "Minion Pro")
	Fonts map[string]bool

	// Layers registra los IDs de Layers referenciadas
	// Clave: ID de la Layer
	Layers map[string]bool

	// Links registra los vínculos a archivos externos referenciados (para imágenes)
	// Clave: ID o URI del vínculo
	Links map[string]bool

	// ColorSpaces registra los espacios de color referenciados (RGB, CMYK, Lab, etc.)
	// Clave: nombre del espacio de color
	ColorSpaces map[string]bool
}

// NewDependencySet crea un nuevo DependencySet vacío.
func NewDependencySet() *DependencySet {
	return &DependencySet{
		Stories:         make(map[string]bool),
		ParagraphStyles: make(map[string]bool),
		CharacterStyles: make(map[string]bool),
		ObjectStyles:    make(map[string]bool),
		Colors:          make(map[string]bool),
		Swatches:        make(map[string]bool),
		Fonts:           make(map[string]bool),
		Layers:          make(map[string]bool),
		Links:           make(map[string]bool),
		ColorSpaces:     make(map[string]bool),
	}
}

// DependencyTracker analiza elementos IDML y registra sus dependencias.
type DependencyTracker struct {
	// deps es el conjunto de dependencias que se está completando
	deps *DependencySet

	// pkg es el paquete IDML que se está analizando
	pkg *idml.Package
}

// NewDependencyTracker crea un nuevo DependencyTracker para el paquete proporcionado.
func NewDependencyTracker(pkg *idml.Package) *DependencyTracker {
	return &DependencyTracker{
		deps: NewDependencySet(),
		pkg:  pkg,
	}
}

// Dependencies devuelve el conjunto de dependencias recopilado.
func (dt *DependencyTracker) Dependencies() *DependencySet {
	return dt.deps
}

// AnalyzeTextFrame analiza un marco de texto y registra todas sus dependencias.
// Esto incluye:
// - El Story padre
// - El estilo de objeto aplicado al marco
// - La Layer en la que se encuentra el marco
func (dt *DependencyTracker) AnalyzeTextFrame(tf *spread.SpreadTextFrame) error {
	// Registrar el Story padre
	if tf.ParentStory != "" {
		// Las referencias a Story normalmente tienen el formato "u1d8"
		// Es necesario convertir esta referencia al nombre completo del archivo
		storyFilename := "Stories/Story_" + tf.ParentStory + ".xml"
		dt.deps.Stories[storyFilename] = true

		// Analizar el contenido del Story para encontrar dependencias de estilos
		story, err := dt.pkg.Story(storyFilename)
		if err == nil {
			if err := dt.AnalyzeStory(story); err != nil {
				// No generar un error; simplemente omitir este Story
				// Es posible que el Story no exista en el paquete
			}
		}
	}

	// Registrar el estilo de objeto aplicado
	if tf.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[tf.AppliedObjectStyle] = true
	}

	// Registrar la Layer
	if tf.ItemLayer != "" {
		dt.deps.Layers[tf.ItemLayer] = true
	}

	return nil
}

// AnalyzeStory analiza un Story y registra todas las dependencias de estilos.
// Esto incluye:
// - Los estilos de párrafo utilizados en el Story
// - Los estilos de carácter utilizados en el Story
// - Las fuentes referenciadas por los estilos (mejora futura)
// - Los colores utilizados en los estilos (mejora futura)
func (dt *DependencyTracker) AnalyzeStory(story *story.Story) error {
	// Analizar cada rango de estilo de párrafo
	for _, psr := range story.StoryElement.ParagraphStyleRanges {
		// Registrar el estilo de párrafo
		if psr.AppliedParagraphStyle != "" {
			dt.deps.ParagraphStyles[psr.AppliedParagraphStyle] = true
		}

		// Analizar cada rango de estilo de carácter dentro del párrafo
		for _, csr := range psr.CharacterStyleRanges {
			// Registrar el estilo de carácter
			if csr.AppliedCharacterStyle != "" {
				dt.deps.CharacterStyles[csr.AppliedCharacterStyle] = true
			}
		}
	}

	return nil
}

// AnalyzeRectangle analiza un rectángulo y registra todas sus dependencias.
// Esto incluye:
// - El estilo de objeto aplicado al rectángulo
// - La Layer en la que se encuentra el rectángulo
// - La imagen y sus dependencias, si el rectángulo contiene una imagen
//
// Nota: Rectangle no posee atributos directos StrokeColor/FillColor.
// Los colores se heredan de AppliedObjectStyle o se definen en el elemento Properties.
func (dt *DependencyTracker) AnalyzeRectangle(rect *spread.Rectangle) error {
	// Registrar el estilo de objeto aplicado
	if rect.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[rect.AppliedObjectStyle] = true
	}

	// Registrar la Layer
	if rect.ItemLayer != "" {
		dt.deps.Layers[rect.ItemLayer] = true
	}

	// Analizar la imagen si está presente
	if rect.Image != nil {
		if err := dt.AnalyzeImage(rect.Image); err != nil {
			return err
		}
	}

	return nil
}

// AnalyzeImage analiza una imagen y registra todas sus dependencias.
// Esto incluye:
// - El estilo de objeto aplicado a la imagen
// - El vínculo al archivo externo
// - El espacio de color utilizado por la imagen
func (dt *DependencyTracker) AnalyzeImage(img *spread.Image) error {
	// Registrar el estilo de objeto aplicado directamente a la imagen
	if img.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[img.AppliedObjectStyle] = true
	}

	// Registrar el espacio de color
	if img.Space != "" {
		dt.deps.ColorSpaces[img.Space] = true
	}

	// Registrar el vínculo si está presente
	if img.Link != nil {
		if img.Link.Self != "" {
			dt.deps.Links[img.Link.Self] = true
		}
		if img.Link.LinkResourceURI != "" {
			dt.deps.Links[img.Link.LinkResourceURI] = true
		}
	}

	return nil
}

// AnalyzeOval analiza un óvalo y registra todas sus dependencias.
// Esto incluye:
// - El estilo de objeto aplicado al óvalo
// - La Layer en la que se encuentra el óvalo
// - La imagen y sus dependencias, si el óvalo contiene una imagen
// - Los colores utilizados en el trazo y el relleno
func (dt *DependencyTracker) AnalyzeOval(oval *spread.Oval) error {
	// Registrar el estilo de objeto aplicado
	if oval.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[oval.AppliedObjectStyle] = true
	}

	// Registrar la Layer
	if oval.ItemLayer != "" {
		dt.deps.Layers[oval.ItemLayer] = true
	}

	// Registrar los colores del trazo y del relleno
	if oval.StrokeColor != "" {
		dt.deps.Colors[oval.StrokeColor] = true
	}
	if oval.FillColor != "" {
		dt.deps.Colors[oval.FillColor] = true
	}

	// Analizar la imagen si está presente
	if oval.Image != nil {
		if err := dt.AnalyzeImage(oval.Image); err != nil {
			return err
		}
	}

	return nil
}

// AnalyzePolygon analiza un polígono y registra todas sus dependencias.
// Esto incluye:
// - El estilo de objeto aplicado al polígono
// - La Layer en la que se encuentra el polígono
// - La imagen y sus dependencias, si el polígono contiene una imagen
// - Los colores utilizados en el trazo y el relleno
func (dt *DependencyTracker) AnalyzePolygon(polygon *spread.Polygon) error {
	// Registrar el estilo de objeto aplicado
	if polygon.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[polygon.AppliedObjectStyle] = true
	}

	// Registrar la Layer
	if polygon.ItemLayer != "" {
		dt.deps.Layers[polygon.ItemLayer] = true
	}

	// Registrar los colores del trazo y del relleno
	if polygon.StrokeColor != "" {
		dt.deps.Colors[polygon.StrokeColor] = true
	}
	if polygon.FillColor != "" {
		dt.deps.Colors[polygon.FillColor] = true
	}

	// Analizar la imagen si está presente
	if polygon.Image != nil {
		if err := dt.AnalyzeImage(polygon.Image); err != nil {
			return err
		}
	}

	return nil
}

// AnalyzeGraphicLine analiza una línea gráfica y registra todas sus dependencias.
// Esto incluye:
// - El estilo de objeto aplicado a la línea
// - La Layer en la que se encuentra la línea
// - Los colores utilizados en el trazo y el relleno
func (dt *DependencyTracker) AnalyzeGraphicLine(line *spread.GraphicLine) error {
	// Registrar el estilo de objeto aplicado
	if line.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[line.AppliedObjectStyle] = true
	}

	// Registrar la Layer
	if line.ItemLayer != "" {
		dt.deps.Layers[line.ItemLayer] = true
	}

	// Registrar el color del trazo
	if line.StrokeColor != "" {
		dt.deps.Colors[line.StrokeColor] = true
	}

	// Registrar el color de relleno (poco frecuente en líneas, pero posible)
	if line.FillColor != "" {
		dt.deps.Colors[line.FillColor] = true
	}

	return nil
}

// AnalyzeGroup analiza un grupo y registra todas sus dependencias.
// Esto incluye:
// - El estilo de objeto aplicado al grupo
// - La Layer en la que se encuentra el grupo
// Nota: el contenido de Group normalmente se analiza por separado
func (dt *DependencyTracker) AnalyzeGroup(group *spread.Group) error {
	// Registrar el estilo de objeto aplicado
	if group.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[group.AppliedObjectStyle] = true
	}

	// Registrar la Layer
	if group.ItemLayer != "" {
		dt.deps.Layers[group.ItemLayer] = true
	}

	return nil
}

// AnalyzeSelection analiza una selección completa y registra todas sus dependencias.
// Este es un método de conveniencia que invoca el método Analyze* correspondiente
// para cada elemento de la selección.
func (dt *DependencyTracker) AnalyzeSelection(sel *idml.Selection) error {
	// Analizar todos los marcos de texto
	for _, tf := range sel.TextFrames {
		if err := dt.AnalyzeTextFrame(tf); err != nil {
			return err
		}
	}

	// Analizar todos los rectángulos
	for _, rect := range sel.Rectangles {
		if err := dt.AnalyzeRectangle(rect); err != nil {
			return err
		}
	}

	// Analizar todos los óvalos
	for _, oval := range sel.Ovals {
		if err := dt.AnalyzeOval(oval); err != nil {
			return err
		}
	}

	// Analizar todos los polígonos
	for _, polygon := range sel.Polygons {
		if err := dt.AnalyzePolygon(polygon); err != nil {
			return err
		}
	}

	// Analizar todas las líneas gráficas
	for _, line := range sel.GraphicLines {
		if err := dt.AnalyzeGraphicLine(line); err != nil {
			return err
		}
	}

	// Analizar todos los grupos
	for _, group := range sel.Groups {
		if err := dt.AnalyzeGroup(group); err != nil {
			return err
		}
	}

	return nil
}

// ResolveStyleHierarchies recorre todas las dependencias de estilos recopiladas y agrega sus estilos padre.
// Esto garantiza que, al exportar un IDMS, se incluyan todos los estilos de la cadena de herencia.
// Por ejemplo, si un estilo de párrafo se basa en otro estilo, ambos deben incluirse.
//
// Este método gestiona:
// - La herencia de estilos de párrafo (relaciones BasedOn)
// - La herencia de estilos de carácter (relaciones BasedOn)
// - La herencia de estilos de objeto (relaciones BasedOn)
// - La detección de referencias circulares (para evitar bucles infinitos)
// - La herencia multinivel (estilos abuelo, etc.)
func (dt *DependencyTracker) ResolveStyleHierarchies() error {
	// Obtener el archivo de recursos Styles
	stylesResource, err := dt.pkg.Resource("Resources/Styles.xml")
	if err != nil {
		// Si no existe el archivo Styles, no hay nada que resolver
		return nil
	}

	// Interpretar la información de la jerarquía de estilos
	styleInfos, err := idml.ParseStylesForHierarchy(stylesResource.RawContent)
	if err != nil {
		return err
	}

	// Construir mapas de jerarquía de estilos para búsquedas rápidas
	styleParents := make(map[string]string) // styleID -> parentStyleID
	for _, info := range styleInfos {
		if info.BasedOn != "" {
			styleParents[info.Self] = info.BasedOn
		}
	}

	// Resolver las jerarquías de estilos de párrafo
	paragraphStylesToResolve := make([]string, 0, len(dt.deps.ParagraphStyles))
	for styleID := range dt.deps.ParagraphStyles {
		paragraphStylesToResolve = append(paragraphStylesToResolve, styleID)
	}
	for _, styleID := range paragraphStylesToResolve {
		if err := dt.resolveStyleChain(styleID, styleParents, dt.deps.ParagraphStyles); err != nil {
			return err
		}
	}

	// Resolver las jerarquías de estilos de carácter
	characterStylesToResolve := make([]string, 0, len(dt.deps.CharacterStyles))
	for styleID := range dt.deps.CharacterStyles {
		characterStylesToResolve = append(characterStylesToResolve, styleID)
	}
	for _, styleID := range characterStylesToResolve {
		if err := dt.resolveStyleChain(styleID, styleParents, dt.deps.CharacterStyles); err != nil {
			return err
		}
	}

	// Resolver las jerarquías de estilos de objeto
	objectStylesToResolve := make([]string, 0, len(dt.deps.ObjectStyles))
	for styleID := range dt.deps.ObjectStyles {
		objectStylesToResolve = append(objectStylesToResolve, styleID)
	}
	for _, styleID := range objectStylesToResolve {
		if err := dt.resolveStyleChain(styleID, styleParents, dt.deps.ObjectStyles); err != nil {
			return err
		}
	}

	return nil
}

// resolveStyleChain recorre recursivamente la jerarquía de estilos hacia arriba y agrega todos los estilos padre.
// Gestiona las referencias circulares mediante el registro de los estilos visitados.
func (dt *DependencyTracker) resolveStyleChain(styleID string, styleParents map[string]string, targetMap map[string]bool) error {
	// Registrar los estilos visitados para detectar referencias circulares
	visited := make(map[string]bool)
	current := styleID

	for {
		// Comprobar si este estilo ya fue visitado (referencia circular)
		if visited[current] {
			// Referencia circular detectada; detener el recorrido aquí
			break
		}
		visited[current] = true

		// Obtener el estilo padre
		parent, hasParent := styleParents[current]
		if !hasParent {
			// No existe un padre; se alcanzó la parte superior de la jerarquía
			break
		}

		// Comprobar si el padre es un estilo integrado de InDesign (comienza con $ID/)
		// Estos estilos siempre están disponibles y no es necesario incluirlos en las dependencias
		if len(parent) > 4 && parent[:4] == "$ID/" {
			// Estilo integrado; detener el recorrido aquí
			break
		}

		// Agregar el estilo padre a las dependencias
		targetMap[parent] = true

		// Avanzar al siguiente estilo padre
		current = parent
	}

	return nil
}

// Summary devuelve un resumen de las dependencias encontradas.
func (dt *DependencyTracker) Summary() DependencySummary {
	return DependencySummary{
		StoriesCount:         len(dt.deps.Stories),
		ParagraphStylesCount: len(dt.deps.ParagraphStyles),
		CharacterStylesCount: len(dt.deps.CharacterStyles),
		ObjectStylesCount:    len(dt.deps.ObjectStyles),
		ColorsCount:          len(dt.deps.Colors),
		SwatchesCount:        len(dt.deps.Swatches),
		FontsCount:           len(dt.deps.Fonts),
		LayersCount:          len(dt.deps.Layers),
		LinksCount:           len(dt.deps.Links),
		ColorSpacesCount:     len(dt.deps.ColorSpaces),
	}
}

// DependencySummary proporciona el conteo de cada tipo de dependencia.
type DependencySummary struct {
	StoriesCount         int
	ParagraphStylesCount int
	CharacterStylesCount int
	ObjectStylesCount    int
	ColorsCount          int
	SwatchesCount        int
	FontsCount           int
	LayersCount          int
	LinksCount           int
	ColorSpacesCount     int
}
