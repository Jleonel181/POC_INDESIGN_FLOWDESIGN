package images

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestEncodeBase64Lines_Roundtrip(t *testing.T) {
	// Datos de prueba de diferentes tamaños
	sizes := []int{0, 1, 57, 76, 100, 200, 1000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		encoded := EncodeBase64Lines(data)
		decoded, err := DecodeBase64Lines(encoded)
		if err != nil {
			t.Errorf("size=%d: DecodeBase64Lines falló: %v", size, err)
			continue
		}

		if len(decoded) != len(data) {
			t.Errorf("size=%d: longitud decodificada %d != original %d", size, len(decoded), len(data))
			continue
		}

		// SHA-256 debe coincidir
		origHash := sha256.Sum256(data)
		decodedHash := sha256.Sum256(decoded)
		if origHash != decodedHash {
			t.Errorf("size=%d: SHA-256 no coincide", size)
		}
	}
}

func TestEncodeBase64Lines_LineLength(t *testing.T) {
	// Datos suficientes para varias líneas
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i)
	}

	encoded := EncodeBase64Lines(data)
	if encoded == "" {
		t.Fatal("encoded vacío")
	}

	// No debe empezar ni terminar con salto de línea
	if encoded[0] == '\n' {
		t.Error("empieza con salto de línea")
	}
	if encoded[len(encoded)-1] == '\n' {
		t.Error("termina con salto de línea")
	}

	// Todas las líneas excepto la última deben tener 76 caracteres
	lines := strings.Split(encoded, "\n")
	if len(lines) < 2 {
		t.Fatal("esperaba al menos 2 líneas")
	}
	for i, line := range lines[:len(lines)-1] {
		if len(line) != 76 {
			t.Errorf("línea %d: %d caracteres, esperado 76", i, len(line))
		}
	}
	// La última línea puede ser más corta
	lastLine := lines[len(lines)-1]
	if len(lastLine) > 76 {
		t.Errorf("última línea: %d caracteres, máximo 76", len(lastLine))
	}
}

func TestEncodeBase64Lines_LinkResourceSize(t *testing.T) {
	data := make([]byte, 12345)
	encoded := base64.StdEncoding.EncodeToString(data)

	// Simular lo que hace el resolver
	linkResourceSize := fmt.Sprintf("0~%x", len(data))
	expectedHex := "0~3039" // 12345 en hex
	if linkResourceSize != expectedHex {
		t.Errorf("LinkResourceSize: %q, esperado %q", linkResourceSize, expectedHex)
	}

	// Verificar que decodificar el encoded da los mismos bytes
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 12345 {
		t.Errorf("longitud decodificada: %d, esperado 12345", len(decoded))
	}
}

func TestEncodeBase64Lines_Empty(t *testing.T) {
	result := EncodeBase64Lines(nil)
	if result != "" {
		t.Errorf("esperaba cadena vacía para nil, obtuvo %q", result)
	}
	result = EncodeBase64Lines([]byte{})
	if result != "" {
		t.Errorf("esperaba cadena vacía para []byte{}, obtuvo %q", result)
	}
}
