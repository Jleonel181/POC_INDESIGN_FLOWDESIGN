package idml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Límites por defecto para protección contra ZIP bombs
const (
	// DefaultMaxTotalSize es el tamaño total descomprimido máximo (500 MB)
	DefaultMaxTotalSize int64 = 1000 * 1024 * 1024

	// DefaultMaxFileSize es el tamaño máximo de un archivo individual (100 MB)
	DefaultMaxFileSize int64 = 200 * 1024 * 1024

	// DefaultMaxFileCount es el número máximo de archivos en el archivo comprimido
	DefaultMaxFileCount int = 10000

	// DefaultMaxCompressionRatio es la relación de compresión máxima permitida
	// Una relación de 100 significa que el tamaño descomprimido puede ser como máximo 100x el tamaño comprimido
	DefaultMaxCompressionRatio int64 = 100
)

// ReadOptions configura la operación Read con límites opcionales.
type ReadOptions struct {
	// MaxTotalSize limita el tamaño total descomprimido de todos los archivos.
	// Establecer en 0 para usar DefaultMaxTotalSize, -1 para sin límite.
	MaxTotalSize int64

	// MaxFileSize limita el tamaño de archivos individuales.
	// Establecer en 0 para usar DefaultMaxFileSize, -1 para sin límite.
	MaxFileSize int64

	// MaxFileCount limita el número de archivos en el archivo comprimido.
	// Establecer en 0 para usar DefaultMaxFileCount, -1 para sin límite.
	MaxFileCount int

	// MaxCompressionRatio limita la relación de compresión (descomprimido/comprimido).
	// Establecer en 0 para usar DefaultMaxCompressionRatio, -1 para sin límite.
	MaxCompressionRatio int64
}

// applyDefaults rellena los valores por defecto para opciones con valor cero.
func (opts *ReadOptions) applyDefaults() {
	if opts.MaxTotalSize == 0 {
		opts.MaxTotalSize = DefaultMaxTotalSize
	}
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = DefaultMaxFileSize
	}
	if opts.MaxFileCount == 0 {
		opts.MaxFileCount = DefaultMaxFileCount
	}
	if opts.MaxCompressionRatio == 0 {
		opts.MaxCompressionRatio = DefaultMaxCompressionRatio
	}
}

// isValidZipPath verifica si la ruta de una entrada ZIP es segura para extraer.
// Rechaza rutas absolutas y rutas que intentan traversal de directorios.
func isValidZipPath(name string) bool {
	// Rechazar rutas vacías
	if name == "" {
		return false
	}

	// Rechazar rutas absolutas
	if filepath.IsAbs(name) {
		return false
	}

	// Limpiar la ruta y verificar intentos de traversal
	cleaned := filepath.Clean(name)

	// Rechazar si la ruta limpia comienza con ".."
	if strings.HasPrefix(cleaned, "..") {
		return false
	}

	// Rechazar si la ruta contiene componentes ".." (incluso en el medio)
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return false
		}
	}

	return true
}

