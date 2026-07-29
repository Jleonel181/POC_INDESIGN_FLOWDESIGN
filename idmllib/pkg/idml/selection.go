package idml

import (
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/spread"
)

// selectPageItemByID es un helper genérico que busca un page item por ID en todos los spreads.
// Reduce la duplicación de código en los métodos SelectXxxByID.
// NOTA: Esta es la búsqueda lineal de respaldo. Los métodos SelectXxxByID ahora usan el índice para lookups O(1).
// selectPageItemByID es una función genérica para seleccionar page items por ID
// Esta función no está en uso actualmente pero se conserva para posible uso futuro
// nolint:unused
func selectPageItemByID[T PageItem](p *Package, itemType string, getter func(*spread.SpreadElement) []T, getSelf func(*T) string, id string) (*T, error) {
	spreads, err := p.Spreads()
	if err != nil {
		return nil, common.WrapError("idml", "select "+itemType, fmt.Errorf("failed to load spreads: %w", err))
	}

	for _, sp := range spreads {
		items := getter(&sp.InnerSpread)
		for i := range items {
			if getSelf(&items[i]) == id {
				return &items[i], nil
			}
		}
	}

	return nil, common.WrapError("idml", "select "+itemType, fmt.Errorf("%s with ID '%s' not found", itemType, id))
}

// Selection representa una colección de page items seleccionados de un documento IDML.
// Se usa típicamente para exportar subconjuntos de un documento (ej. a snippets IDMS).
type Selection struct {
	// TextFrames contiene los text frames seleccionados
	TextFrames []*spread.SpreadTextFrame

	// Rectangles contiene los marcos rectangulares seleccionados (frecuentemente con imágenes)
	Rectangles []*spread.Rectangle

	// Ovals contiene los marcos ovalados seleccionados
	Ovals []*spread.Oval

	// Polygons contiene los marcos poligonales seleccionados
	Polygons []*spread.Polygon

	// GraphicLines contiene las líneas gráficas seleccionadas
	GraphicLines []*spread.GraphicLine

	// Groups contiene los grupos seleccionados
	Groups []*spread.Group
}

// NewSelection crea una nueva Selection vacía.
func NewSelection() *Selection {
	return &Selection{
		TextFrames:   []*spread.SpreadTextFrame{},
		Rectangles:   []*spread.Rectangle{},
		Ovals:        []*spread.Oval{},
		Polygons:     []*spread.Polygon{},
		GraphicLines: []*spread.GraphicLine{},
		Groups:       []*spread.Group{},
	}
}

// IsEmpty retorna true si la selección no contiene elementos.
func (s *Selection) IsEmpty() bool {
	return len(s.TextFrames) == 0 &&
		len(s.Rectangles) == 0 &&
		len(s.Ovals) == 0 &&
		len(s.Polygons) == 0 &&
		len(s.GraphicLines) == 0 &&
		len(s.Groups) == 0
}

// Count retorna el número total de elementos seleccionados.
func (s *Selection) Count() int {
	return len(s.TextFrames) +
		len(s.Rectangles) +
		len(s.Ovals) +
		len(s.Polygons) +
		len(s.GraphicLines) +
		len(s.Groups)
}

// AddPageItem agrega cualquier page item a la selección usando la interfaz PageItem.
// El método determina el tipo concreto y lo agrega al slice correspondiente.
func (s *Selection) AddPageItem(item PageItem) {
	switch v := item.(type) {
	case *spread.SpreadTextFrame:
		s.AddTextFrame(v)
	case *spread.Rectangle:
		s.AddRectangle(v)
	case *spread.Oval:
		s.AddOval(v)
	case *spread.Polygon:
		s.AddPolygon(v)
	case *spread.GraphicLine:
		s.AddGraphicLine(v)
	case *spread.Group:
		s.AddGroup(v)
	}
}

// GetAllPageItems retorna todos los page items seleccionados como un slice de la interfaz PageItem.
// Permite operaciones polimórficas sobre toda la selección.
func (s *Selection) GetAllPageItems() []PageItem {
	var items []PageItem

	for i := range s.TextFrames {
		items = append(items, s.TextFrames[i])
	}
	for i := range s.Rectangles {
		items = append(items, s.Rectangles[i])
	}
	for i := range s.Ovals {
		items = append(items, s.Ovals[i])
	}
	for i := range s.Polygons {
		items = append(items, s.Polygons[i])
	}
	for i := range s.GraphicLines {
		items = append(items, s.GraphicLines[i])
	}
	for i := range s.Groups {
		items = append(items, s.Groups[i])
	}

	return items
}

// AddTextFrame agrega un text frame a la selección.
func (s *Selection) AddTextFrame(tf *spread.SpreadTextFrame) {
	s.TextFrames = append(s.TextFrames, tf)
}

