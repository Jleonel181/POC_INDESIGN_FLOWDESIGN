package common

import (
	"errors"
	"fmt"
)

// Errores centinela comunes compartidos entre todos los paquetes.
var (
	// ErrNotFound se retorna cuando el elemento solicitado no existe.
	ErrNotFound = errors.New("not found")

	// ErrInvalidFormat se retorna cuando el XML o los datos tienen una estructura inválida.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrAlreadyExists se retorna cuando se intenta agregar un elemento que ya existe.
	ErrAlreadyExists = errors.New("already exists")

	// ErrMissingDependency se retorna cuando falta un recurso requerido.
	ErrMissingDependency = errors.New("missing required dependency")

	// ErrMissingMetadata se retorna cuando faltan metadatos requeridos.
	ErrMissingMetadata = errors.New("missing required metadata")
)

// Error representa un error de operación con contexto.
// Es el tipo de error unificado usado en todos los paquetes IDML.
//
// DECISIÓN DE DISEÑO: Contexto de error estructurado
// En lugar de usar errores simples o múltiples tipos de error por paquete, usamos un único
// struct Error que captura el contexto de dónde y cómo ocurrió el error.
// Esto provee manejo de errores consistente en todos los paquetes manteniendo
// compatibilidad con los patrones de wrapping de Go (errors.Is, errors.As, errors.Unwrap).
// El enfoque estructurado permite mejor debugging y reporte de errores en operaciones
// complejas que abarcan múltiples paquetes y archivos.
type Error struct {
	// Package identifica el paquete donde se originó el error.
	// Ejemplos: "idml", "idms", "spread", "story", "resources"
	Package string

	// Op describe la operación que se estaba realizando cuando ocurrió el error.
	// Ejemplos: "read", "write", "parse", "marshal"
	Op string

	// Path es la ruta del archivo o recurso involucrado, si aplica.
	// Puede estar vacío si no hay una ruta específica involucrada.
	Path string

	// Err es el error subyacente que causó este error.
	Err error
}

// Error implementa la interfaz error con un formato consistente.
func (e *Error) Error() string {
	// Formato: "paquete: operación [ruta]: error subyacente"
	var msg string
	if e.Package != "" {
		msg = e.Package + ": "
	}
	if e.Op != "" {
		msg += e.Op
	}
	if e.Path != "" {
		msg += " " + e.Path
	}
	if e.Err != nil {
		if msg != "" {
			msg += ": "
		}
		msg += e.Err.Error()
	}
	return msg
}

// Unwrap retorna el error subyacente para uso con errors.Is/As.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError crea un nuevo Error con los parámetros dados.
// Es un constructor de conveniencia para crear errores.
func NewError(pkg, op, path string, err error) *Error {
	return &Error{
		Package: pkg,
		Op:      op,
		Path:    path,
		Err:     err,
	}
}

// WrapError envuelve un error existente con contexto de paquete y operación.
// Si err es nil, retorna nil.
func WrapError(pkg, op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Package: pkg,
		Op:      op,
		Err:     err,
	}
}

// WrapErrorWithPath envuelve un error existente con contexto de paquete, operación y ruta.
// Si err es nil, retorna nil.
func WrapErrorWithPath(pkg, op, path string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Package: pkg,
		Op:      op,
		Path:    path,
		Err:     err,
	}
}

// Errorf crea un nuevo Error con un mensaje formateado como error subyacente.
func Errorf(pkg, op, path, format string, args ...interface{}) *Error {
	return &Error{
		Package: pkg,
		Op:      op,
		Path:    path,
		Err:     fmt.Errorf(format, args...),
	}
}

// IsNotFound verifica si un error es o envuelve ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidFormat verifica si un error es o envuelve ErrInvalidFormat.
func IsInvalidFormat(err error) bool {
	return errors.Is(err, ErrInvalidFormat)
}

// IsAlreadyExists verifica si un error es o envuelve ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
