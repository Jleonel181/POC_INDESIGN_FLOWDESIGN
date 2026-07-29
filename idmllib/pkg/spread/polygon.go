package spread

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NewPolygon crea un nuevo Polygon con los valores por defecto requeridos.
//
// Parámetros:
//   - self: identificador único del polígono (ej: "u1e6")
//   - vertices: arreglo de pares de coordenadas [x, y] que definen los vértices del polígono
//   - layer: ID de la capa (ej: "uba")
//
// Retorna un Polygon con valores por defecto razonables.
func NewPolygon(self string, vertices [][2]float64, layer string) *Polygon {
	polygon := &Polygon{
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

	// Establece los vértices usando PathGeometry
	polygon.SetVertices(vertices)

	return polygon
}

// NewRegularPolygon crea un polígono regular (todos los lados iguales) centrado en un punto.
//
// Parámetros:
//   - self: identificador único
//   - centerX, centerY: coordenadas del punto central
//   - radius: distancia del centro a cada vértice
//   - sides: número de lados (mínimo 3)
//   - rotation: ángulo de rotación en grados (0 = primer vértice apunta a la derecha)
//   - layer: ID de la capa
//
// Retorna un Polygon con los vértices distribuidos en un patrón regular.
func NewRegularPolygon(self string, centerX, centerY, radius float64, sides int, rotation float64, layer string) *Polygon {
	if sides < 3 {
		sides = 3
	}

	vertices := make([][2]float64, sides)
	angleStep := 360.0 / float64(sides)

	for i := 0; i < sides; i++ {
		// Calcula el ángulo para este vértice
		angle := (float64(i)*angleStep + rotation) * math.Pi / 180

		// Calcula la posición del vértice
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)

		vertices[i] = [2]float64{x, y}
	}

	return NewPolygon(self, vertices, layer)
}

// GetVertices extrae los vértices del polígono desde PathGeometry.
//
// Retorna:
//   - arreglo de pares de coordenadas [x, y]
//   - error: si PathGeometry falta o está mal formado
func (p *Polygon) GetVertices() ([][2]float64, error) {
	if p.Properties == nil || p.Properties.PathGeometry == nil {
		return nil, common.Errorf("spread", "get polygon vertices", "", "PathGeometry is nil")
	}

	pathGeom := p.Properties.PathGeometry
	if pathGeom.GeometryPathType == nil {
		return nil, common.Errorf("spread", "get polygon vertices", "", "GeometryPathType is nil")
	}

	geomPathType := pathGeom.GeometryPathType
	if geomPathType.PathPointArray == nil || len(geomPathType.PathPointArray.PathPoints) < 3 {
		return nil, common.Errorf("spread", "get polygon vertices", "", "insufficient path points (need at least 3 for a polygon)")
	}

	points := geomPathType.PathPointArray.PathPoints
	vertices := make([][2]float64, len(points))

	for i, point := range points {
		// Parsea el anchor (formato "y x")
		anchor := strings.Fields(point.Anchor)
		if len(anchor) != 2 {
			return nil, common.Errorf("spread", "get polygon vertices", "", "invalid anchor format for point %d: %s", i, point.Anchor)
		}

		y, err := strconv.ParseFloat(anchor[0], 64)
		if err != nil {
			return nil, common.WrapError("spread", "get polygon vertices", err)
		}

		x, err := strconv.ParseFloat(anchor[1], 64)
		if err != nil {
			return nil, common.WrapError("spread", "get polygon vertices", err)
		}

		vertices[i] = [2]float64{x, y}
	}

	return vertices, nil
}

// SetVertices actualiza los vértices del polígono en PathGeometry.
//
// Parámetros:
//   - vertices: arreglo de pares de coordenadas [x, y]
func (p *Polygon) SetVertices(vertices [][2]float64) {
	if len(vertices) < 3 {
		return // Se necesitan al menos 3 vértices para un polígono
	}

	// Inicializa Properties si es necesario
	if p.Properties == nil {
		p.Properties = &common.Properties{}
	}

	// Inicializa PathGeometry si es necesario
	if p.Properties.PathGeometry == nil {
		p.Properties.PathGeometry = &common.PathGeometry{
			GeometryPathType: &common.GeometryPathType{
				PathOpen: "false", // Los polígonos son paths cerrados
				PathPointArray: &common.PathPointArray{
					PathPoints: make([]common.PathPointType, len(vertices)),
				},
			},
		}
	}

	// Asegura que haya suficientes puntos
	if len(p.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints) != len(vertices) {
		p.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints = make([]common.PathPointType, len(vertices))
	}

	// Establece cada vértice
	for i, vertex := range vertices {
		anchor := fmt.Sprintf("%g %g", vertex[1], vertex[0]) // formato "y x"

		p.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints[i] = common.PathPointType{
			Anchor:         anchor,
			LeftDirection:  anchor, // Para lados rectos
			RightDirection: anchor, // Para lados rectos
		}
	}
}

