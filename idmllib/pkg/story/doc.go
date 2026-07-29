// Package story provee tipos y funciones para trabajar con archivos XML de Story de InDesign.
//
// Las stories son los contenedores de texto principales en documentos InDesign. Contienen párrafos,
// formato de caracteres y todo el contenido de texto.
//
// # Arquitectura
//
// Los archivos de story usan una estructura de doble envoltura con namespace idPkg:
//
//	<idPkg:Story DOMVersion="...">
//	  <Story Self="...">
//	    <ParagraphStyleRange AppliedParagraphStyle="...">
//	      <CharacterStyleRange AppliedCharacterStyle="...">
//	        <Content>Texto aquí</Content>
//	        <Br/>
//	      </CharacterStyleRange>
//	    </ParagraphStyleRange>
//	  </Story>
//	</idPkg:Story>
//
// # Tipos principales
//
//   - Story: El wrapper externo con DOMVersion y namespace
//   - StoryElement: El elemento interno <Story> que contiene el contenido real
//   - ParagraphStyleRange: Agrupa caracteres por estilo de párrafo
//   - CharacterStyleRange: Agrupa texto por estilo de carácter con marshal/unmarshal custom
//   - Content: Contenido de texto real
//   - Br: Elemento de salto de línea
//
// # Uso
//
// Parsear un archivo XML de story:
//
//	data, err := os.ReadFile("Stories/Story_u12a.xml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	story, err := story.ParseStory(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Acceder al contenido de la story
//	for _, psr := range story.StoryElement.ParagraphStyleRanges {
//	    for _, csr := range psr.CharacterStyleRanges {
//	        contents := csr.GetContent()
//	        for _, content := range contents {
//	            fmt.Println(content.Text)
//	        }
//	    }
//	}
//
// Serializar una story de vuelta a XML:
//
//	data, err := story.MarshalStory(&story)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", data, 0644)
//
// # Marshaling custom
//
// CharacterStyleRange usa UnmarshalXML/MarshalXML custom para preservar el orden exacto
// de los elementos Content y Br, lo cual es crítico para la compatibilidad con InDesign.
// El campo Children almacena el contenido mixto en orden.
//
// # Compatibilidad hacia atrás
//
// Los métodos helper GetContent(), SetContent() y AddContent() proveen compatibilidad
// hacia atrás para código que no necesita manejar el orden del contenido mixto.
package story