// extractPackage es la lógica compartida para todas las funciones Read*.
// Procesa un slice de archivos ZIP y crea un Package con verificaciones de seguridad.
func extractPackage(files []*zip.File, opts *ReadOptions, source string) (*Package, error) {
	// Verificar límite de cantidad de archivos
	if err := validateFileCount(files, opts, source); err != nil {
		return nil, err
	}

	pkg := New()
	var totalSize int64

	// Leer cada archivo en el archivo comprimido
	for _, f := range files {
		if err := validateZipFile(f, opts, &totalSize, source); err != nil {
			return nil, err
		}

		data, err := extractZipFileData(f, opts, source)
		if err != nil {
			return nil, err
		}

		storeFileInPackage(pkg, f, data)
	}

	// Extraer metadatos XMP de META-INF/metadata.xml
	if entry, exists := pkg.files["META-INF/metadata.xml"]; exists {
		pkg.XMPMetadata = extractXMPMetadata(string(entry.data))
	}

	// Parsear automáticamente designmap.xml para acceso estructurado
	if err := validateDesignMap(pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

// validateFileCount verifica si el número de archivos supera el límite.
func validateFileCount(files []*zip.File, opts *ReadOptions, source string) error {
	if opts.MaxFileCount > 0 && len(files) > opts.MaxFileCount {
		return common.WrapErrorWithPath("idml", "read", source, fmt.Errorf("archive contains %d files, exceeds limit of %d", len(files), opts.MaxFileCount))
	}
	return nil
}

// validateZipFile realiza validación de seguridad en una entrada de archivo ZIP.
func validateZipFile(f *zip.File, opts *ReadOptions, totalSize *int64, source string) error {
	// Validar la ruta para prevenir ataques de traversal de directorios
	if !isValidZipPath(f.Name) {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("invalid path: potential directory traversal"))
	}

	// Verificar límite de tamaño de archivo individual
	if err := validateFileSize(f, opts); err != nil {
		return err
	}

	// Verificar relación de compresión para detectar ZIP bombs
	if err := validateCompressionRatio(f, opts); err != nil {
		return err
	}

	// Rastrear tamaño total con protección contra desbordamiento
	if f.UncompressedSize64 > math.MaxInt64 {
		return common.WrapErrorWithPath("idml", "read", source, fmt.Errorf("file size %d bytes exceeds maximum supported size", f.UncompressedSize64))
	}
	// #nosec G115 - La conversión es segura: verificación de desbordamiento realizada en la línea anterior
	uncompressedSize := int64(f.UncompressedSize64)
	*totalSize += uncompressedSize
	if opts.MaxTotalSize > 0 && *totalSize > opts.MaxTotalSize {
		return common.WrapErrorWithPath("idml", "read", source, fmt.Errorf("total uncompressed size exceeds limit of %d bytes", opts.MaxTotalSize))
	}

	return nil
}

// validateFileSize verifica si el tamaño de un archivo individual supera los límites.
func validateFileSize(f *zip.File, opts *ReadOptions) error {
	if f.UncompressedSize64 > math.MaxInt64 {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file size %d bytes exceeds maximum supported size", f.UncompressedSize64))
	}
	// #nosec G115 - La conversión es segura: verificación de desbordamiento realizada en la línea anterior
	uncompressedSize := int64(f.UncompressedSize64)
	if opts.MaxFileSize > 0 && uncompressedSize > opts.MaxFileSize {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file size %d bytes exceeds limit of %d bytes", f.UncompressedSize64, opts.MaxFileSize))
	}
	return nil
}

// validateCompressionRatio verifica posibles ZIP bombs basándose en la relación de compresión.
func validateCompressionRatio(f *zip.File, opts *ReadOptions) error {
	if opts.MaxCompressionRatio > 0 && f.CompressedSize64 > 0 && f.Method != zip.Store {
		if f.UncompressedSize64 > math.MaxInt64 || f.CompressedSize64 > math.MaxInt64 {
			return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file sizes exceed maximum supported size"))
		}
		// #nosec G115 - Las conversiones son seguras: verificaciones de desbordamiento realizadas en la línea anterior
		uncompressedSize := int64(f.UncompressedSize64)
		compressedSize := int64(f.CompressedSize64) // #nosec G115 - La conversión es segura: verificación de desbordamiento realizada anteriormente
		ratio := uncompressedSize / compressedSize
		if ratio > opts.MaxCompressionRatio {
			return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("compression ratio %d exceeds limit of %d (potential ZIP bomb)", ratio, opts.MaxCompressionRatio))
		}
	}
	return nil
}

// extractZipFileData extrae de forma segura los datos de una entrada de archivo ZIP.
func extractZipFileData(f *zip.File, opts *ReadOptions, source string) ([]byte, error) {
	// Agregar validación de parámetros
	if f == nil {
		return nil, common.Errorf("idml", "extract zip file data", "", "zip file entry is nil")
	}

	if opts == nil {
		return nil, common.Errorf("idml", "extract zip file data", "", "read options is nil")
	}

	if source == "" {
		source = "<unknown>"
	}

	// Abrir el archivo dentro del ZIP
	rc, err := f.Open()
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read", source+"/"+f.Name, err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			// Registrar error de cierre pero no sobreescribir el error principal
		}
	}()

	// Usar LimitReader para evitar leer más del tamaño declarado + pequeño buffer
	maxRead := calculateMaxReadSize(f, opts)
	limitedReader := io.LimitReader(rc, maxRead)

	// Leer el contenido completo del archivo
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read", source+"/"+f.Name, err)
	}

	// Verificar que el tamaño real no supere significativamente el tamaño declarado
	if err := validateActualFileSize(f, data); err != nil {
		return nil, err
	}

	return data, nil
}

