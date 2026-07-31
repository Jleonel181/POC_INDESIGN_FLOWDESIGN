package idml

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/beevik/etree"
)

// referenceMinimalIDML es una exportación de InDesign de la misma forma que produce
// NewFromTemplate: 1 página, 1 marco de texto. Es el oráculo del conjunto de rutas.
const referenceMinimalIDML = "../../testdata/plain.idml"

// zipEntryNames devuelve los nombres de las entradas de un paquete, en su orden.
func zipEntryNames(t *testing.T, path string) []string {
	t.Helper()

	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("no se pudo abrir %s como ZIP: %v", path, err)
	}
	defer r.Close()

	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}

// writeTemplatePackage genera un paquete y lo escribe en un archivo temporal,
// pasando por la misma ruta de escritura que usaría un llamador real.
func writeTemplatePackage(t *testing.T, opts *TemplateOptions) string {
	t.Helper()

	pkg, err := NewFromTemplate(opts)
	if err != nil {
		t.Fatalf("NewFromTemplate falló: %v", err)
	}

	path := filepath.Join(t.TempDir(), "generado.idml")
	if err := Write(pkg, path); err != nil {
		t.Fatalf("Write falló: %v", err)
	}
	return path
}

// TestNewFromTemplate_ConjuntoDeRutasCoincideConInDesign es el criterio 5 de la
// Tarea 3: el paquete generado tiene las mismas entradas que una exportación real de
// InDesign de la misma forma. Antes le faltaban cuatro.
func TestNewFromTemplate_ConjuntoDeRutasCoincideConInDesign(t *testing.T) {
	if _, err := os.Stat(referenceMinimalIDML); err != nil {
		t.Skipf("%s ausente: %v", referenceMinimalIDML, err)
	}

	want := zipEntryNames(t, referenceMinimalIDML)
	got := zipEntryNames(t, writeTemplatePackage(t, nil))

	if len(want) != 13 {
		t.Errorf("%s debería tener 13 entradas y tiene %d: ¿cambió el fixture?", referenceMinimalIDML, len(want))
	}

	// El orden importa solo para mimetype, que tiene que ir primero. El resto se
	// compara como conjunto.
	if len(got) == 0 || got[0] != PathMimetype {
		t.Errorf("la primera entrada del ZIP debe ser %q, es %q", PathMimetype, got[0])
	}

	// Los identificadores de los nombres de archivo son arbitrarios: InDesign llamó a
	// su master spread MasterSpread_ubb.xml y la plantilla lo llama ub4. Comparar los
	// nombres literales fijaría un dato sin significado, así que se compara la forma:
	// mismo directorio y mismo tipo de archivo.
	sortedWant := shapeOfPaths(want)
	sortedGot := shapeOfPaths(got)
	if strings.Join(sortedGot, "\n") != strings.Join(sortedWant, "\n") {
		t.Errorf("el conjunto de rutas no coincide\nInDesign: %v\ngenerado: %v", sortedWant, sortedGot)
	}
}

// shapeOfPaths sustituye el identificador de los nombres que lo llevan por un
// comodín, y ordena el resultado.
func shapeOfPaths(paths []string) []string {
	shape := make([]string, 0, len(paths))
	for _, p := range paths {
		dir, file := filepath.Split(p)
		for _, prefix := range []string{"MasterSpread_", "Spread_", "Story_"} {
			if strings.HasPrefix(file, prefix) {
				file = prefix + "*" + ExtXML
				break
			}
		}
		shape = append(shape, dir+file)
	}
	sort.Strings(shape)
	return shape
}

// TestNewFromTemplate_DeclaraLasTresReferencias es el criterio 2: el designmap
// declara las referencias que antes no tenía. Sin idPkg:Spread el documento no tiene
// ni una página.
func TestNewFromTemplate_DeclaraLasTresReferencias(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatalf("NewFromTemplate falló: %v", err)
	}

	designmap := string(pkg.files[PathDesignmap].data)
	for _, ref := range []string{"idPkg:Spread", "idPkg:Story", "idPkg:BackingStory"} {
		if !strings.Contains(designmap, "<"+ref+" ") {
			t.Errorf("el designmap no declara %s", ref)
		}
	}
}

