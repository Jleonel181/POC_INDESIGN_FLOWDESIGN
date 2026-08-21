// Package images resuelve imágenes para su inclusión en un documento IDML.
//
// Acepta dos formas de origen (exactamente una a la vez):
//   - Ruta local: un archivo en disco, restringido a BaseDir (defensa de path traversal)
//   - Carga base64: bytes ya codificados en base64 estándar
//
// No realiza ninguna petición de red.
package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // registro de decodificador JPEG
	_ "image/png"  // registro de decodificador PNG
	"math"
	"os"
	"path/filepath"
	"strings"
)

// ResolverOptions configura el resolvedor de imágenes.
type ResolverOptions struct {
	// BaseDir es el directorio base para rutas locales. Obligatorio si se usan rutas.
	BaseDir string

	// MaxDecodedSize es el tamaño máximo en bytes de una imagen decodificada desde
	// base64. Por defecto: 64 MiB.
	MaxDecodedSize int64
}

const defaultMaxDecodedSize = 64 * 1024 * 1024 // 64 MiB

// ImageSource describe el origen de una imagen. Exactamente un campo debe estar poblado.
type ImageSource struct {
	// Path es la ruta local al archivo de imagen (relativa a BaseDir o absoluta dentro de él).
	Path string

	// Base64 es la imagen ya codificada en base64 estándar (RFC 4648).
	Base64 string
}

// FrameBounds describe el tamaño del marco de destino en puntos.
type FrameBounds struct {
	Width  float64
	Height float64
}

// ResolvedImage contiene los datos de una imagen resuelta, lista para insertar en el IDML.
type ResolvedImage struct {
	// Contents es la carga base64 con líneas de 76 caracteres, sin salto inicial ni final.
	Contents string

	// LinkResourceURI es la URI file: del origen.
	LinkResourceURI string

	// LinkResourceSize con la forma "0~<hex>" (bytes decodificados en hexadecimal).
	LinkResourceSize string

	// StoredState es "Embedded" para el modo embebido, "Normal" para enlazado.
	StoredState string

	// Format es el formato detectado: "JPEG", "PNG", etc.
	Format string

	// WidthPx y HeightPx son las dimensiones en píxeles.
	WidthPx  int
	HeightPx int

	// ActualPpi es la resolución real de la imagen (basada en dimensiones).
	ActualPpi string

	// EffectivePpi es la resolución efectiva dentro del marco.
	EffectivePpi string
}

// Resolver resuelve imágenes para inclusión en IDML.
type Resolver struct {
	opts ResolverOptions
}

// NewResolver crea un resolvedor con las opciones dadas.
func NewResolver(opts ResolverOptions) *Resolver {
	if opts.MaxDecodedSize <= 0 {
		opts.MaxDecodedSize = defaultMaxDecodedSize
	}
	return &Resolver{opts: opts}
}

// Resolve resuelve una imagen para su inclusión en el IDML.
// Error si: ambos campos de origen poblados, ninguno, ruta fuera de BaseDir, bytes vacíos, etc.
func (r *Resolver) Resolve(source ImageSource, bounds FrameBounds) (*ResolvedImage, error) {
	// Validar que exactamente un origen esté presente.
	hasPath := source.Path != ""
	hasBase64 := source.Base64 != ""

	if hasPath && hasBase64 {
		return nil, fmt.Errorf("imagen: origen ambiguo, declarados tanto path como base64")
	}
	if !hasPath && !hasBase64 {
		return nil, fmt.Errorf("imagen: origen vacío, se requiere path o base64")
	}

	var rawBytes []byte
	var fileURI string
	var err error

	if hasPath {
		rawBytes, fileURI, err = r.resolveFromPath(source.Path)
	} else {
		rawBytes, fileURI, err = r.resolveFromBase64(source.Base64)
	}
	if err != nil {
		return nil, err
	}

	if len(rawBytes) == 0 {
		return nil, fmt.Errorf("imagen %q: los bytes están vacíos", source.Path+source.Base64[:min(20, len(source.Base64))])
	}

	// Obtener dimensiones y formato.
	cfg, format, err := decodeImageConfig(rawBytes)
	if err != nil {
		return nil, fmt.Errorf("imagen: no se pudo decodificar como imagen: %w", err)
	}

	// Codificar a base64 con líneas de 76 caracteres.
	contents := EncodeBase64Lines(rawBytes)

	// Calcular LinkResourceSize: "0~<hex>"
	linkResourceSize := fmt.Sprintf("0~%x", len(rawBytes))

	// Calcular PPI.
	actualPpi := fmt.Sprintf("%d %d", 72, 72) // por defecto, las imágenes IDML asumen 72
	effectivePpi := actualPpi
	if bounds.Width > 0 && bounds.Height > 0 && cfg.Width > 0 && cfg.Height > 0 {
		hPpi := float64(cfg.Width) * 72.0 / bounds.Width
		vPpi := float64(cfg.Height) * 72.0 / bounds.Height
		effectivePpi = fmt.Sprintf("%d %d", int(math.Round(hPpi)), int(math.Round(vPpi)))
	}

	return &ResolvedImage{
		Contents:         contents,
		LinkResourceURI:  fileURI,
		LinkResourceSize: linkResourceSize,
		StoredState:      "Embedded",
		Format:           strings.ToUpper(format),
		WidthPx:          cfg.Width,
		HeightPx:         cfg.Height,
		ActualPpi:        actualPpi,
		EffectivePpi:     effectivePpi,
	}, nil
}

