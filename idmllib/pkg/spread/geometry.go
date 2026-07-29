package spread

import (
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Bounds representa el ancho y alto de un frame en puntos.
type Bounds struct {
	Width  float64
	Height float64
}

// Position representa las coordenadas x e y de un frame en puntos.
type Position struct {
	X float64
	Y float64
}

// Transform representa una matriz de transformación completa de 6 valores.
// Formato: [a b c d x y]
// Donde:
//   - a, d: factores de escala
//   - b, c: factores de rotación/inclinación
//   - x, y: traslación (posición)
type Transform struct {
	A float64 // Escala X
	B float64 // Rotación/Inclinación
	C float64 // Rotación/Inclinación
	D float64 // Escala Y
	X float64 // Traslación X
	Y float64 // Traslación Y
}

// Bounds retorna el ancho y alto del text frame desde su GeometricBounds.
// Formato de GeometricBounds: "y1 x1 y2 x2" (coordenadas superior-izquierda e inferior-derecha)
//
// Si GeometricBounds está vacío o es inválido, recurre a calcular los bounds desde
// PathGeometry en Properties (usado comúnmente en InDesign para frames con formas personalizadas).
//
// Retorna un error si tanto GeometricBounds como PathGeometry no están disponibles o están malformados.
func (tf *SpreadTextFrame) Bounds() (Bounds, error) {
	// Intentar GeometricBounds primero (caso más común)
	bounds, err := parseGeometricBounds(tf.GeometricBounds)
	if err == nil && (bounds.Width > 0 && bounds.Height > 0) {
		return bounds, nil
	}

	// Recurrir a PathGeometry si GeometricBounds está vacío o es inválido
	return tf.BoundsFromPathGeometry()
}

// Position retorna la posición x e y del text frame desde su ItemTransform.
// Formato de ItemTransform: "a b c d x y" (matriz de transformación de 6 valores)
//
// Retorna un error si ItemTransform está vacío o malformado.
func (tf *SpreadTextFrame) Position() (Position, error) {
	return parseItemTransformPosition(tf.ItemTransform)
}

// Transform retorna la matriz de transformación completa del text frame.
// Formato de ItemTransform: "a b c d x y" (matriz de transformación de 6 valores)
//
// Retorna un error si ItemTransform está vacío o malformado.
func (tf *SpreadTextFrame) Transform() (Transform, error) {
	return parseItemTransform(tf.ItemTransform)
}

// Bounds retorna el ancho y alto del rectángulo desde su GeometricBounds.
func (r *Rectangle) Bounds() (Bounds, error) {
	return parseGeometricBounds(r.GeometricBounds)
}

// Position retorna la posición x e y del rectángulo desde su ItemTransform.
func (r *Rectangle) Position() (Position, error) {
	return parseItemTransformPosition(r.ItemTransform)
}

// Transform retorna la matriz de transformación completa del rectángulo.
func (r *Rectangle) Transform() (Transform, error) {
	return parseItemTransform(r.ItemTransform)
}

// Bounds retorna el ancho y alto del óvalo desde su GeometricBounds.
func (o *Oval) Bounds() (Bounds, error) {
	return parseGeometricBounds(o.GeometricBounds)
}

// Position retorna la posición x e y del óvalo desde su ItemTransform.
func (o *Oval) Position() (Position, error) {
	return parseItemTransformPosition(o.ItemTransform)
}

// Transform retorna la matriz de transformación completa del óvalo.
func (o *Oval) Transform() (Transform, error) {
	return parseItemTransform(o.ItemTransform)
}

// Bounds retorna el ancho y alto del polígono desde su GeometricBounds.
func (p *Polygon) Bounds() (Bounds, error) {
	return parseGeometricBounds(p.GeometricBounds)
}

// Position retorna la posición x e y del polígono desde su ItemTransform.
func (p *Polygon) Position() (Position, error) {
	return parseItemTransformPosition(p.ItemTransform)
}

// Transform retorna la matriz de transformación completa del polígono.
func (p *Polygon) Transform() (Transform, error) {
	return parseItemTransform(p.ItemTransform)
}

// parseGeometricBounds parsea el formato de string GeometricBounds.
// Formato: "y1 x1 y2 x2" (coordenadas superior-izquierda e inferior-derecha en puntos)
func parseGeometricBounds(bounds string) (Bounds, error) {
	if bounds == "" {
		return Bounds{}, common.Errorf("spread", "parse geometric bounds", "", "GeometricBounds is empty")
	}

	parts := strings.Fields(bounds)
	if len(parts) != 4 {
		return Bounds{}, common.Errorf("spread", "parse geometric bounds", "", "invalid GeometricBounds format: expected 4 values, got %d", len(parts))
	}

	// Agregar recuperación ante posibles panics durante el parseo
	defer func() {
		if r := recover(); r != nil {
			// Esto no debería ocurrir con la validación anterior, pero provee seguridad
		}
	}()

	y1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	x1, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	y2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	x2, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	return Bounds{
		Width:  x2 - x1,
		Height: y2 - y1,
	}, nil
}

// parseItemTransformPosition parsea solo la posición (x, y) desde ItemTransform.
// Formato: "a b c d x y" (matriz de transformación de 6 valores)
func parseItemTransformPosition(transform string) (Position, error) {
	if transform == "" {
		return Position{}, common.Errorf("spread", "parse item transform position", "", "ItemTransform is empty")
	}

	parts := strings.Fields(transform)
	if len(parts) != 6 {
		return Position{}, common.Errorf("spread", "parse item transform position", "", "invalid ItemTransform format: expected 6 values, got %d", len(parts))
	}

	x, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return Position{}, common.WrapError("spread", "parse item transform position", err)
	}

	y, err := strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return Position{}, common.WrapError("spread", "parse item transform position", err)
	}

	return Position{X: x, Y: y}, nil
}