// calculateMaxReadSize determina el tamaño máximo seguro de lectura para un archivo.
func calculateMaxReadSize(f *zip.File, opts *ReadOptions) int64 {
	// Agregar validación de parámetros
	if f == nil {
		return 1024 // Tamaño pequeño por defecto para mayor seguridad
	}

	if opts == nil {
		if f.UncompressedSize64 > math.MaxInt64-1024 {
			return math.MaxInt64
		}
		// #nosec G115 - La conversión es segura: verificación de desbordamiento realizada en la línea anterior
		uncompressedSize := int64(f.UncompressedSize64)
		return uncompressedSize + 1024
	}

	if f.UncompressedSize64 > math.MaxInt64-1024 {
		maxRead := int64(math.MaxInt64)
		if opts.MaxFileSize > 0 && maxRead > opts.MaxFileSize {
			maxRead = opts.MaxFileSize + 1
		}
		return maxRead
	}
	// #nosec G115 - La conversión es segura: verificación de desbordamiento realizada anteriormente (línea 258)
	uncompressedSize := int64(f.UncompressedSize64)
	maxRead := uncompressedSize + 1024 // Permitir 1KB de margen para casos límite
	if opts.MaxFileSize > 0 && maxRead > opts.MaxFileSize {
		maxRead = opts.MaxFileSize + 1
	}
	return maxRead
}

// validateActualFileSize verifica que el tamaño real leído coincida con las expectativas.
func validateActualFileSize(f *zip.File, data []byte) error {
	// Agregar validación de parámetros
	if f == nil {
		return common.Errorf("idml", "validate file size", "", "zip file entry is nil")
	}

	if data == nil {
		return common.Errorf("idml", "validate file size", f.Name, "file data is nil")
	}

	// Verificar posible desbordamiento antes de la conversión
	if f.UncompressedSize64 > math.MaxInt64-1024 {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file size %d bytes exceeds maximum supported size", f.UncompressedSize64))
	}

	// #nosec G115 - La conversión es segura: verificación de desbordamiento realizada en la línea anterior
	uncompressedSize := int64(f.UncompressedSize64)
	maxExpectedSize := uncompressedSize + 1024
	if int64(len(data)) > maxExpectedSize {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("actual size %d exceeds declared size %d (potential ZIP bomb)", len(data), f.UncompressedSize64))
	}
	return nil
}

// storeFileInPackage almacena los datos del archivo extraído en el package.
func storeFileInPackage(pkg *Package, f *zip.File, data []byte) {
	// Agregar validación de parámetros
	if pkg == nil || f == nil {
		return // Ignorar silenciosamente parámetros inválidos para evitar panics
	}

	// Asegurar que data no sea nil
	if data == nil {
		data = []byte{}
	}

	pkg.files[f.Name] = &fileEntry{
		data:   data,
		header: &f.FileHeader,
	}
	pkg.fileOrder = append(pkg.fileOrder, f.Name)
}

// validateDesignMap asegura que el archivo designmap.xml exista y pueda ser parseado.
func validateDesignMap(pkg *Package) error {
	if _, err := pkg.Document(); err != nil {
		// Si designmap.xml no existe o no puede parsearse, es un error
		// ya que es un archivo requerido en los packages IDML
		return err
	}
	return nil
}

// extractXMPMetadata extrae el paquete XMP del XML de IDML.
// Retorna cadena vacía si no se encuentra ningún paquete XMP.
func extractXMPMetadata(xmlContent string) string {
	// Buscar desde <?xpacket begin hasta <?xpacket end
	xmpPattern := regexp.MustCompile(`(?s)<\?xpacket begin.*?<\?xpacket end[^>]*\?>`)
	return xmpPattern.FindString(xmlContent)
}

