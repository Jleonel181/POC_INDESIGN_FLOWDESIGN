package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
)

// GoldenFile provee utilidades para testing con golden files.
type GoldenFile struct {
	g *goldie.Goldie
}

// NewGoldenFile crea un nuevo tester de GoldenFile.
// El parámetro dir especifica dónde se almacenan los golden files.
func NewGoldenFile(t *testing.T, dir string) *GoldenFile {
	t.Helper()

	return &GoldenFile{
		g: goldie.New(t,
			goldie.WithFixtureDir(dir),
			goldie.WithNameSuffix(".golden"),
		),
	}
}

// NewGoldenFileInTestdata crea un tester GoldenFile usando el directorio testdata/golden.
// Es un método de conveniencia para el caso más común.
func NewGoldenFileInTestdata(t *testing.T) *GoldenFile {
	t.Helper()
	return NewGoldenFile(t, filepath.Join("testdata", "golden"))
}

// Assert compara los datos actuales contra el golden file.
// Si difieren, el test falla con un diff detallado.
func (gf *GoldenFile) Assert(t *testing.T, name string, actual []byte) {
	t.Helper()
	gf.g.Assert(t, name, actual)
}

// AssertFile compara el contenido de un archivo contra el golden file.
func (gf *GoldenFile) AssertFile(t *testing.T, name, filePath string) {
	t.Helper()

	// #nosec G304 - Función utilitaria de test con rutas de datos de test controladas
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	gf.g.Assert(t, name, data)
}

// Update es un wrapper de conveniencia sobre Assert que hace clara la intención.
// Los golden files se actualizan mediante el flag -update al correr los tests.
//
// Uso:
//
//	go test ./pkg/idml/... -run TestGolden -update
//
// Este método existe principalmente para documentación y claridad en el código de test.
func (gf *GoldenFile) Update(t *testing.T, name string, actual []byte) {
	t.Helper()

	// Nota: goldie maneja las actualizaciones automáticamente cuando se usa el flag -update
	// Esto solo llama a Assert, que compara o actualiza según el flag
	gf.g.Assert(t, name, actual)

	t.Logf("✅ Golden file checked/updated: %s", name)
}
