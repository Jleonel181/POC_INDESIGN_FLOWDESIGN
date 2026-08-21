// cmd/idmlgen genera un archivo IDML a partir de una descripción de documento JSON
// recibida por stdin. Escribe el IDML en stdout (por defecto) o en la ruta indicada
// por -out.
//
// Códigos de salida:
//   - 0: generación exitosa
//   - 1: error de entrada (JSON inválido, validación fallida)
//   - 2: error de generación (fallo interno de idmllib)
//
// Ejemplo:
//
//	cat documento.json | idmlgen > edicion.idml
//	cat documento.json | idmlgen -out edicion.idml
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dimelords/idmllib/v2/internal/idmlgen"
	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
)

func main() {
	outPath := flag.String("out", "", "ruta de salida del archivo IDML (por defecto: stdout)")
	_ = flag.String("base-dir", "", "directorio base para imágenes locales (obligatorio si se usan rutas)")
	flag.Parse()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: no se pudo leer stdin: %v\n", err)
		os.Exit(1)
	}

	var input idmlgen.DocumentInput
	if err := json.Unmarshal(data, &input); err != nil {
		if synErr, ok := err.(*json.SyntaxError); ok {
			fmt.Fprintf(os.Stderr, "error: JSON inválido en posición %d: %v\n", synErr.Offset, err)
		} else {
			fmt.Fprintf(os.Stderr, "error: JSON inválido: %v\n", err)
		}
		os.Exit(1)
	}

	if err := idmlgen.Validate(&input); err != nil {
		fmt.Fprintf(os.Stderr, "error de validación: %v\n", err)
		os.Exit(1)
	}

	pkg, err := idmlgen.Generate(&input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error de generación: %v\n", err)
		os.Exit(2)
	}

	if *outPath != "" {
		if err := idmlpkg.Write(pkg, *outPath); err != nil {
			fmt.Fprintf(os.Stderr, "error al escribir %s: %v\n", *outPath, err)
			os.Exit(2)
		}
	} else {
		if err := idmlpkg.WriteTo(pkg, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "error al escribir a stdout: %v\n", err)
			os.Exit(2)
		}
	}
}
