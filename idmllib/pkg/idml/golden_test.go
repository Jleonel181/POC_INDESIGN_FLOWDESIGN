// Arnés de fidelidad estructural sobre el corpus versionado en testdata/.
//
// El arnés está en el paquete idml (y no en idml_test) porque necesita los bytes
// originales de cada archivo del paquete para compararlos contra el resultado de
// re-serializar, y no hay accesor exportado para eso. Es el mismo motivo por el
// que xml_test.go vive aquí.
package idml

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/internal/testutil"
	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Rutas del corpus de fidelidad. Cada elemento se resuelve desde su constante y
// se puede sobreescribir por variable de entorno para apuntar el arnés a otra
// copia del corpus sin recompilar.
const (
	// CorpusReferenceDir es el Documento_Referencia: se versiona descomprimido
	// (43 archivos .xml más mimetype) porque no existe un .idml original suyo, y
	// porque en forma de texto git almacena diferencias y se puede hacer grep.
	// Se recorre como árbol de archivos.
	CorpusReferenceDir = "../../testdata/documento_referencia"

	// CorpusImagesFixture es el Archivo_Evidencia_Imagenes: se versiona como
	// paquete .idml para que el arnés ejercite también la apertura de un ZIP por
	// la ruta de apertura de paquete de la librería, que es lo que hace un
	// llamador real. Es la única fuente de verdad del formato de imagen embebida.
	CorpusImagesFixture = "../../testdata/archivo_evidencia_imagenes.idml"

	// Variables de entorno que sobreescriben las rutas anteriores.
	EnvCorpusReferenceDir  = "IDMLLIB_REFERENCE_DIR"
	EnvCorpusImagesFixture = "IDMLLIB_IMAGES_FIXTURE"
)

// corpusPath resuelve la ruta de un elemento del corpus, dando prioridad a la
// variable de entorno cuando trae un valor no vacío.
func corpusPath(env, def string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return def
}

// fidelityTally acumula el resultado del ciclo por origen del corpus.
type fidelityTally struct {
	origin    string
	clean     int // archivos con 0 diferencias
	differing int // archivos con al menos 1 diferencia
	unparsed  int // archivos sin modelo tipado, copiados sin parsear
	failed    int // archivos que fallaron al parsear o al serializar
}

func (f *fidelityTally) total() int {
	return f.clean + f.differing + f.unparsed + f.failed
}

func (f *fidelityTally) report(t *testing.T) {
	t.Helper()
	t.Logf("resumen [%s]: %d sin diferencias, %d con diferencias, %d copiados sin parsear, %d con fallo, %d en total",
		f.origin, f.clean, f.differing, f.unparsed, f.failed, f.total())
}

