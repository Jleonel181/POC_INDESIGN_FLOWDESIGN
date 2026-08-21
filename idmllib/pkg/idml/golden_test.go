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
	"strconv"
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

	// Los tres IDML siguientes ya estaban versionados en testdata/ y el arnés no
	// los recorría. Se incorporaron al corpus porque medir contra dos documentos
	// resultó insuficiente dos veces: cada uno de ellos contiene formas que los dos
	// primeros no tienen, y un criterio de aceptación del tipo «cero diferencias de
	// categoría X» es falso si el corpus no contiene la forma que falla.
	//
	// Van de menor a mayor complejidad: plain es un documento casi vacío, example
	// tiene más recursos y estilos, y tripple tiene tres spreads.
	CorpusPlainIDML   = "../../testdata/plain.idml"
	CorpusExampleIDML = "../../testdata/example.idml"
	CorpusTrippleIDML = "../../testdata/tripple.idml"

	// Variables de entorno que sobreescriben las rutas anteriores.
	EnvCorpusReferenceDir  = "IDMLLIB_REFERENCE_DIR"
	EnvCorpusImagesFixture = "IDMLLIB_IMAGES_FIXTURE"
	EnvCorpusPlainIDML     = "IDMLLIB_PLAIN_IDML"
	EnvCorpusExampleIDML   = "IDMLLIB_EXAMPLE_IDML"
	EnvCorpusTrippleIDML   = "IDMLLIB_TRIPPLE_IDML"

	// DefaultMaxDiffsPerFile es el tope de diferencias a recolectar por archivo que
	// fija el Req 1, criterio 2. Al alcanzarlo, el reporte de ese archivo se marca
	// como truncado.
	DefaultMaxDiffsPerFile = 100

	// EnvMaxDiffsPerFile levanta o baja ese tope sin recompilar. Con valor 0 no hay
	// límite, que es la forma de ver la cifra real de un archivo truncado.
	EnvMaxDiffsPerFile = "IDMLLIB_MAX_DIFFS"
)

// maxDiffsPerFile resuelve el tope, dando prioridad a la variable de entorno
// cuando trae un entero no negativo.
func maxDiffsPerFile() int {
	if v := os.Getenv(EnvMaxDiffsPerFile); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return DefaultMaxDiffsPerFile
}

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
	truncated int // archivos cuyo reporte llegó al máximo de diferencias

	// byCategory cuenta las diferencias por categoría del Req 1, criterio 2. Es el
	// marcador de progreso de las tareas de modelo: cada una tiene que bajar a cero
	// la categoría que le toca.
	byCategory map[string]int
}

func newFidelityTally(origin string) *fidelityTally {
	return &fidelityTally{origin: origin, byCategory: map[string]int{}}
}

func (f *fidelityTally) total() int {
	return f.clean + f.differing + f.unparsed + f.failed
}