// AddRectangle agrega un rectángulo a la selección.
func (s *Selection) AddRectangle(rect *spread.Rectangle) {
	s.Rectangles = append(s.Rectangles, rect)
}

// AddOval agrega un óvalo a la selección.
func (s *Selection) AddOval(oval *spread.Oval) {
	s.Ovals = append(s.Ovals, oval)
}

// AddPolygon agrega un polígono a la selección.
func (s *Selection) AddPolygon(polygon *spread.Polygon) {
	s.Polygons = append(s.Polygons, polygon)
}

// AddGraphicLine agrega una línea gráfica a la selección.
func (s *Selection) AddGraphicLine(line *spread.GraphicLine) {
	s.GraphicLines = append(s.GraphicLines, line)
}

// AddGroup agrega un grupo a la selección.
func (s *Selection) AddGroup(group *spread.Group) {
	s.Groups = append(s.Groups, group)
}

// SelectPageItemByID busca y retorna cualquier page item por su ID Self usando la interfaz PageItem.
// Usa un índice interno para lookups O(1).
// Retorna error si el page item no es encontrado.
func (p *Package) SelectPageItemByID(id string) (PageItem, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select page item", fmt.Errorf("failed to build index: %w", err))
	}

	idx := p.indexState.index

	// Verificar cada tipo de page item
	if tf, ok := idx.textFrames[id]; ok {
		return tf, nil
	}
	if rect, ok := idx.rectangles[id]; ok {
		return rect, nil
	}
	if oval, ok := idx.ovals[id]; ok {
		return oval, nil
	}
	if poly, ok := idx.polygons[id]; ok {
		return poly, nil
	}
	if line, ok := idx.graphicLines[id]; ok {
		return line, nil
	}
	if group, ok := idx.groups[id]; ok {
		return group, nil
	}

	return nil, common.WrapError("idml", "select page item", fmt.Errorf("page item with ID '%s' not found", id))
}

// SelectTextFrameByID busca y retorna un text frame por su ID Self.
// Usa un índice interno para lookups O(1).
// Retorna error si el text frame no es encontrado.
func (p *Package) SelectTextFrameByID(id string) (*spread.SpreadTextFrame, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select text frame", fmt.Errorf("failed to build index: %w", err))
	}

	if tf, ok := p.indexState.index.textFrames[id]; ok {
		return tf, nil
	}

	return nil, common.WrapError("idml", "select text frame", fmt.Errorf("text frame with ID '%s' not found", id))
}

// SelectRectangleByID busca y retorna un rectángulo por su ID Self.
// Usa un índice interno para lookups O(1).
// Retorna error si el rectángulo no es encontrado.
func (p *Package) SelectRectangleByID(id string) (*spread.Rectangle, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select rectangle", fmt.Errorf("failed to build index: %w", err))
	}

	if rect, ok := p.indexState.index.rectangles[id]; ok {
		return rect, nil
	}

	return nil, common.WrapError("idml", "select rectangle", fmt.Errorf("rectangle with ID '%s' not found", id))
}

// SelectOvalByID busca y retorna un óvalo por su ID Self.
// Usa un índice interno para lookups O(1).
// Retorna error si el óvalo no es encontrado.
func (p *Package) SelectOvalByID(id string) (*spread.Oval, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select oval", fmt.Errorf("failed to build index: %w", err))
	}

	if oval, ok := p.indexState.index.ovals[id]; ok {
		return oval, nil
	}

	return nil, common.WrapError("idml", "select oval", fmt.Errorf("oval with ID '%s' not found", id))
}

// SelectPolygonByID busca y retorna un polígono por su ID Self.
// Usa un índice interno para lookups O(1).
// Retorna error si el polígono no es encontrado.
func (p *Package) SelectPolygonByID(id string) (*spread.Polygon, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select polygon", fmt.Errorf("failed to build index: %w", err))
	}

	if poly, ok := p.indexState.index.polygons[id]; ok {
		return poly, nil
	}

	return nil, common.WrapError("idml", "select polygon", fmt.Errorf("polygon with ID '%s' not found", id))
}

// SelectGraphicLineByID busca y retorna una línea gráfica por su ID Self.
// Usa un índice interno para lookups O(1).
// Retorna error si la línea gráfica no es encontrada.
func (p *Package) SelectGraphicLineByID(id string) (*spread.GraphicLine, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select graphic line", fmt.Errorf("failed to build index: %w", err))
	}

	if line, ok := p.indexState.index.graphicLines[id]; ok {
		return line, nil
	}

	return nil, common.WrapError("idml", "select graphic line", fmt.Errorf("graphic line with ID '%s' not found", id))
}

// SelectGroupByID busca y retorna un grupo por su ID Self.
// Usa un índice interno para lookups O(1).
// Retorna error si el grupo no es encontrado.
func (p *Package) SelectGroupByID(id string) (*spread.Group, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select group", fmt.Errorf("failed to build index: %w", err))
	}

	if group, ok := p.indexState.index.groups[id]; ok {
		return group, nil
	}

	return nil, common.WrapError("idml", "select group", fmt.Errorf("group with ID '%s' not found", id))
}

