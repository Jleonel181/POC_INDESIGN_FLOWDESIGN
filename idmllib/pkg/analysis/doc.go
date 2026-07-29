// Package analysis provee herramientas para analizar documentos IDML y rastrear dependencias.
//
// Este paquete se usa principalmente para la exportación IDMS, para recolectar todos los recursos
// necesarios para los page items seleccionados. Analiza el grafo de dependencias de los elementos
// IDML para asegurar que los snippets exportados contengan todos los estilos, fuentes, colores y
// otros recursos necesarios.
//
// # Tipos principales
//
//   - DependencySet: Rastrea todas las dependencias de un conjunto de page items
//   - DependencyTracker: Analiza elementos IDML y construye conjuntos de dependencias
//   - DependencySummary: Provee conteos de cada tipo de dependencia
//
// # Uso
//
// Analizar dependencias para una selección:
//
//	// Crear un tracker para el paquete IDML
//	tracker := analysis.NewDependencyTracker(pkg)
//
//	// Analizar una selección de page items
//	selection := idml.NewSelection()
//	selection.AddTextFrame(textFrame)
//	selection.AddRectangle(rectangle)
//
//	err := tracker.AnalyzeSelection(selection)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Resolver jerarquías de estilos para incluir estilos padre
//	err = tracker.ResolveStyleHierarchies()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Obtener el conjunto de dependencias
//	deps := tracker.Dependencies()
//	fmt.Printf("Encontrados %d stories, %d estilos\n",
//	    len(deps.Stories), len(deps.ParagraphStyles))
//
// # Tipos de dependencias
//
// El tracker identifica estos tipos de dependencias:
//   - Stories: Archivos de story referenciados por nombre de archivo
//   - ParagraphStyles: IDs de estilos de párrafo referenciados
//   - CharacterStyles: IDs de estilos de carácter referenciados
//   - ObjectStyles: IDs de estilos de objeto referenciados
//   - Colors: IDs de colores referenciados
//   - Swatches: IDs de muestras de color referenciadas
//   - Fonts: Familias de fuentes referenciadas
//   - Layers: IDs de capas referenciadas
//   - Links: Enlaces a archivos externos referenciados (para imágenes)
//   - ColorSpaces: Espacios de color referenciados (RGB, CMYK, Lab, etc.)
//
// # Resolución de jerarquías de estilos
//
// El método ResolveStyleHierarchies recorre las cadenas de herencia de estilos
// para asegurar que al exportar un IDMS, todos los estilos en la jerarquía BasedOn
// estén incluidos. Esto contempla:
//   - Herencia multinivel (estilos abuelos, etc.)
//   - Detección de referencias circulares (para evitar bucles infinitos)
//   - Estilos nativos de InDesign (que no necesitan ser incluidos)
//
// # Arquitectura
//
// Este paquete es parte de la arquitectura domain-specific que soporta
// la funcionalidad de exportación IDMS. Trabaja en conjunto con:
//   - pkg/idml: Para acceder al contenido del paquete IDML
//   - pkg/spread: Para analizar page items
//   - pkg/story: Para analizar contenido de texto
//   - pkg/resources: Para información de jerarquías de estilos
package analysis