// parseItemTransform parsea la matriz ItemTransform completa.
// Formato: "a b c d x y" (matriz de transformación de 6 valores)
func parseItemTransform(transform string) (Transform, error) {
	if transform == "" {
		return Transform{}, common.Errorf("spread", "parse item transform", "", "ItemTransform is empty")
	}

	parts := strings.Fields(transform)
	if len(parts) != 6 {
		return Transform{}, common.Errorf("spread", "parse item transform", "", "invalid ItemTransform format: expected 6 values, got %d", len(parts))
	}

	var t Transform
	var err error

	t.A, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.B, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.C, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.D, err = strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.X, err = strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.Y, err = strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	return t, nil
}

// BoundsFromPathGeometry calcula los bounds desde PathGeometry en Properties.
// Se usa como alternativa cuando el atributo GeometricBounds no está presente.
// PathGeometry contiene PathPointArray con puntos de anclaje que definen la forma del frame.
//
// Retorna un error si Properties, PathGeometry o PathPointArray es nil/vacío.
func (tf *SpreadTextFrame) BoundsFromPathGeometry() (Bounds, error) {
	if tf.Properties == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "Properties is nil")
	}
	if tf.Properties.PathGeometry == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "PathGeometry is nil")
	}
	if tf.Properties.PathGeometry.GeometryPathType == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "GeometryPathType is nil")
	}
	if tf.Properties.PathGeometry.GeometryPathType.PathPointArray == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "PathPointArray is nil")
	}

	points := tf.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints
	if len(points) == 0 {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "PathPointArray is empty")
	}

	// Parsear todos los puntos de anclaje para encontrar X e Y mínimos/máximos
	minX, minY := parseAnchorPoint(points[0].Anchor)
	maxX, maxY := minX, minY

	for _, point := range points[1:] {
		x, y := parseAnchorPoint(point.Anchor)
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	return Bounds{
		Width:  maxX - minX,
		Height: maxY - minY,
	}, nil
}

// parseAnchorPoint parsea un string de punto de anclaje "x y" y retorna x, y como floats.
// Retorna 0, 0 si el parseo falla (maneja errores silenciosamente por conveniencia).
func parseAnchorPoint(anchor string) (float64, float64) {
	parts := strings.Fields(anchor)
	if len(parts) != 2 {
		return 0, 0
	}

	x, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0
	}

	y, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0
	}

	return x, y
}
