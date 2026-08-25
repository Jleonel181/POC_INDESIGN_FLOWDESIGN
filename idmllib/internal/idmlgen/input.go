// Package idmlgen genera archivos IDML a partir de una descripción de documento JSON.
// Es agnóstico al dominio editorial: solo recibe páginas con marcos posicionados en mm.
package idmlgen

// DocumentInput es el modelo de entrada que describe un documento IDML a generar.
// Corresponde al IdmlDocumentDTO del backend TypeScript.
type DocumentInput struct {
	Document           DocumentSpec        `json:"document"`
	Pages              []PageSpec          `json:"pages"`
	MasterSpreadSource *MasterSpreadSource `json:"masterSpreadSource,omitempty"`

	// BaseDir es el directorio base para resolver rutas de imágenes locales.
	// Se configura desde el flag -base-dir del CLI, no desde el JSON.
	BaseDir string `json:"-"`
}

// DocumentSpec describe las propiedades globales del documento.
type DocumentSpec struct {
	WidthMm     float64     `json:"widthMm"`
	HeightMm    float64     `json:"heightMm"`
	Margins     MarginsSpec `json:"margins"`
	FacingPages bool        `json:"facingPages"`
	Columns     int         `json:"columns"`
	Guides      []GuideSpec `json:"guides"`
}

// MarginsSpec describe los márgenes del documento en milímetros.
type MarginsSpec struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// GuideSpec describe una guía (línea de referencia) en el documento.
type GuideSpec struct {
	Orientation string  `json:"orientation"` // "vertical" o "horizontal"
	LocationMm  float64 `json:"locationMm"`  // posición en mm desde el borde de la página
}

// PageSpec describe una página del documento con sus marcos.
type PageSpec struct {
	Frames []FrameSpec `json:"frames"`
}

// FrameSpec describe un marco de texto posicionado en la página.
type FrameSpec struct {
	Type    string       `json:"type"` // "text" o "image"
	Name    string       `json:"name"`
	Bounds  BoundsSpec   `json:"bounds"`
	Content string       `json:"content"` // texto para type:"text"
	Options FrameOptions `json:"options"`

	// Campos de imagen (para type:"image")
	ImagePath   string `json:"imagePath,omitempty"`   // ruta local relativa a -base-dir
	ImageBase64 string `json:"imageBase64,omitempty"` // imagen en base64 estándar
}

// BoundsSpec describe los límites de un marco en milímetros.
type BoundsSpec struct {
	TopMm    float64 `json:"topMm"`
	LeftMm   float64 `json:"leftMm"`
	BottomMm float64 `json:"bottomMm"`
	RightMm  float64 `json:"rightMm"`
}

// FrameOptions describe opciones adicionales del marco.
type FrameOptions struct {
	VerticalJustification string `json:"verticalJustification"`
	ContentIsRaw          bool   `json:"contentIsRaw"`
}

// MasterSpreadSource describe de dónde extraer un MasterSpread existente para
// inyectarlo en el documento generado. El dominio envía la ruta al IDML plantilla y
// el nombre del master spread dentro de ese archivo.
type MasterSpreadSource struct {
	// TemplatePath es la ruta absoluta al archivo .idml que contiene el master spread.
	TemplatePath string `json:"templatePath"`

	// MasterSpreadName es el atributo Name del <MasterSpread> a extraer.
	// Ejemplo: "02-Noticias Apertura"
	MasterSpreadName string `json:"masterSpreadName"`

	// FolioDate es la fecha a inyectar en el encabezado del folio. Reemplaza el
	// placeholder {{fecha}} en las stories del master. Si está vacío no se sustituye.
	FolioDate string `json:"folioDate,omitempty"`
}
