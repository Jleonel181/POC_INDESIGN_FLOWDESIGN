package spread

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NewGraphicLine crea un nuevo GraphicLine con los valores por defecto requeridos.
//
// Parámetros:
//   - self: identificador único de la línea (ej: "u1e6")
//   - x1, y1: coordenadas del punto inicial en puntos
//   - x2, y2: coordenadas del punto final en puntos
//   - layer: ID de la capa (ej: "uba")
//
// Retorna un GraphicLine con valores por defecto razonables para una línea recta simple.
func NewGraphicLine(self string, x1, y1, x2, y2 float64, layer string) *GraphicLine {
	line := &GraphicLine{
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

	// Establece los extremos usando PathGeometry
	line.SetEndpoints(x1, y1, x2, y2)

	return line
}

// GetEndpoints extrae los extremos de la línea desde PathGeometry.
//
// Retorna:
//   - x1, y1: coordenadas del punto inicial
//   - x2, y2: coordenadas del punto final
//   - error: si PathGeometry falta o está mal formado
//
// Para líneas rectas simples, PathGeometry contiene dos entradas PathPointType.
func (gl *GraphicLine) GetEndpoints() (x1, y1, x2, y2 float64, err error) {
	if gl.Properties == nil || gl.Properties.PathGeometry == nil {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "PathGeometry is nil")
	}

	pathGeom := gl.Properties.PathGeometry
	if pathGeom.GeometryPathType == nil {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "GeometryPathType is nil")
	}

	geomPathType := pathGeom.GeometryPathType
	if geomPathType.PathPointArray == nil || len(geomPathType.PathPointArray.PathPoints) < 2 {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "insufficient path points (need at least 2)")
	}

	points := geomPathType.PathPointArray.PathPoints

	// Parsea el primer punto (formato "y1 x1")
	anchor1 := strings.Fields(points[0].Anchor)
	if len(anchor1) != 2 {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "invalid anchor format for first point: %s", points[0].Anchor)
	}

	y1, err = strconv.ParseFloat(anchor1[0], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	x1, err = strconv.ParseFloat(anchor1[1], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	// Parsea el segundo punto (formato "y2 x2")
	anchor2 := strings.Fields(points[1].Anchor)
	if len(anchor2) != 2 {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "invalid anchor format for second point: %s", points[1].Anchor)
	}

	y2, err = strconv.ParseFloat(anchor2[0], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	x2, err = strconv.ParseFloat(anchor2[1], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	return x1, y1, x2, y2, nil
}

// SetEndpoints actualiza los extremos de la línea en PathGeometry.
//
// Parámetros:
//   - x1, y1: coordenadas del punto inicial
//   - x2, y2: coordenadas del punto final
//
// Crea PathGeometry si no existe.
func (gl *GraphicLine) SetEndpoints(x1, y1, x2, y2 float64) {
	// Formato: "y x" para cada punto
	anchor1 := fmt.Sprintf("%g %g", y1, x1)
	anchor2 := fmt.Sprintf("%g %g", y2, x2)

	// Inicializa Properties si es necesario
	if gl.Properties == nil {
		gl.Properties = &common.Properties{}
	}

	// Inicializa PathGeometry si es necesario
	if gl.Properties.PathGeometry == nil {
		gl.Properties.PathGeometry = &common.PathGeometry{
			GeometryPathType: &common.GeometryPathType{
				PathOpen: "true", // Las líneas son paths abiertos
				PathPointArray: &common.PathPointArray{
					PathPoints: make([]common.PathPointType, 2),
				},
			},
		}
	}

	// Asegura que haya al menos 2 puntos
	if len(gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints) < 2 {
		gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints = make([]common.PathPointType, 2)
	}

	// Establece los puntos de anclaje (para líneas rectas, las direcciones izq/der coinciden con el anchor)
	gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints[0] = common.PathPointType{
		Anchor:         anchor1,
		LeftDirection:  anchor1,
		RightDirection: anchor1,
	}

	gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints[1] = common.PathPointType{
		Anchor:         anchor2,
		LeftDirection:  anchor2,
		RightDirection: anchor2,
	}
}

// Length calcula la distancia euclidiana entre los extremos de la línea.
//
// Retorna:
//   - la longitud de la línea en puntos
//   - 0 si PathGeometry es inválido
func (gl *GraphicLine) Length() float64 {
	x1, y1, x2, y2, err := gl.GetEndpoints()
	if err != nil {
		return 0
	}

	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// Angle retorna el ángulo de la línea en grados (0-360).
//
// El ángulo se mide desde el eje horizontal (derecha es 0°),
// en sentido contrario a las agujas del reloj.
//
// Retorna:
//   - ángulo en grados [0, 360)
//   - 0 si PathGeometry es inválido
func (gl *GraphicLine) Angle() float64 {
	x1, y1, x2, y2, err := gl.GetEndpoints()
	if err != nil {
		return 0
	}

	// Calcula el ángulo en radianes
	radians := math.Atan2(y2-y1, x2-x1)

	// Convierte a grados
	degrees := radians * 180 / math.Pi

	// Normaliza a [0, 360)
	if degrees < 0 {
		degrees += 360
	}

	return degrees
}

// IsHorizontal retorna true si la línea es perfectamente horizontal (dentro de la tolerancia).
func (gl *GraphicLine) IsHorizontal() bool {
	_, y1, _, y2, err := gl.GetEndpoints()
	if err != nil {
		return false
	}
	return math.Abs(y2-y1) < 0.01 // Tolerancia de 0.01 puntos
}

// IsVertical retorna true si la línea es perfectamente vertical (dentro de la tolerancia).
func (gl *GraphicLine) IsVertical() bool {
	x1, _, x2, _, err := gl.GetEndpoints()
	if err != nil {
		return false
	}
	return math.Abs(x2-x1) < 0.01 // Tolerancia de 0.01 puntos
}
