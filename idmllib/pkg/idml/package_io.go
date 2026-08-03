package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Métodos helper de I/O de archivos para el struct Package.
// Estos métodos proveen utilidades internas de acceso y manipulación de archivos.

// SetFileData establece o reemplaza el contenido de un archivo en el paquete.
// Si el archivo ya existe, preserva su ZIP header. Si es nuevo, lo agrega al final
// del orden de archivos.
func (p *Package) SetFileData(filename string, data []byte) {
	p.setFileData(filename, data)
}

// GetFileData retorna el contenido crudo de un archivo del paquete.
func (p *Package) GetFileData(filename string) ([]byte, error) {
	return p.getFileData(filename)
}

// InvalidateCache descarta los structs parseados en caché para un archivo, forzando
// su re-lectura desde los datos crudos la próxima vez que se acceda.
func (p *Package) InvalidateCache(path string) {
	p.invalidateCache(path)
}

// hasFile verifica si un archivo existe en el package.
func (p *Package) hasFile(filename string) bool {
	_, exists := p.files[filename]
	return exists
}

// getFileData retorna los datos crudos de un archivo.
// Retorna ErrNotFound si el archivo no existe.
func (p *Package) getFileData(filename string) ([]byte, error) {
	// Agregar validación para el nombre de archivo
	if filename == "" {
		return nil, common.Errorf("idml", "get file data", "", "filename is empty")
	}

	entry, exists := p.files[filename]
	if !exists {
		return nil, common.WrapErrorWithPath("idml", "get file data", filename, common.ErrNotFound)
	}

	// Agregar validación para la entrada del archivo
	if entry == nil {
		return nil, common.WrapErrorWithPath("idml", "get file data", filename, common.Errorf("idml", "get file data", filename, "file entry is nil"))
	}

	return entry.data, nil
}

// setFileData establece los datos crudos de un archivo.
// Crea un nuevo fileEntry si el archivo no existe.
// Preserva el ZIP header existente si el archivo ya existe.
func (p *Package) setFileData(filename string, data []byte) {
	// Agregar validación para el nombre de archivo
	if filename == "" {
		// Registrar error pero no fallar - esta es una función void
		return
	}

	// Agregar validación para data (nil está permitido para archivos vacíos)
	if data == nil {
		data = []byte{}
	}

	if entry, exists := p.files[filename]; exists {
		// Preservar el header existente, actualizar los datos
		entry.data = data
	} else {
		// Crear nueva entrada
		p.files[filename] = &fileEntry{
			data: data,
			// el header se creará durante Write() si es necesario
		}
		// Agregar al orden de archivos si es un archivo nuevo
		p.fileOrder = append(p.fileOrder, filename)
	}
}

// removeFile elimina un archivo del package.
// Retorna true si el archivo fue eliminado, false si no existía.
func (p *Package) removeFile(filename string) bool {
	// Agregar validación para el nombre de archivo
	if filename == "" {
		return false
	}

	if _, exists := p.files[filename]; !exists {
		return false
	}

	// Eliminar del mapa de archivos
	delete(p.files, filename)

	// Eliminar de fileOrder
	for i, name := range p.fileOrder {
		if name == filename {
			p.fileOrder = append(p.fileOrder[:i], p.fileOrder[i+1:]...)
			break
		}
	}

	return true
}

// getFileEntry retorna el fileEntry completo de un archivo.
// Provee acceso tanto a los datos como a los metadatos ZIP.
// Retorna ErrNotFound si el archivo no existe.
func (p *Package) getFileEntry(filename string) (*fileEntry, error) {
	// Agregar validación para el nombre de archivo
	if filename == "" {
		return nil, common.Errorf("idml", "get file entry", "", "filename is empty")
	}

	entry, exists := p.files[filename]
	if !exists {
		return nil, common.WrapErrorWithPath("idml", "get file entry", filename, common.ErrNotFound)
	}

	// Agregar validación para la entrada del archivo
	if entry == nil {
		return nil, common.WrapErrorWithPath("idml", "get file entry", filename, common.Errorf("idml", "get file entry", filename, "file entry is nil"))
	}

	return entry, nil
}

// copyFileData crea una copia de los datos de un archivo para prevenir modificaciones accidentales.
// Retorna ErrNotFound si el archivo no existe.
func (p *Package) copyFileData(filename string) ([]byte, error) {
	// Agregar validación para el nombre de archivo
	if filename == "" {
		return nil, common.Errorf("idml", "copy file data", "", "filename is empty")
	}

	entry, exists := p.files[filename]
	if !exists {
		return nil, common.WrapErrorWithPath("idml", "copy file data", filename, common.ErrNotFound)
	}

	// Agregar validación para la entrada del archivo
	if entry == nil {
		return nil, common.WrapErrorWithPath("idml", "copy file data", filename, common.Errorf("idml", "copy file data", filename, "file entry is nil"))
	}

	// Manejar data nil de forma segura
	if entry.data == nil {
		return []byte{}, nil
	}

	// Crear una copia para prevenir la modificación de los datos originales
	dataCopy := make([]byte, len(entry.data))
	copy(dataCopy, entry.data)
	return dataCopy, nil
}

// getFileSize retorna el tamaño de un archivo en bytes.
// Retorna 0 si el archivo no existe.
func (p *Package) getFileSize(filename string) int {
	entry, exists := p.files[filename]
	if !exists {
		return 0
	}
	return len(entry.data)
}

// listFilesByPattern retorna todos los nombres de archivo que coinciden con un patrón.
// Es útil para encontrar todos los archivos en un directorio (por ejemplo, "Stories/", "Spreads/").
func (p *Package) listFilesByPattern(pattern string) []string {
	var matches []string
	for filename := range p.files {
		// Coincidencia simple por prefijo - podría mejorarse con regex si fuera necesario
		if len(filename) >= len(pattern) && filename[:len(pattern)] == pattern {
			matches = append(matches, filename)
		}
	}
	return matches
}
