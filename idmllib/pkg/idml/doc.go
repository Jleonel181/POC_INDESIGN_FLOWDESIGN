// Package idml provee funcionalidad para leer, escribir y manipular
// archivos Adobe InDesign IDML (InDesign Markup Language).
//
// IDML es el formato de archivo basado en XML de Adobe InDesign, estructurado como
// un archivo ZIP que contiene archivos XML que definen la estructura, el contenido
// y el estilo del documento.
//
// Esta librería se enfoca en:
//   - Leer archivos IDML con total fidelidad
//   - Escribir archivos IDML que InDesign pueda abrir
//   - Preservar todo el contenido durante operaciones de roundtrip
//   - Proveer una API limpia y con tipos seguros para la manipulación de documentos
//
// Uso básico:
//
//	// Leer un archivo IDML
//	pkg, err := idml.Read("document.idml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Escribirlo de vuelta
//	err = idml.Write(pkg, "output.idml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// La librería está diseñada para manejar archivos IDML en fases:
//   - Fase 1: Roundtrip sin procesamiento (leer y escribir sin parseo completo)
//   - Fase 2: Parseo completo con estructuras de tipos seguros
//   - Fase 3: API de modificación de contenido
//   - Fase 4: Exportación de snippets IDMS
//
// # Construcciones fuera de alcance
//
// Las siguientes construcciones de InDesign se preservan en el roundtrip (vía
// OtherElements/OtherAttrs) pero NO se generan por el constructor de documentos:
//
//   - Table, Cell (tablas y celdas)
//   - Footnote (notas al pie)
//   - Hyperlink (hipervínculos)
//   - CrossReferenceSource, CrossReferenceFormat (referencias cruzadas)
//   - KinsokuTable, MojikumiTable (reglas tipográficas CJK)
//
// Un JSON de entrada que declare una de estas construcciones es rechazado por el
// validador con código de salida 1.
package idml