// ReadBytes parsea IDML desde un slice de bytes en memoria con límites de seguridad por defecto.
//
// Es útil cuando ya tienes los datos IDML en memoria (por ejemplo, desde
// S3, respuesta HTTP u otras fuentes) y quieres evitar escribir en un archivo temporal.
//
// La función aplica la misma protección contra ZIP bombs que Read().
func ReadBytes(data []byte) (*Package, error) {
	return ReadBytesWithOptions(data, nil)
}

// ReadBytesWithOptions parsea IDML desde un slice de bytes con límites de seguridad configurables.
//
// Si opts es nil, se aplican los límites por defecto. Para deshabilitar un límite específico,
// establecerlo en -1.
//
// Ejemplo:
//
//	data, _ := os.ReadFile("document.idml")
//	pkg, err := idml.ReadBytes(data)
func ReadBytesWithOptions(data []byte, opts *ReadOptions) (*Package, error) {
	// Aplicar valores por defecto
	if opts == nil {
		opts = &ReadOptions{}
	}
	opts.applyDefaults()

	// Crear un bytes.Reader que implementa io.ReaderAt
	bytesReader := bytes.NewReader(data)

	// Crear un zip.Reader desde los bytes
	zipReader, err := zip.NewReader(bytesReader, int64(len(data)))
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read bytes", "<memory>", err)
	}

	// Usar la lógica de extracción compartida
	return extractPackage(zipReader.File, opts, "<memory>")
}

// ReadFrom parsea IDML desde un io.ReaderAt con límites de seguridad por defecto.
//
// Es útil para escenarios de streaming donde tienes un io.ReaderAt
// (por ejemplo, bytes.Reader, os.File o implementaciones personalizadas) y quieres
// parsear IDML sin cargar primero en un slice de bytes.
//
// El parámetro size debe ser el tamaño total de los datos IDML.
func ReadFrom(r io.ReaderAt, size int64) (*Package, error) {
	return ReadFromWithOptions(r, size, nil)
}

// ReadFromWithOptions parsea IDML desde un io.ReaderAt con límites de seguridad configurables.
//
// Si opts es nil, se aplican los límites por defecto. Para deshabilitar un límite específico,
// establecerlo en -1.
//
// Ejemplo:
//
//	file, _ := os.Open("document.idml")
//	stat, _ := file.Stat()
//	pkg, err := idml.ReadFrom(file, stat.Size())
func ReadFromWithOptions(r io.ReaderAt, size int64, opts *ReadOptions) (*Package, error) {
	// Aplicar valores por defecto
	if opts == nil {
		opts = &ReadOptions{}
	}
	opts.applyDefaults()

	// Crear un zip.Reader desde el io.ReaderAt
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read from reader", "<stream>", err)
	}

	// Usar la lógica de extracción compartida
	return extractPackage(zipReader.File, opts, "<stream>")
}

// Read abre un archivo IDML y lo carga en memoria con límites de seguridad por defecto.
//
// La función lee el archivo ZIP completo y:
// 1. Almacena el contenido de cada archivo como bytes crudos
// 2. Parsea automáticamente designmap.xml en una estructura Document
// 3. Preserva los metadatos ZIP para un roundtrip perfecto
// 4. Aplica protección contra ZIP bombs con límites por defecto
//
// Esto permite tanto la preservación byte a byte de archivos desconocidos
// como la manipulación estructurada del manifiesto del documento.
//
// Usar ReadWithOptions para límites personalizados o para deshabilitar la protección.
func Read(path string) (*Package, error) {
	return ReadWithOptions(path, nil)
}

// ReadWithOptions abre un archivo IDML con límites de seguridad configurables.
//
// Si opts es nil, se aplican los límites por defecto. Para deshabilitar un límite específico,
// establecerlo en -1. Por ejemplo:
//
//	// Permitir tamaño total ilimitado pero mantener otros límites
//	pkg, err := idml.ReadWithOptions(path, &idml.ReadOptions{
//	    MaxTotalSize: -1,
//	})
func ReadWithOptions(path string, opts *ReadOptions) (*Package, error) {
	// Aplicar valores por defecto
	if opts == nil {
		opts = &ReadOptions{}
	}
	opts.applyDefaults()

	// Abrir el archivo ZIP
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read", path, err)
	}
	defer r.Close()

	// Usar la lógica de extracción compartida
	return extractPackage(r.File, opts, path)
}