// SelectAllGraphicsInSpread retorna todos los rectángulos que contienen imágenes en el spread especificado.
// El spreadFilename debe tener el formato "Spreads/Spread_*.xml".
// Retorna error si el spread no es encontrado o no puede cargarse.
func (p *Package) SelectAllGraphicsInSpread(spreadFilename string) ([]*spread.Rectangle, error) {
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "select all graphics", spreadFilename, fmt.Errorf("failed to load spread '%s': %w", spreadFilename, err))
	}

	// Recopilar todos los rectángulos que tienen imágenes
	var graphics []*spread.Rectangle
	for i := range sp.InnerSpread.Rectangles {
		rect := &sp.InnerSpread.Rectangles[i]
		// Verificar si el rectángulo contiene una imagen
		if rect.ContentType == "GraphicType" || rect.Image != nil {
			graphics = append(graphics, rect)
		}
	}

	return graphics, nil
}

// SelectAllTextFramesInSpread retorna todos los text frames en el spread especificado.
// El spreadFilename debe tener el formato "Spreads/Spread_*.xml".
// Retorna error si el spread no es encontrado o no puede cargarse.
func (p *Package) SelectAllTextFramesInSpread(spreadFilename string) ([]*spread.SpreadTextFrame, error) {
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "select all text frames", spreadFilename, fmt.Errorf("failed to load spread '%s': %w", spreadFilename, err))
	}

	// Recopilar todos los text frames
	textFrames := make([]*spread.SpreadTextFrame, len(sp.InnerSpread.TextFrames))
	for i := range sp.InnerSpread.TextFrames {
		textFrames[i] = &sp.InnerSpread.TextFrames[i]
	}

	return textFrames, nil
}

// SelectPageItemsByIDs crea una Selection con todos los page items que tengan los IDs especificados.
// Usa un índice interno para lookups O(1) por ID.
// Los elementos no encontrados se omiten silenciosamente.
// Retorna tanto una Selection (por compatibilidad hacia atrás) como un slice de interfaces PageItem.
func (p *Package) SelectPageItemsByIDs(ids ...string) (*Selection, []PageItem, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, nil, common.WrapError("idml", "select page items by IDs", fmt.Errorf("failed to build index: %w", err))
	}

	selection := NewSelection()
	var pageItems []PageItem
	idx := p.indexState.index

	for _, id := range ids {
		// Verificar cada tipo de page item
		if tf, ok := idx.textFrames[id]; ok {
			selection.AddTextFrame(tf)
			pageItems = append(pageItems, tf)
			continue
		}
		if rect, ok := idx.rectangles[id]; ok {
			selection.AddRectangle(rect)
			pageItems = append(pageItems, rect)
			continue
		}
		if oval, ok := idx.ovals[id]; ok {
			selection.AddOval(oval)
			pageItems = append(pageItems, oval)
			continue
		}
		if poly, ok := idx.polygons[id]; ok {
			selection.AddPolygon(poly)
			pageItems = append(pageItems, poly)
			continue
		}
		if line, ok := idx.graphicLines[id]; ok {
			selection.AddGraphicLine(line)
			pageItems = append(pageItems, line)
			continue
		}
		if group, ok := idx.groups[id]; ok {
			selection.AddGroup(group)
			pageItems = append(pageItems, group)
		}
		// Si no se encuentra en ninguno, omitir silenciosamente (según documentación)
	}

	return selection, pageItems, nil
}

// SelectByIDs crea una Selection con todos los elementos que tengan los IDs especificados.
// Usa un índice interno para lookups O(1) por ID.
// Los elementos no encontrados se omiten silenciosamente.
func (p *Package) SelectByIDs(ids ...string) (*Selection, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select by IDs", fmt.Errorf("failed to build index: %w", err))
	}

	selection := NewSelection()
	idx := p.indexState.index

	for _, id := range ids {
		// Verificar cada tipo de page item
		if tf, ok := idx.textFrames[id]; ok {
			selection.AddTextFrame(tf)
			continue
		}
		if rect, ok := idx.rectangles[id]; ok {
			selection.AddRectangle(rect)
			continue
		}
		if oval, ok := idx.ovals[id]; ok {
			selection.AddOval(oval)
			continue
		}
		if poly, ok := idx.polygons[id]; ok {
			selection.AddPolygon(poly)
			continue
		}
		if line, ok := idx.graphicLines[id]; ok {
			selection.AddGraphicLine(line)
			continue
		}
		if group, ok := idx.groups[id]; ok {
			selection.AddGroup(group)
		}
		// Si no se encuentra en ninguno, omitir silenciosamente (según documentación)
	}

	return selection, nil
}
