package spread

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NewOval crea un nuevo Oval con los valores por defecto requeridos.
//
// Parámetros:
//   - self: identificador único del óvalo (ej: "u1e6")
//   - centerX, centerY: coordenadas del punto central en puntos
//   - width, height: dimensiones del óvalo en puntos
//   - layer: ID de la capa (ej: "uba")
//
// Retorna un Oval con valores por defecto razonables para una elipse simple.
func NewOval(self string, centerX, centerY, width, height float64, layer string) *Oval {
	oval := &Oval{
		PageItemBase: PageItemBase{
			Self:      self,
			ItemLayer: layer,
			Visible:   "true",
		},
		ContentType:         "Unassigned",
		LockState:           "None",
		Locked:              "false",
		LocalDisplaySetting: "Default",
	}

	// Establece los bounds
	oval.SetBounds(centerX, centerY, width, height)

	return oval
}

// GetBounds extrae el bounding box del óvalo desde GeometricBounds.
//
// Retorna:
//   - centerX, centerY: punto central del óvalo
//   - width, height: dimensiones del óvalo
//   - error: si GeometricBounds falta o está mal formado
//
// Formato de GeometricBounds: "y1 x1 y2 x2" (de la esquina superior izquierda a la inferior derecha)
func (o *Oval) GetBounds() (centerX, centerY, width, height float64, err error) {
	if o.GeometricBounds == "" {
		return 0, 0, 0, 0, common.Errorf("spread", "get oval bounds", "", "GeometricBounds is empty")
	}

	parts := strings.Fields(o.GeometricBounds)
	if len(parts) != 4 {
		return 0, 0, 0, 0, common.Errorf("spread", "get oval bounds", "", "invalid GeometricBounds format: %s (expected 4 values)", o.GeometricBounds)
	}

	// Parsea los bounds (y1 x1 y2 x2)
	y1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}
	x1, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}
	y2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}
	x2, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}

	// Calcula las dimensiones
	width = x2 - x1
	height = y2 - y1
	centerX = x1 + width/2
	centerY = y1 + height/2

	return centerX, centerY, width, height, nil
}

// SetBounds actualiza el bounding box del óvalo en GeometricBounds.
//
// Parámetros:
//   - centerX, centerY: punto central del óvalo
//   - width, height: dimensiones del óvalo
func (o *Oval) SetBounds(centerX, centerY, width, height float64) {
	// Calcula los bounds (de la esquina superior izquierda a la inferior derecha)
	x1 := centerX - width/2
	y1 := centerY - height/2
	x2 := centerX + width/2
	y2 := centerY + height/2

	// Formato: "y1 x1 y2 x2"
	o.GeometricBounds = fmt.Sprintf("%g %g %g %g", y1, x1, y2, x2)
}

// Area calcula el área del óvalo.
//
// Retorna:
//   - el área en puntos cuadrados
//   - 0 si GeometricBounds es inválido
func (o *Oval) Area() float64 {
	_, _, width, height, err := o.GetBounds()
	if err != nil {
		return 0
	}

	// Área de la elipse: π * a * b (donde a y b son los semiejes mayor y menor)
	a := width / 2
	b := height / 2
	return math.Pi * a * b
}

// Circumference calcula la circunferencia aproximada del óvalo.
//
// Retorna:
//   - la circunferencia aproximada en puntos
//   - 0 si GeometricBounds es inválido
//
// Usa la aproximación de Ramanujan para la circunferencia de una elipse.
func (o *Oval) Circumference() float64 {
	_, _, width, height, err := o.GetBounds()
	if err != nil {
		return 0
	}

	a := width / 2  // semieje mayor
	b := height / 2 // semieje menor

	// Primera aproximación de Ramanujan: π * (3(a+b) - √((3a+b)(a+3b)))
	sum := a + b
	term := (3*a + b) * (a + 3*b)
	return math.Pi * (3*sum - math.Sqrt(term))
}

// IsCircle retorna true si el óvalo es un círculo perfecto (ancho igual a alto, dentro de la tolerancia).
func (o *Oval) IsCircle() bool {
	_, _, width, height, err := o.GetBounds()
	if err != nil {
		return false
	}
	return math.Abs(width-height) < 0.01 // Tolerancia de 0.01 puntos
}

// Width retorna el ancho del óvalo.
func (o *Oval) Width() float64 {
	_, _, width, _, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return width
}

// Height retorna el alto del óvalo.
func (o *Oval) Height() float64 {
	_, _, _, height, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return height
}

// CenterX retorna la coordenada X del centro del óvalo.
func (o *Oval) CenterX() float64 {
	centerX, _, _, _, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return centerX
}

// CenterY retorna la coordenada Y del centro del óvalo.
func (o *Oval) CenterY() float64 {
	_, centerY, _, _, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return centerY
}

// SetStroke establece las propiedades de borde del óvalo.
//
// Parámetros:
//   - color: referencia de color (ej: "Color/Black")
//   - weight: ancho del borde en puntos
//   - tint: porcentaje de tinta (0-100), usar 100 para color sólido
func (o *Oval) SetStroke(color string, weight float64, tint float64) {
	o.StrokeColor = color
	o.StrokeWeight = fmt.Sprintf("%g", weight)
	o.StrokeTint = fmt.Sprintf("%g", tint)
}

// SetFill establece las propiedades de relleno del óvalo.
//
// Parámetros:
//   - color: referencia de color (ej: "Color/Red")
//   - tint: porcentaje de tinta (0-100), usar 100 para color sólido
func (o *Oval) SetFill(color string, tint float64) {
	o.FillColor = color
	o.FillTint = fmt.Sprintf("%g", tint)
}
