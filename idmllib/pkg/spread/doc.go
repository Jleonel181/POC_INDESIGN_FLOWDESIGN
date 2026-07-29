// Package spread proporciona tipos y funciones para trabajar con spreads e items de página IDML.
//
// El paquete spread contiene todos los tipos relacionados con el layout de spreads en archivos IDML, incluyendo:
//   - Spread: El contenedor spread raíz con información de layout de página
//   - SpreadElement: El contenido spread actual con páginas e items de página
//   - Page: Páginas individuales dentro de un spread
//   - SpreadTextFrame: Text frames en spreads
//   - Rectangle: Frames rectangulares (pueden contener texto, imágenes, o estar vacíos)
//   - Image: Imágenes vinculadas en frames
//   - GraphicLine: Elementos de línea vectorial
//   - Oval, Polygon, Group: Otros tipos de page items
//   - Funciones de parseo: ParseSpread, MarshalSpread
//
// # Arquitectura
//
//   - pkg/common: Tipos compartidos entre todos los paquetes
//   - pkg/document: Metadatos del documento
//   - pkg/spread: Tipos de spread y layout de página
//   - pkg/story: Contenido de texto y tipos de story
//   - pkg/resources: Estilos, fuentes, gráficos
//
// # Uso
//
// Parsear un archivo XML de spread:
//
//	data, _ := os.ReadFile("Spreads/Spread_u210.xml")
//	spread, err := spread.ParseSpread(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Páginas:", len(spread.InnerSpread.Pages))
//	fmt.Println("Text frames:", len(spread.InnerSpread.TextFrames))
//	fmt.Println("Rectángulos:", len(spread.InnerSpread.Rectangles))
//
// Serializar de vuelta a XML:
//
//	xmlData, err := spread.MarshalSpread(spread)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", xmlData, 0644)
//
// # Manejo de Namespaces
//
// Los spreads usan el wrapper de namespace idPkg:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<idPkg:Spread xmlns:idPkg="..." DOMVersion="20.4">
//	  <Spread Self="u210" ...>
//	    <Page Self="u211" .../>
//	    <TextFrame Self="uf3" .../>
//	  </Spread>
//	</idPkg:Spread>
//
// Los métodos custom UnmarshalXML/MarshalXML manejan este wrapper correctamente.
package spread