// VertexCount retorna el número de vértices del polígono.
func (p *Polygon) VertexCount() int {
	vertices, err := p.GetVertices()
	if err != nil {
		return 0
	}
	return len(vertices)
}

// Perimeter calcula la longitud total de todos los lados.
func (p *Polygon) Perimeter() float64 {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 2 {
		return 0
	}

	perimeter := 0.0

	// Suma las distancias entre vértices consecutivos
	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices) // Vuelve al primer vértice al llegar al final
		dx := vertices[next][0] - vertices[i][0]
		dy := vertices[next][1] - vertices[i][1]
		perimeter += math.Sqrt(dx*dx + dy*dy)
	}

	return perimeter
}

// Area calcula el área usando la fórmula del área de Gauss (shoelace formula).
func (p *Polygon) Area() float64 {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 3 {
		return 0
	}

	// Fórmula del área de Gauss (shoelace formula)
	area := 0.0
	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)
		area += vertices[i][0] * vertices[next][1]
		area -= vertices[next][0] * vertices[i][1]
	}

	return math.Abs(area) / 2.0
}

// Centroid calcula el centro geométrico del polígono.
//
// Retorna:
//   - centerX, centerY: coordenadas del centroide
//   - error: si no se pueden obtener los vértices
func (p *Polygon) Centroid() (centerX, centerY float64, err error) {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 3 {
		return 0, 0, common.Errorf("spread", "calculate centroid", "", "insufficient vertices")
	}

	// Calcula el centroide
	for _, vertex := range vertices {
		centerX += vertex[0]
		centerY += vertex[1]
	}

	centerX /= float64(len(vertices))
	centerY /= float64(len(vertices))

	return centerX, centerY, nil
}

// BoundingBox retorna el bounding box alineado a los ejes del polígono.
//
// Retorna:
//   - minX, minY, maxX, maxY: coordenadas del bounding box
//   - error: si no se pueden obtener los vértices
func (p *Polygon) BoundingBox() (minX, minY, maxX, maxY float64, err error) {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) == 0 {
		return 0, 0, 0, 0, common.Errorf("spread", "calculate bounding box", "", "no vertices")
	}

	// Inicializa con el primer vértice
	minX, maxX = vertices[0][0], vertices[0][0]
	minY, maxY = vertices[0][1], vertices[0][1]

	// Busca el mínimo/máximo
	for i := 1; i < len(vertices); i++ {
		if vertices[i][0] < minX {
			minX = vertices[i][0]
		}
		if vertices[i][0] > maxX {
			maxX = vertices[i][0]
		}
		if vertices[i][1] < minY {
			minY = vertices[i][1]
		}
		if vertices[i][1] > maxY {
			maxY = vertices[i][1]
		}
	}

	return minX, minY, maxX, maxY, nil
}

// IsRegular verifica si el polígono es aproximadamente regular (todos los lados iguales).
// Usa una tolerancia del 1% para la variación de longitud de los lados.
func (p *Polygon) IsRegular() bool {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 3 {
		return false
	}

	// Calcula la longitud de todos los lados
	sides := make([]float64, len(vertices))
	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)
		dx := vertices[next][0] - vertices[i][0]
		dy := vertices[next][1] - vertices[i][1]
		sides[i] = math.Sqrt(dx*dx + dy*dy)
	}

	// Verifica si todos los lados son aproximadamente iguales
	avgLength := 0.0
	for _, length := range sides {
		avgLength += length
	}
	avgLength /= float64(len(sides))

	tolerance := avgLength * 0.01 // Tolerancia del 1%

	for _, length := range sides {
		if math.Abs(length-avgLength) > tolerance {
			return false
		}
	}

	return true
}

// SetStroke establece las propiedades de borde del polígono.
func (p *Polygon) SetStroke(color string, weight float64, tint float64) {
	p.StrokeColor = color
	p.StrokeWeight = fmt.Sprintf("%g", weight)
	p.StrokeTint = fmt.Sprintf("%g", tint)
}

// SetFill establece las propiedades de relleno del polígono.
func (p *Polygon) SetFill(color string, tint float64) {
	p.FillColor = color
	p.FillTint = fmt.Sprintf("%g", tint)
}
