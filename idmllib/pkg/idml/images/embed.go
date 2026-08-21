package images

import (
	"encoding/base64"
	"strings"
)

// EncodeBase64Lines codifica bytes en base64 estándar (RFC 4648), ajustando a 76
// caracteres por línea. Sin salto de línea inicial ni final.
//
// Es el formato que InDesign usa en Properties/Contents para imágenes embebidas.
func EncodeBase64Lines(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	if len(encoded) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(encoded) + len(encoded)/76) // espacio para los saltos de línea

	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(encoded[i:end])
	}

	return sb.String()
}

// DecodeBase64Lines decodifica una cadena base64 que puede tener saltos de línea.
func DecodeBase64Lines(s string) ([]byte, error) {
	// Eliminar saltos de línea para decodificar.
	clean := strings.ReplaceAll(s, "\n", "")
	clean = strings.ReplaceAll(clean, "\r", "")
	return base64.StdEncoding.DecodeString(clean)
}
