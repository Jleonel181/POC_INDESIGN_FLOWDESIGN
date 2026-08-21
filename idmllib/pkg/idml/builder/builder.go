// Package builder expone la API pública de generación de documentos IDML.
//
// Es una fachada sobre internal/idmlgen que hace pública la capacidad de construir
// un documento IDML completo a partir de una especificación declarativa. La implementación
// real vive en internal/idmlgen y es la misma que usa cmd/idmlgen.
//
// Uso:
//
//	opts := builder.DocumentOptions{
//	    Width:  595.28, // A4 en puntos
//	    Height: 841.89,
//	    Margins: builder.Margins{Top: 36, Bottom: 36, Left: 36, Right: 36},
//	}
//	doc := builder.NewDocument(opts)
//	doc.AddPage(builder.PageOptions{
//	    Frames: []builder.FrameOptions{{
//	        Type:    "text",
//	        Bounds:  builder.Bounds{Top: 36, Left: 36, Bottom: 805, Right: 559},
//	        Content: "Hello World",
//	    }},
//	})
//	pkg, err := doc.Build()
package builder

import (
	"github.com/dimelords/idmllib/v2/internal/idmlgen"
	"github.com/dimelords/idmllib/v2/pkg/idml"
)

// DocumentOptions configura las propiedades globales del documento.
type DocumentOptions struct {
	// Dimensiones en puntos (1 pt = 1/72 pulgada).
	Width  float64
	Height float64

	// Margins en puntos.
	Margins Margins

	// FacingPages activa páginas enfrentadas.
	FacingPages bool

	// Columns define el número de columnas de texto (por defecto: 1).
	Columns int

	// Guides define guías editoriales globales.
	Guides []GuideOptions

	// MasterSpreadSource permite inyectar un MasterSpread desde un .idml externo.
	MasterSpreadSource *MasterSpreadSource
}

// Margins define los márgenes del documento en puntos.
type Margins struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

// GuideOptions define una guía editorial.
type GuideOptions struct {
	Orientation string  // "vertical" o "horizontal"
	LocationMm  float64 // posición en mm
}

// MasterSpreadSource permite inyectar un master spread desde un paquete externo.
type MasterSpreadSource = idmlgen.MasterSpreadSource

// PageOptions configura una página y sus frames.
type PageOptions struct {
	Frames []FrameOptions
}

// FrameOptions configura un frame de contenido.
type FrameOptions struct {
	// Type: "text" o "image"
	Type string

	// Name es el nombre de visualización del frame.
	Name string

	// Bounds define la posición y tamaño en mm.
	Bounds Bounds

	// Content es el texto para frames de tipo "text".
	Content string

	// Options adicionales del frame (vertical justification, raw content, etc.).
	Options FrameExtraOptions
}

// Bounds define la geometría de un frame en mm.
type Bounds struct {
	TopMm    float64
	LeftMm   float64
	BottomMm float64
	RightMm  float64
}

// FrameExtraOptions define opciones adicionales para un frame.
type FrameExtraOptions = idmlgen.FrameOptions

// Builder construye un documento IDML paso a paso.
type Builder struct {
	opts  DocumentOptions
	pages []PageOptions
}

// NewDocument crea un builder con las opciones del documento.
func NewDocument(opts DocumentOptions) *Builder {
	return &Builder{opts: opts}
}

// AddPage agrega una página al documento con sus frames.
func (b *Builder) AddPage(page PageOptions) {
	b.pages = append(b.pages, page)
}

// Build construye el paquete IDML completo.
// Error si las opciones son inválidas (dimensiones ≤ 0, sin páginas, etc.).
func (b *Builder) Build() (*idml.Package, error) {
	input := b.toInput()
	if err := idmlgen.Validate(input); err != nil {
		return nil, err
	}
	return idmlgen.Generate(input)
}

// toInput convierte las opciones del builder al formato de internal/idmlgen.
func (b *Builder) toInput() *idmlgen.DocumentInput {
	guides := make([]idmlgen.GuideSpec, 0, len(b.opts.Guides))
	for _, g := range b.opts.Guides {
		guides = append(guides, idmlgen.GuideSpec{
			Orientation: g.Orientation,
			LocationMm:  g.LocationMm,
		})
	}

	pages := make([]idmlgen.PageSpec, 0, len(b.pages))
	for _, p := range b.pages {
		frames := make([]idmlgen.FrameSpec, 0, len(p.Frames))
		for _, f := range p.Frames {
			frames = append(frames, idmlgen.FrameSpec{
				Type:    f.Type,
				Name:    f.Name,
				Content: f.Content,
				Bounds: idmlgen.BoundsSpec{
					TopMm:    f.Bounds.TopMm,
					LeftMm:   f.Bounds.LeftMm,
					BottomMm: f.Bounds.BottomMm,
					RightMm:  f.Bounds.RightMm,
				},
				Options: idmlgen.FrameOptions(f.Options),
			})
		}
		pages = append(pages, idmlgen.PageSpec{Frames: frames})
	}

	return &idmlgen.DocumentInput{
		Document: idmlgen.DocumentSpec{
			WidthMm:  b.opts.Width / 2.834645669, // pt → mm
			HeightMm: b.opts.Height / 2.834645669,
			Margins: idmlgen.MarginsSpec{
				Top:    b.opts.Margins.Top / 2.834645669,
				Bottom: b.opts.Margins.Bottom / 2.834645669,
				Left:   b.opts.Margins.Left / 2.834645669,
				Right:  b.opts.Margins.Right / 2.834645669,
			},
			FacingPages: b.opts.FacingPages,
			Columns:     b.opts.Columns,
			Guides:      guides,
		},
		Pages:              pages,
		MasterSpreadSource: b.opts.MasterSpreadSource,
	}
}
