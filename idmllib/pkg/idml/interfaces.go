package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// PackageReader provee acceso de solo lectura al contenido IDML.
// Esta interfaz permite el testing y mocking de dependencias de Package.
type PackageReader interface {
	// Stories retorna todos los archivos Story parseados del directorio Stories/.
	Stories() (map[string]*story.Story, error)

	// Story retorna un Story parseado del directorio Stories/.
	Story(filename string) (*story.Story, error)

	// Spreads retorna todos los archivos Spread parseados del directorio Spreads/.
	Spreads() (map[string]*spread.Spread, error)

	// Spread retorna un Spread parseado del directorio Spreads/.
	Spread(filename string) (*spread.Spread, error)

	// Fonts retorna el archivo Fonts.xml tipado.
	Fonts() (*resources.FontsFile, error)

	// Styles retorna el archivo Styles.xml tipado.
	Styles() (*resources.StylesFile, error)

	// Graphics retorna el archivo Graphic.xml tipado.
	Graphics() (*resources.GraphicFile, error)
}

// PackageWriter provee acceso de escritura al contenido IDML.
// Esta interfaz permite el testing y mocking de modificaciones de Package.
type PackageWriter interface {
	// SetFonts actualiza el archivo de fuentes en caché.
	// El archivo será serializado cuando se llame a Write().
	SetFonts(fonts *resources.FontsFile)

	// SetStyles actualiza el archivo de estilos en caché.
	// El archivo será serializado cuando se llame a Write().
	SetStyles(styles *resources.StylesFile)

	// SetGraphics actualiza el archivo de gráficos en caché.
	// El archivo será serializado cuando se llame a Write().
	SetGraphics(graphics *resources.GraphicFile)
}

// PackageAccessor combina el acceso de lectura y escritura al contenido IDML.
// Es la interfaz principal usada por ResourceManager y otros componentes
// que necesitan tanto leer como escribir.
type PackageAccessor interface {
	PackageReader
	PackageWriter
}

// PageItem representa cualquier elemento visual que puede colocarse en un spread.
// Todos los elementos de página comparten atributos comunes como ID Self, capa, bounds, transformación, visibilidad y nombre.
// Esta interfaz permite operaciones polimórficas sobre distintos tipos de elementos de página.
type PageItem interface {
	// GetSelf retorna el identificador único de este elemento de página
	GetSelf() string

	// GetItemLayer retorna el ID de la capa en la que se encuentra este elemento
	GetItemLayer() string

	// GetGeometricBounds retorna el bounding box en formato "y1 x1 y2 x2"
	GetGeometricBounds() string

	// GetItemTransform retorna la matriz de transformación de 6 valores
	GetItemTransform() string

	// GetVisible retorna el estado de visibilidad ("true" o "false")
	GetVisible() string

	// GetName retorna el nombre de visualización del elemento de página
	GetName() string
}

// Verificación en tiempo de compilación de que Package implementa PackageAccessor.
var _ PackageAccessor = (*Package)(nil)