// resolveFromPath lee una imagen de una ruta local, verificando que esté dentro de BaseDir.
func (r *Resolver) resolveFromPath(path string) ([]byte, string, error) {
	if r.opts.BaseDir == "" {
		return nil, "", fmt.Errorf("imagen %q: BaseDir no configurado (obligatorio para rutas locales)", path)
	}

	// Resolver BaseDir como absoluta.
	baseDir, err := filepath.Abs(r.opts.BaseDir)
	if err != nil {
		return nil, "", fmt.Errorf("imagen %q: error resolviendo BaseDir: %w", path, err)
	}

	// Resolver la ruta del archivo.
	var absPath string
	if filepath.IsAbs(path) {
		absPath = path
	} else {
		absPath = filepath.Join(baseDir, path)
	}

	// EvalSymlinks para resolver symlinks y detectar path traversal.
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("imagen %q: no se pudo resolver la ruta: %w", path, err)
	}

	realBase, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return nil, "", fmt.Errorf("imagen %q: no se pudo resolver BaseDir: %w", path, err)
	}

	// Verificar que la ruta resuelta está dentro de BaseDir.
	if !strings.HasPrefix(realPath, realBase+string(filepath.Separator)) && realPath != realBase {
		return nil, "", fmt.Errorf("imagen %q: la ruta resuelta %q está fuera de BaseDir %q", path, realPath, realBase)
	}

	data, err := os.ReadFile(realPath)
	if err != nil {
		return nil, "", fmt.Errorf("imagen %q: error de lectura: %w", path, err)
	}

	// URI file: con la ruta absoluta.
	fileURI := "file:" + realPath

	return data, fileURI, nil
}

// resolveFromBase64 decodifica una imagen desde base64.
func (r *Resolver) resolveFromBase64(b64 string) ([]byte, string, error) {
	// Estimar tamaño decodificado (base64 expande ~33%).
	estimatedSize := int64(len(b64)) * 3 / 4
	if estimatedSize > r.opts.MaxDecodedSize {
		return nil, "", fmt.Errorf("imagen base64: tamaño estimado %d excede el límite de %d bytes",
			estimatedSize, r.opts.MaxDecodedSize)
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, "", fmt.Errorf("imagen base64: error de decodificación: %w", err)
	}

	if int64(len(data)) > r.opts.MaxDecodedSize {
		return nil, "", fmt.Errorf("imagen base64: tamaño decodificado %d excede el límite de %d bytes",
			len(data), r.opts.MaxDecodedSize)
	}

	// Para base64 generamos una URI file: genérica (el archivo no existe en disco).
	fileURI := "file:embedded_image"

	return data, fileURI, nil
}

// decodeImageConfig obtiene las dimensiones y formato de una imagen sin decodificarla completa.
func decodeImageConfig(data []byte) (image.Config, string, error) {
	r := bytes.NewReader(data)
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		return image.Config{}, "", err
	}
	return cfg, format, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
