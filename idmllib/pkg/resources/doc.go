// Package resources provee tipos y funciones para trabajar con los archivos de recursos IDML.
//
// Los archivos de recursos contienen las definiciones de estilos y formato utilizadas
// en todo un documento IDML. Esto incluye fuentes, colores, estilos de párrafo,
// estilos de carácter, estilos de objeto y configuraciones gráficas.
//
// # Archivos de recursos
//
// Los paquetes IDML contienen varios archivos de recursos en el directorio Resources/:
//   - Fonts.xml: Definiciones de familias tipográficas y fuentes
//   - Styles.xml: Definiciones de estilos de párrafo, carácter y objeto
//   - Graphic.xml: Colores, muestras, degradados y estilos de trazo
//   - Preferences.xml: Preferencias y configuraciones a nivel de documento
//   - Tags.xml: Definiciones de etiquetado XML (opcional)
//
// # Tipos principales
//
// ## Recursos de fuentes
//   - FontsFile: Contenedor raíz para definiciones de fuentes
//   - FontFamily: Agrupa fuentes por nombre de familia (ej., "Minion Pro")
//   - Font: Fuente individual con estilo, métricas y estado de instalación
//   - CompositeFont: Definiciones de fuentes multi-script (principalmente CJK)
//
// ## Recursos de estilos
//   - StylesFile: Contenedor raíz para todas las definiciones de estilos
//   - ParagraphStyleGroup: Organización jerárquica de estilos de párrafo
//   - CharacterStyleGroup: Organización jerárquica de estilos de carácter
//   - ObjectStyleGroup: Organización jerárquica de estilos de objeto
//   - ParagraphStyle, CharacterStyle, ObjectStyle: Definiciones de estilos individuales
//
// ## Recursos gráficos
//   - GraphicFile: Contenedor raíz para colores y configuraciones gráficas
//   - Color: Definiciones de color (RGB, CMYK, Lab, Spot)
//   - Swatch: Muestras de color y tintes
//   - Gradient: Definiciones de degradados
//   - StrokeStyle: Patrones y estilos de trazo personalizados
//
// # Uso
//
// Parsear un archivo de recursos de fuentes:
//
//	data, _ := os.ReadFile("Resources/Fonts.xml")
//	fonts, err := resources.ParseFontsFile(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Acceder a las familias tipográficas
//	for _, family := range fonts.FontFamilies {
//	    fmt.Printf("Familia: %s\n", family.Name)
//	    for _, font := range family.Fonts {
//	        fmt.Printf("  Fuente: %s (%s)\n", font.FontStyleName, font.Status)
//	    }
//	}
//
// Parsear un archivo de recursos de estilos:
//
//	data, _ := os.ReadFile("Resources/Styles.xml")
//	styles, err := resources.ParseStylesFile(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Acceder a los estilos de párrafo
//	for _, style := range styles.RootParagraphStyleGroup.ParagraphStyles {
//	    fmt.Printf("Estilo de párrafo: %s\n", style.Name)
//	}
//
// Serializar de vuelta a XML:
//
//	xmlData, err := resources.MarshalFontsFile(fonts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", xmlData, 0644)
//
// # Manejo de namespaces
//
// Los archivos de recursos usan el contenedor del namespace idPkg:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<idPkg:Fonts xmlns:idPkg="..." DOMVersion="20.4">
//	  <FontFamily Self="..." Name="Minion Pro">
//	    <Font Self="..." FontStyleName="Regular" .../>
//	  </FontFamily>
//	</idPkg:Fonts>
//
// Los métodos personalizados UnmarshalXML/MarshalXML manejan este contenedor correctamente.
//
// # Jerarquías de estilos
//
// Los estilos soportan herencia mediante relaciones BasedOn:
//   - Los estilos de párrafo pueden basarse en otros estilos de párrafo
//   - Los estilos de carácter pueden basarse en otros estilos de carácter
//   - Los estilos de objeto pueden basarse en otros estilos de objeto
//
// El paquete pkg/analysis provee herramientas para resolver estas jerarquías
// al exportar snippets IDMS.
//
// # Estado de fuentes
//
// Las fuentes tienen indicadores de estado:
//   - "Installed": La fuente está disponible en el sistema
//   - "Substituted": La fuente fue sustituida por una similar
//   - "NotAvailable": La fuente no está disponible y debe instalarse
//
// # Espacios de color
//
// Los colores soportan múltiples espacios de color:
//   - RGB: Rojo, Verde, Azul (colores de pantalla)
//   - CMYK: Cian, Magenta, Amarillo, Negro (colores de impresión)
//   - Lab: Luminosidad, A, B (colores independientes del dispositivo)
//   - Spot: Colores especiales con nombre para tintas especiales
//
// # Arquitectura
//
// Este paquete es parte de la arquitectura domain-driven:
//   - pkg/common: Tipos y utilidades compartidas
//   - pkg/document: Metadatos y estructura del documento
//   - pkg/spread: Maquetación de páginas y elementos de página
//   - pkg/story: Contenido de texto y formato
//   - pkg/resources: Definiciones de estilos y recursos (este paquete)
//   - pkg/analysis: Seguimiento y análisis de dependencias
//   - pkg/idms: Funcionalidad de exportación IDMS
package resources