// TestNewFromTemplate_CierreReferencial es el criterio 3, y es el que de verdad
// dice si el documento es coherente. Cada referencia se resuelve contra el elemento
// al que apunta, en el archivo donde ese elemento vive.
func TestNewFromTemplate_CierreReferencial(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatalf("NewFromTemplate falló: %v", err)
	}

	parse := func(path string) *etree.Element {
		t.Helper()
		entry, ok := pkg.files[path]
		if !ok {
			t.Fatalf("el paquete no contiene %s", path)
		}
		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(entry.data); err != nil {
			t.Fatalf("%s no es XML válido: %v", path, err)
		}
		return doc.Root()
	}

	designmap := parse(PathDesignmap)
	spread := parse(PathSpread).FindElement("Spread")
	story := parse(PathStory).FindElement("Story")
	backing := parse(PathBackingStory).FindElement("XmlStory")

	if spread == nil || story == nil || backing == nil {
		t.Fatal("falta el elemento interno de spread, story o backing story")
	}

	textFrame := spread.FindElement("TextFrame")
	page := spread.FindElement("Page")
	if textFrame == nil || page == nil {
		t.Fatal("el spread debe llevar una Page y un TextFrame")
	}

	// El ItemLayer del marco apunta a una Layer declarada en el designmap.
	itemLayer := textFrame.SelectAttrValue("ItemLayer", "")
	if findBySelf(designmap, "Layer", itemLayer) == nil {
		t.Errorf("ItemLayer=%q del marco no corresponde a ninguna Layer del designmap", itemLayer)
	}

	// El ParentStory del marco apunta a la story emitida.
	parentStory := textFrame.SelectAttrValue("ParentStory", "")
	if got := story.SelectAttrValue("Self", ""); got != parentStory {
		t.Errorf("ParentStory=%q del marco no coincide con la story emitida, Self=%q", parentStory, got)
	}

	// El PageStart del Section apunta a la página emitida.
	section := designmap.FindElement("Section")
	if section == nil {
		t.Fatal("el designmap debe llevar un Section")
	}
	pageStart := section.SelectAttrValue("PageStart", "")
	if got := page.SelectAttrValue("Self", ""); got != pageStart {
		t.Errorf("PageStart=%q del Section no coincide con la página emitida, Self=%q", pageStart, got)
	}

	// El AppliedMaster de la página apunta al master spread emitido.
	appliedMaster := page.SelectAttrValue("AppliedMaster", "")
	master := parse(PathMasterSpread).FindElement("MasterSpread")
	if master == nil {
		t.Fatal("falta el elemento MasterSpread")
	}
	if got := master.SelectAttrValue("Self", ""); got != appliedMaster {
		t.Errorf("AppliedMaster=%q de la página no coincide con el master spread, Self=%q", appliedMaster, got)
	}

	// El StoryList del Document contiene la story del marco y la backing story.
	storyList := strings.Fields(designmap.SelectAttrValue("StoryList", ""))
	for _, want := range []string{story.SelectAttrValue("Self", ""), backing.SelectAttrValue("Self", "")} {
		if !contains(storyList, want) {
			t.Errorf("StoryList=%v no contiene %q", storyList, want)
		}
	}
}