// roundtripXML aplica el ciclo parseo → serialización usando el mismo par tipado
// que la librería usa para esa ruta al leer y escribir un paquete.
//
// typed es false cuando la ruta todavía no tiene modelo tipado y el paquete la
// copia sin parsear; hoy es el caso de MasterSpreads/*.xml.
func roundtripXML(path string, data []byte) (out []byte, typed bool, err error) {
	switch {
	case path == PathDesignmap:
		parsed, err := document.ParseDocumentWithMetadata(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = document.MarshalDocumentWithMetadata(parsed)

	case IsSpreadPath(path):
		parsed, err := spread.ParseSpread(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = spread.MarshalSpread(parsed)

	case IsStoryPath(path):
		parsed, err := story.ParseStory(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = story.MarshalStory(parsed)

	case path == PathFonts:
		parsed, err := resources.ParseFontsFile(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = resources.MarshalFontsFile(parsed)

	case path == PathGraphic:
		parsed, err := resources.ParseGraphicFile(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = resources.MarshalGraphicFile(parsed)

	case path == PathStyles:
		parsed, err := resources.ParseStylesFile(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = resources.MarshalStylesFile(parsed)

	case IsResourcePath(path):
		parsed, err := ParseResourceFile(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = MarshalResourceFile(parsed)

	case IsMetaInfPath(path) || IsXMLPath(path):
		parsed, err := ParseMetadataFile(path, data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = MarshalMetadataFile(parsed)

	default:
		return nil, false, nil
	}

	if err != nil {
		return nil, true, fmt.Errorf("serialización: %w", err)
	}
	return out, true, nil
}

// compareRoundtrip ejecuta el ciclo sobre un archivo y acumula el resultado.
//
// Las diferencias se registran como diagnóstico, no como fallo: hoy son la
// medición que las tareas de modelo tienen que llevar a cero. La Tarea 22 es la
// que convierte una diferencia en fallo, cuando la cifra ya sea 0.
func compareRoundtrip(t *testing.T, tally *fidelityTally, path string, data []byte) {
	t.Helper()

	out, typed, err := roundtripXML(path, data)
	switch {
	case err != nil:
		tally.failed++
		t.Errorf("[%s] %s: %v", tally.origin, path, err)
		return
	case !typed:
		tally.unparsed++
		return
	}

	diffs, err := xmlutil.CompareXMLWithDetails(data, out, xmlutil.DefaultCompareOptions())
	if err != nil {
		tally.failed++
		t.Errorf("[%s] %s: comparación: %v", tally.origin, path, err)
		return
	}

	if len(diffs) == 0 {
		tally.clean++
		return
	}
	tally.differing++
	t.Logf("[%s] %s: %d diferencia(s)\n%s", tally.origin, path, len(diffs), xmlutil.FormatDifferences(diffs))
}

// TestGoldenRoundtrip_ExampleIDML ejecuta el ciclo parseo → serialización →
// comparación estructural sobre los dos elementos del corpus versionado, cada
// uno por su propia ruta de lectura: el Documento_Referencia como árbol de
// archivos descomprimido y el Archivo_Evidencia_Imagenes como paquete .idml.
func TestGoldenRoundtrip_ExampleIDML(t *testing.T) {
	t.Run("documento_referencia", testFidelityReferenceDir)
	t.Run("archivo_evidencia_imagenes", testFidelityImagesFixture)
}

// testFidelityReferenceDir recorre el Documento_Referencia como directorio
// descomprimido: todos los .xml de forma recursiva, con lo que mimetype queda
// fuera por extensión.
func testFidelityReferenceDir(t *testing.T) {
	dir := corpusPath(EnvCorpusReferenceDir, CorpusReferenceDir)

	if _, err := os.Stat(dir); err != nil {
		t.Skipf("Documento_Referencia ausente en la copia de trabajo: %v (ruta esperada: %s)", err, absPathForMsg(dir))
	}

	var paths []string
	root := filepath.Clean(dir)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ExtXML) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("recorrido de %s falló: %v", absPathForMsg(root), err)
	}

	if len(paths) == 0 {
		t.Fatalf("el Documento_Referencia existe pero el recorrido seleccionó 0 archivos %s: %s", ExtXML, absPathForMsg(root))
	}
	sort.Strings(paths)

	tally := &fidelityTally{origin: "documento_referencia"}
	for _, rel := range paths {
		// #nosec G304 - ruta derivada del recorrido del corpus de prueba
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			tally.failed++
			t.Errorf("[%s] %s: lectura: %v", tally.origin, rel, err)
			continue
		}
		compareRoundtrip(t, tally, rel, data)
	}
	tally.report(t)
}

// testFidelityImagesFixture abre el Archivo_Evidencia_Imagenes por la ruta de
// apertura de paquete de la librería y aplica el mismo ciclo a sus .xml.
func testFidelityImagesFixture(t *testing.T) {
	path := corpusPath(EnvCorpusImagesFixture, CorpusImagesFixture)

	if _, err := os.Stat(path); err != nil {
		t.Skipf("Archivo_Evidencia_Imagenes ausente en la copia de trabajo: %v (ruta esperada: %s)", err, absPathForMsg(path))
	}

	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read(%s) falló: %v", absPathForMsg(path), err)
	}

	names := make([]string, 0, pkg.FileCount())
	for _, name := range pkg.Files() {
		if strings.EqualFold(filepath.Ext(name), ExtXML) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		t.Fatalf("el paquete se abrió pero no contiene ningún archivo %s: %s", ExtXML, absPathForMsg(path))
	}
	sort.Strings(names)

	tally := &fidelityTally{origin: "archivo_evidencia_imagenes"}
	for _, name := range names {
		data, err := pkg.getFileData(name)
		if err != nil {
			tally.failed++
			t.Errorf("[%s] %s: lectura: %v", tally.origin, name, err)
			continue
		}
		compareRoundtrip(t, tally, name, data)
	}
	tally.report(t)
}

// absPathForMsg devuelve la ruta absoluta para los mensajes de diagnóstico, de
// modo que una ruta mal configurada sea identificable. Si no se puede resolver,
// devuelve la ruta tal cual.
func absPathForMsg(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// TestGoldenRoundtrip_StructuralComparison tests XML-level structural comparison.
func TestGoldenRoundtrip_StructuralComparison(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		xmlPath  string
	}{
		{
			name:     "designmap.xml",
			filename: "plain.idml",
			xmlPath:  "designmap.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read IDML
			inputPath := testutil.TestDataPath(t, tt.filename)
			pkg, err := Read(inputPath)
			if err != nil {
				t.Fatalf("Read() failed: %v", err)
			}

			// Write back
			outputPath := testutil.TempIDML(t, tt.filename)
			if err := Write(pkg, outputPath); err != nil {
				t.Fatalf("Write() failed: %v", err)
			}

			// Re-read to get XML
			pkg2, err := Read(outputPath)
			if err != nil {
				t.Fatalf("Read(output) failed: %v", err)
			}

			// For this test, we'd need to expose a method to get file contents
			// For now, this demonstrates the pattern
			_ = pkg2
		})
	}
}