func (f *fidelityTally) report(t *testing.T) {
	t.Helper()
	t.Logf("resumen [%s]: %d sin diferencias, %d con diferencias, %d copiados sin parsear, %d con fallo, %d en total",
		f.origin, f.clean, f.differing, f.unparsed, f.failed, f.total())

	if f.truncated > 0 {
		t.Logf("resumen [%s]: %d archivo(s) con el reporte truncado en el máximo de %d diferencias, así que su cifra real es mayor (%s=0 para verla)",
			f.origin, f.truncated, maxDiffsPerFile(), EnvMaxDiffsPerFile)
	}

	if len(f.byCategory) == 0 {
		return
	}
	categories := make([]string, 0, len(f.byCategory))
	for category := range f.byCategory {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		if f.byCategory[categories[i]] != f.byCategory[categories[j]] {
			return f.byCategory[categories[i]] > f.byCategory[categories[j]]
		}
		return categories[i] < categories[j]
	})
	for _, category := range categories {
		t.Logf("resumen [%s]: %6d %s", f.origin, f.byCategory[category], category)
	}
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

	case IsMasterSpreadPath(path):
		parsed, err := spread.ParseMasterSpread(data)
		if err != nil {
			return nil, true, fmt.Errorf("parseo: %w", err)
		}
		out, err = spread.MarshalMasterSpread(parsed)

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

	limit := maxDiffsPerFile()
	opts := xmlutil.DefaultCompareOptions()
	opts.MaxDifferences = limit
	diffs, err := xmlutil.CompareXMLWithDetails(data, out, opts)
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
	for _, d := range diffs {
		tally.byCategory[d.Type]++
	}

	// El comparador deja de recolectar al alcanzar el máximo, así que un archivo
	// que llega justo al tope puede tener muchas más. Sin esta marca, «100
	// diferencia(s)» se lee como «tiene cien» cuando puede tener miles.
	//
	// ponytail: un archivo con exactamente maxDiffsPerFile diferencias reales se
	// marca como truncado sin serlo. La vía de mejora es que CompareXMLWithDetails
	// devuelva la cifra real de diferencias encontradas además de la lista.
	truncated := ""
	if limit > 0 && len(diffs) >= limit {
		tally.truncated++
		truncated = " (truncado, el archivo tiene al menos esas)"
	}
	t.Logf("[%s] %s: %d diferencia(s)%s\n%s", tally.origin, path, len(diffs), truncated, xmlutil.FormatDifferences(diffs))
}

// TestGoldenRoundtrip_ExampleIDML ejecuta el ciclo parseo → serialización →
// comparación estructural sobre los dos elementos del corpus versionado, cada
// uno por su propia ruta de lectura: el Documento_Referencia como árbol de
// archivos descomprimido y el Archivo_Evidencia_Imagenes como paquete .idml.
func TestGoldenRoundtrip_ExampleIDML(t *testing.T) {
	// Los tallies se acumulan para emitir un total agregado al final. Un origen que
	// se omite por estar ausente no aporta ninguno, así que el total refleja lo que
	// de verdad se midió y no cuenta ceros de archivos que no estaban.
	var tallies []*fidelityTally

	t.Run("documento_referencia", func(t *testing.T) {
		tallies = append(tallies, testFidelityReferenceDir(t))
	})

	for _, pkg := range corpusPackages {
		t.Run(pkg.origin, func(t *testing.T) {
			tallies = append(tallies, testFidelityPackage(t, pkg))
		})
	}

	reportCorpusTotal(t, tallies)
}

// corpusPackage describe un elemento del corpus que se versiona como paquete .idml.
// Cada uno se resuelve desde su constante y admite override por variable de entorno,
// igual que el Documento_Referencia.
type corpusPackage struct {
	origin string
	def    string
	env    string
}

var corpusPackages = []corpusPackage{
	{origin: "archivo_evidencia_imagenes", def: CorpusImagesFixture, env: EnvCorpusImagesFixture},
	{origin: "plain", def: CorpusPlainIDML, env: EnvCorpusPlainIDML},
	{origin: "example", def: CorpusExampleIDML, env: EnvCorpusExampleIDML},
	{origin: "tripple", def: CorpusTrippleIDML, env: EnvCorpusTrippleIDML},
}

// reportCorpusTotal emite el total agregado por categoría sobre todo el corpus. Es
// la cifra única de avance: el desglose por origen dice dónde está el problema, este
// total dice si el problema se está reduciendo.
func reportCorpusTotal(t *testing.T, tallies []*fidelityTally) {
	t.Helper()

	if len(tallies) == 0 {
		t.Log("total del corpus: ningún origen disponible en la copia de trabajo")
		return
	}

	origins := make([]string, 0, len(tallies))
	for _, tally := range tallies {
		origins = append(origins, tally.origin)
	}

	total := aggregateTallies(tallies)
	t.Logf("total del corpus sobre %d origen(es): %s", len(origins), strings.Join(origins, ", "))
	total.report(t)
	assertCategoriasCerradas(t, total)

	if total.truncated > 0 {
		// Aviso necesario: con el tope puesto, el reparto por categoría de un archivo
		// truncado depende de en qué punto cortó, así que estas cifras no sirven para
		// comparar entre ejecuciones. La línea base se registra sin tope.
		t.Logf("total del corpus: con el tope puesto el desglose por categoría es parcial y no comparable entre ejecuciones; usar %s=0 para la cifra de referencia", EnvMaxDiffsPerFile)
	}
}