// findBySelf busca entre los hijos directos un elemento con esa etiqueta y ese Self.
func findBySelf(parent *etree.Element, tag, self string) *etree.Element {
	if self == "" {
		return nil
	}
	for _, child := range parent.ChildElements() {
		if child.Tag == tag && child.SelectAttrValue("Self", "") == self {
			return child
		}
	}
	return nil
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

// TestNewFromTemplate_DocumentPreferenceSoloEnPreferencias es el criterio 4.
func TestNewFromTemplate_DocumentPreferenceSoloEnPreferencias(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatalf("NewFromTemplate falló: %v", err)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(pkg.files[PathPreferences].data); err != nil {
		t.Fatalf("%s no es XML válido: %v", PathPreferences, err)
	}

	for _, tag := range []string{"DocumentPreference", "MarginPreference", "ViewPreference"} {
		n := 0
		for _, child := range doc.Root().ChildElements() {
			if child.Tag == tag {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%s debe contener exactamente un %s, contiene %d", PathPreferences, tag, n)
		}
	}

	// El MarginPreference de Resources/Preferences.xml va sin ColumnsPositions; el de
	// los spreads sí lo lleva (Req 8, criterio 9).
	prefsMargin := doc.Root().FindElement("MarginPreference")
	if prefsMargin.SelectAttr("ColumnsPositions") != nil {
		t.Error("el MarginPreference de Resources/Preferences.xml no debe llevar ColumnsPositions")
	}
}

// TestNewFromTemplate_ColumnasSeCalculan cubre el criterio 8: las dos variantes de
// columnas se generan y el ColumnsPositions emitido corresponde a la geometría.
//
// Con 1 columna el valor coincide con lo que emite InDesign en plain.idml para un A4
// con márgenes de 36: de 0 al ancho utilizable.
func TestNewFromTemplate_ColumnasSeCalculan(t *testing.T) {
	tests := []struct {
		name    string
		columns int
		want    string
	}{
		{name: "1 columna", columns: 1, want: "0 523.276"},
		{name: "5 columnas", columns: 5, want: "0 95.0552 107.0552 202.1104 214.1104 309.1656 321.1656 416.2208 428.2208 523.276"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultTemplateOptions()
			opts.Preset = PresetA4
			opts.ColumnCount = tt.columns

			pkg, err := NewFromTemplate(opts)
			if err != nil {
				t.Fatalf("NewFromTemplate falló: %v", err)
			}

			doc := etree.NewDocument()
			if err := doc.ReadFromBytes(pkg.files[PathSpread].data); err != nil {
				t.Fatalf("%s no es XML válido: %v", PathSpread, err)
			}
			margin := doc.Root().FindElement("Spread/Page/MarginPreference")
			if margin == nil {
				t.Fatal("la página del spread debe llevar MarginPreference")
			}

			if got := margin.SelectAttrValue("ColumnsPositions", ""); got != tt.want {
				t.Errorf("ColumnsPositions = %q, esperado %q", got, tt.want)
			}
			if got := margin.SelectAttrValue("ColumnCount", ""); got != strconv.Itoa(tt.columns) {
				t.Errorf("ColumnCount = %q, esperado %d", got, tt.columns)
			}
		})
	}
}

// TestNewFromTemplate_MarginesInvalidos comprueba que una geometría imposible se
// rechaza con un error que dice qué medidas la hacen imposible, en lugar de emitir un
// documento con un marco de tamaño negativo.
func TestNewFromTemplate_MarginesInvalidos(t *testing.T) {
	opts := DefaultTemplateOptions()
	opts.Preset = PresetLetterUS
	opts.Margins.Left = 400
	opts.Margins.Right = 400

	if _, err := NewFromTemplate(opts); err == nil {
		t.Error("márgenes que no dejan área utilizable deben producir error")
	} else if !strings.Contains(err.Error(), "márgenes") {
		t.Errorf("el error debe explicar el problema de los márgenes, dice: %v", err)
	}
}

// TestNewFromTemplate_ColumnasQueNoCaben cubre el otro límite de la geometría.
func TestNewFromTemplate_ColumnasQueNoCaben(t *testing.T) {
	opts := DefaultTemplateOptions()
	opts.Preset = PresetLetterUS
	opts.ColumnCount = 100
	opts.ColumnGutter = 12

	if _, err := NewFromTemplate(opts); err == nil {
		t.Error("100 columnas con medianil 12 no caben en US Letter y deben producir error")
	} else if !strings.Contains(err.Error(), "columnas") {
		t.Errorf("el error debe mencionar las columnas, dice: %v", err)
	}
}

// EnvProbeDir es el directorio donde escribir las sondas para abrirlas a mano en
// Adobe InDesign. Sin esta variable el test se omite: emite archivos para inspección
// humana, no comprueba nada por sí mismo.
const EnvProbeDir = "IDMLLIB_PROBE_DIR"

// TestNewFromTemplate_SondaParaInDesign emite las dos variantes que pide el criterio
// 8 de la Tarea 3, una de 1 columna y otra de 5, para la verificación manual que
// cierra los riesgos 2, 4 y 6 de requirements.md.
//
// Uso:
//
//	IDMLLIB_PROBE_DIR=/tmp/sondas go test ./pkg/idml/ -run SondaParaInDesign -v -count=1
func TestNewFromTemplate_SondaParaInDesign(t *testing.T) {
	dir := os.Getenv(EnvProbeDir)
	if dir == "" {
		t.Skipf("sin %s: este test emite archivos para abrir en InDesign, no comprueba nada", EnvProbeDir)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("no se pudo crear %s: %v", dir, err)
	}

	for _, columns := range []int{1, 5} {
		opts := DefaultTemplateOptions()
		opts.Preset = PresetA4
		opts.ColumnCount = columns

		pkg, err := NewFromTemplate(opts)
		if err != nil {
			t.Fatalf("%d columna(s): NewFromTemplate falló: %v", columns, err)
		}

		path := filepath.Join(dir, "sonda_a4_"+strconv.Itoa(columns)+"col.idml")
		if err := Write(pkg, path); err != nil {
			t.Fatalf("%d columna(s): Write falló: %v", columns, err)
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(pkg.files[PathSpread].data); err != nil {
			t.Fatalf("%d columna(s): el spread no es XML válido: %v", columns, err)
		}
		positions := doc.Root().FindElement("Spread/Page/MarginPreference").SelectAttrValue("ColumnsPositions", "")

		t.Logf("escrito %s", path)
		t.Logf("   ColumnsPositions emitido: %s", positions)
		t.Logf("   comprobar a mano: InDesign lo abre sin avisar de daño ni pedir recuperación,")
		t.Logf("   y si respeta ese ColumnsPositions o lo recalcula al abrir (riesgo 6)")
	}
}

// TestNewFromTemplate_TodosLosArchivosParsean comprueba que cada XML del paquete es
// XML bien formado. Es la red mínima: una plantilla con una llave mal cerrada
// produciría un archivo roto que el resto de los tests no mira.
func TestNewFromTemplate_TodosLosArchivosParsean(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatalf("NewFromTemplate falló: %v", err)
	}

	for _, path := range pkg.Files() {
		if !strings.EqualFold(filepath.Ext(path), ExtXML) {
			continue
		}
		entry := pkg.files[path]
		if err := xml.Unmarshal(entry.data, new(struct {
			XMLName xml.Name
			Inner   []byte `xml:",innerxml"`
		})); err != nil {
			t.Errorf("%s no es XML bien formado: %v", path, err)
		}
	}
}
