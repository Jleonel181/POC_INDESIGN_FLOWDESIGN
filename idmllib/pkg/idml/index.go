package idml

import (
	"sync"

	"github.com/dimelords/idmllib/v2/pkg/spread"
)

// itemIndex provee lookups O(1) para page items por su ID Self.
// Cachea punteros a todos los page items de todos los spreads.
// El índice se construye de forma lazy en el primer acceso.
type itemIndex struct {
	textFrames   map[string]*spread.SpreadTextFrame
	rectangles   map[string]*spread.Rectangle
	ovals        map[string]*spread.Oval
	polygons     map[string]*spread.Polygon
	graphicLines map[string]*spread.GraphicLine
	groups       map[string]*spread.Group
}

// newItemIndex crea un nuevo índice vacío con los mapas inicializados.
func newItemIndex() *itemIndex {
	return &itemIndex{
		textFrames:   make(map[string]*spread.SpreadTextFrame),
		rectangles:   make(map[string]*spread.Rectangle),
		ovals:        make(map[string]*spread.Oval),
		polygons:     make(map[string]*spread.Polygon),
		graphicLines: make(map[string]*spread.GraphicLine),
		groups:       make(map[string]*spread.Group),
	}
}

// itemIndexState contiene el estado del índice para un Package.
// Se embebe en el struct Package.
type itemIndexState struct {
	index *itemIndex
	once  sync.Once
	err   error
}

// ensureItemIndex construye el índice de elementos si aún no fue construido.
// Usa sync.Once para inicialización lazy y thread-safe.
func (p *Package) ensureItemIndex() error {
	p.indexState.once.Do(func() {
		p.indexState.index = newItemIndex()
		p.indexState.err = p.buildItemIndex()
	})

	return p.indexState.err
}

// buildItemIndex puebla el índice con todos los page items de todos los spreads.
func (p *Package) buildItemIndex() error {
	spreads, err := p.Spreads()
	if err != nil {
		return err
	}

	for _, sp := range spreads {
		// Indexar text frames
		for _, tf := range sp.InnerSpread.TextFrames() {
			p.indexState.index.textFrames[tf.Self] = tf
		}

		// Indexar rectángulos
		for _, rect := range sp.InnerSpread.Rectangles() {
			p.indexState.index.rectangles[rect.Self] = rect
		}

		// Indexar óvalos
		for _, oval := range sp.InnerSpread.Ovals() {
			p.indexState.index.ovals[oval.Self] = oval
		}

		// Indexar polígonos
		for _, poly := range sp.InnerSpread.Polygons() {
			p.indexState.index.polygons[poly.Self] = poly
		}

		// Indexar líneas gráficas
		for _, line := range sp.InnerSpread.GraphicLines() {
			p.indexState.index.graphicLines[line.Self] = line
		}

		// Indexar grupos
		for _, group := range sp.InnerSpread.Groups() {
			p.indexState.index.groups[group.Self] = group
		}
	}

	return nil
}

// ItemCount retorna el número total de elementos indexados.
// Retorna 0 si el índice aún no fue construido.
func (p *Package) ItemCount() int {
	if p.indexState.index == nil {
		return 0
	}
	return len(p.indexState.index.textFrames) +
		len(p.indexState.index.rectangles) +
		len(p.indexState.index.ovals) +
		len(p.indexState.index.polygons) +
		len(p.indexState.index.graphicLines) +
		len(p.indexState.index.groups)
}

// TextFrameCount retorna el número de text frames indexados.
func (p *Package) TextFrameCount() int {
	if p.indexState.index == nil {
		return 0
	}
	return len(p.indexState.index.textFrames)
}

// RectangleCount retorna el número de rectángulos indexados.
func (p *Package) RectangleCount() int {
	if p.indexState.index == nil {
		return 0
	}
	return len(p.indexState.index.rectangles)
}