// categoriasCerradas son las categorías de diferencia que ya están en cero sobre todo
// el corpus. A diferencia del resto del arnés, que solo registra, **estas hacen fallar
// el test si vuelven a aparecer**.
//
// El motivo: llevar los atributos perdidos de 13.995 a 0 costó varias tareas, y sin
// una guarda que falle, una regresión solo se vería si alguien se fija en una cifra de
// un log. Se cierra cada categoría en cuanto llega a cero, en lugar de esperar a la
// tarea de cierre del plan para cerrarlas todas de golpe.
//
// La única que sigue abierta, y por eso no está aquí:
//   - orden-elementos-distinto: 71, pendiente de las tareas del contenedor ordenado
//
// Con `texto-distinto` cerrada, el corpus **no pierde nada de contenido**: ni un
// atributo, ni un elemento, ni un carácter. Lo único que aún no se reproduce es el
// orden de los hijos en 71 nodos.
var categoriasCerradas = []string{
	xmlutil.CategoryAttributeMissing,
	xmlutil.CategoryAttributeValue,
	xmlutil.CategoryAttributeExtra,
	xmlutil.CategoryElementMissing,
	xmlutil.CategoryElementExtra,
	xmlutil.CategoryElementOrder,
	xmlutil.CategoryTag,
	xmlutil.CategoryNamespace,
	xmlutil.CategoryText,
}

// assertCategoriasCerradas falla si una categoría ya cerrada vuelve a aparecer.
func assertCategoriasCerradas(t *testing.T, total *fidelityTally) {
	t.Helper()

	for _, categoria := range categoriasCerradas {
		if n := total.byCategory[categoria]; n > 0 {
			t.Errorf("regresión de fidelidad: la categoría %q estaba en 0 y ahora tiene %d diferencias en el corpus. "+
				"Ejecutar con %s=0 y buscar «%s» para ver dónde",
				categoria, n, EnvMaxDiffsPerFile, categoria)
		}
	}
}

// aggregateTallies suma los resultados de varios orígenes en uno.
func aggregateTallies(tallies []*fidelityTally) *fidelityTally {
	total := newFidelityTally("TOTAL")
	for _, tally := range tallies {
		total.clean += tally.clean
		total.differing += tally.differing
		total.unparsed += tally.unparsed
		total.failed += tally.failed
		total.truncated += tally.truncated
		for category, n := range tally.byCategory {
			total.byCategory[category] += n
		}
	}
	return total
}

// testFidelityReferenceDir recorre el Documento_Referencia como directorio
// descomprimido: todos los .xml de forma recursiva, con lo que mimetype queda
// fuera por extensión.
func testFidelityReferenceDir(t *testing.T) *fidelityTally {
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

	tally := newFidelityTally("documento_referencia")
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
	return tally
}

// testFidelityPackage abre un elemento del corpus versionado como paquete .idml por
// la ruta de apertura de paquete de la librería, que es lo que hace un llamador real,
// y aplica el mismo ciclo a sus .xml.
func testFidelityPackage(t *testing.T, cp corpusPackage) *fidelityTally {
	path := corpusPath(cp.env, cp.def)

	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s ausente en la copia de trabajo: %v (ruta esperada: %s)", cp.origin, err, absPathForMsg(path))
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

	tally := newFidelityTally(cp.origin)
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
	return tally
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
