package images

import (
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testPNG crea un PNG de 2x2 píxeles para las pruebas.
func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{255, 255, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestResolve_Base64(t *testing.T) {
	data := testPNG(t)
	b64 := base64.StdEncoding.EncodeToString(data)

	r := NewResolver(ResolverOptions{})
	result, err := r.Resolve(ImageSource{Base64: b64}, FrameBounds{Width: 100, Height: 100})
	if err != nil {
		t.Fatalf("Resolve falló: %v", err)
	}

	if result.StoredState != "Embedded" {
		t.Errorf("StoredState: %q, esperado Embedded", result.StoredState)
	}
	if result.Format != "PNG" {
		t.Errorf("Format: %q, esperado PNG", result.Format)
	}
	if result.WidthPx != 2 || result.HeightPx != 2 {
		t.Errorf("Dimensiones: %dx%d, esperado 2x2", result.WidthPx, result.HeightPx)
	}
	if result.LinkResourceSize == "" {
		t.Error("LinkResourceSize vacío")
	}
	if !strings.HasPrefix(result.LinkResourceSize, "0~") {
		t.Errorf("LinkResourceSize: %q, no tiene prefijo '0~'", result.LinkResourceSize)
	}
	if result.Contents == "" {
		t.Error("Contents vacío")
	}
	// Verificar que Contents tiene líneas de 76 caracteres
	lines := strings.Split(result.Contents, "\n")
	for i, line := range lines[:len(lines)-1] { // la última puede ser menor
		if len(line) != 76 {
			t.Errorf("línea %d tiene %d caracteres, esperado 76", i, len(line))
			break
		}
	}
}

func TestResolve_PathLocal(t *testing.T) {
	data := testPNG(t)
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	if err := os.WriteFile(imgPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	r := NewResolver(ResolverOptions{BaseDir: dir})
	result, err := r.Resolve(ImageSource{Path: "test.png"}, FrameBounds{Width: 50, Height: 50})
	if err != nil {
		t.Fatalf("Resolve falló: %v", err)
	}

	if !strings.HasPrefix(result.LinkResourceURI, "file:") {
		t.Errorf("LinkResourceURI: %q, no empieza con 'file:'", result.LinkResourceURI)
	}
	if result.WidthPx != 2 {
		t.Errorf("WidthPx: %d, esperado 2", result.WidthPx)
	}
}

func TestResolve_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(ResolverOptions{BaseDir: dir})

	_, err := r.Resolve(ImageSource{Path: "../etc/passwd"}, FrameBounds{})
	if err == nil {
		t.Error("esperaba error por path traversal")
	}
	if !strings.Contains(err.Error(), "fuera de BaseDir") && !strings.Contains(err.Error(), "resolver la ruta") {
		t.Errorf("error inesperado: %v", err)
	}
}

func TestResolve_SymlinkOutsideBaseDir(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	secretFile := filepath.Join(outside, "secret.png")
	os.WriteFile(secretFile, testPNG(t), 0644)

	// Crear symlink dentro de dir que apunta afuera
	linkPath := filepath.Join(dir, "link.png")
	if err := os.Symlink(secretFile, linkPath); err != nil {
		t.Skip("symlinks no soportados en este sistema")
	}

	r := NewResolver(ResolverOptions{BaseDir: dir})
	_, err := r.Resolve(ImageSource{Path: "link.png"}, FrameBounds{})
	if err == nil {
		t.Error("esperaba error por symlink fuera de BaseDir")
	}
}

func TestResolve_BaseDirNotConfigured(t *testing.T) {
	r := NewResolver(ResolverOptions{}) // sin BaseDir
	_, err := r.Resolve(ImageSource{Path: "image.png"}, FrameBounds{})
	if err == nil {
		t.Error("esperaba error por BaseDir no configurado")
	}
	if !strings.Contains(err.Error(), "BaseDir no configurado") {
		t.Errorf("error inesperado: %v", err)
	}
}

func TestResolve_AmbiguousSource(t *testing.T) {
	r := NewResolver(ResolverOptions{})
	_, err := r.Resolve(ImageSource{Path: "x.png", Base64: "abc"}, FrameBounds{})
	if err == nil {
		t.Error("esperaba error por origen ambiguo")
	}
}

func TestResolve_EmptySource(t *testing.T) {
	r := NewResolver(ResolverOptions{})
	_, err := r.Resolve(ImageSource{}, FrameBounds{})
	if err == nil {
		t.Error("esperaba error por origen vacío")
	}
}

func TestResolve_Base64ExceedsLimit(t *testing.T) {
	// Crear un base64 que excede un límite pequeño
	r := NewResolver(ResolverOptions{MaxDecodedSize: 10})
	bigData := make([]byte, 100)
	b64 := base64.StdEncoding.EncodeToString(bigData)

	_, err := r.Resolve(ImageSource{Base64: b64}, FrameBounds{})
	if err == nil {
		t.Error("esperaba error por exceder límite")
	}
	if !strings.Contains(err.Error(), "excede el límite") {
		t.Errorf("error inesperado: %v", err)
	}
}
