package xmlutil

import (
	"bytes"
	"regexp"
)

// CompactEmptyElements convierte elementos XML vacíos a etiquetas auto-cerradas.
// Convierte <tag></tag> a <tag /> que es el formato que usa IDML.
//
// Ejemplo:
//
//	Entrada: <KeyValuePair Key="test" Value="1"></KeyValuePair>
//	Salida:  <KeyValuePair Key="test" Value="1" />
//
// Esta función es esencial para mantener compatibilidad byte a byte
// con archivos IDML generados por Adobe InDesign, que usa etiquetas auto-cerradas
// para elementos vacíos.
func CompactEmptyElements(xml []byte) []byte {
	// Nota: la regexp de Go no soporta backreferences, por lo que no se puede usar un patrón
	// simple como <([^>]+)></\1>. En su lugar, se hace match del patrón general
	// y luego se verifica que los nombres de etiqueta coincidan.
	//
	// El patrón hace match de: <tagname atributos></...>
	// donde puede no haber contenido entre las etiquetas
	pattern := regexp.MustCompile(`<([^>\s]+)([^>]*)></([^>\s]+)>`)

	// Replace with self-closing format if opening and closing tags match
	result := pattern.ReplaceAllFunc(xml, func(match []byte) []byte {
		// Extraer nombres de etiqueta y atributos
		submatches := pattern.FindSubmatch(match)
		if len(submatches) != 4 {
			return match // no debería ocurrir, pero es seguro manejarlo
		}

		openTag := submatches[1]
		attrs := submatches[2]
		closeTag := submatches[3]

		// Solo convertir a auto-cerrada si las etiquetas coinciden
		if !bytes.Equal(openTag, closeTag) {
			return match
		}

		// Construir la etiqueta auto-cerrada
		// Si attrs está vacío o termina con espacio, no se necesita espacio extra
		// Si attrs no termina con espacio, agregar uno antes de />
		if len(attrs) == 0 || attrs[len(attrs)-1] == ' ' {
			result := append([]byte("<"), openTag...)
			result = append(result, attrs...)
			result = append(result, []byte("/>")...)
			return result
		}

		result := append([]byte("<"), openTag...)
		result = append(result, attrs...)
		result = append(result, []byte(" />")...)
		return result
	})

	return result
}
