package xmlutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/google/go-cmp/cmp"
)

// XMLDifference representa una única diferencia encontrada durante la comparación de XML.
type XMLDifference struct {
	Path        string // Ruta tipo XPath al elemento (ej: "/root/Story[0]/ParagraphStyleRange[2]")
	Type        string // "tag", "namespace", "attribute", "text", "structure"
	Description string // Descripción legible de la diferencia
	Expected    string // Valor esperado (del original)
	Got         string // Valor obtenido (del generado)
}

// CompareOptions controla cómo se realiza la comparación de XML.
type CompareOptions struct {
	// SortElements lista los nombres de etiquetas que deben ordenarse antes de comparar.
	// Útil para Resources y Styles de IDML donde el orden no importa.
	SortElements []string

	// MaxDifferences limita cuántas diferencias recolectar (0 = sin límite).
	// Evita acumular miles de diffs en XML muy distintos.
	MaxDifferences int

	// IgnoreWhitespace controla si se ignoran las diferencias de whitespace en texto.
	// Por defecto es true (el whitespace se recorta antes de comparar).
	IgnoreWhitespace bool
}

// DefaultCompareOptions retorna valores predeterminados razonables para comparación IDML.
func DefaultCompareOptions() *CompareOptions {
	return &CompareOptions{
		SortElements: []string{
			// Elementos IDML donde el orden no importa
			"FontFamily",
			"Color",
			"ParagraphStyle",
			"CharacterStyle",
			"ObjectStyle",
		},
		MaxDifferences:   100,
		IgnoreWhitespace: true,
	}
}

// CompareXMLWithDetails compara dos slices de bytes XML y retorna todas las diferencias encontradas.
// Es más informativo que CompareXML, que se detiene al primer error.
//
// Retorna un slice de XMLDifference describiendo las diferencias, o un error si el parseo falla.
// Un slice vacío significa que el XML es estructuralmente equivalente.
func CompareXMLWithDetails(original, generated []byte, opts *CompareOptions) ([]XMLDifference, error) {
	if opts == nil {
		opts = DefaultCompareOptions()
	}

	// Parsear el documento original
	origDoc := etree.NewDocument()
	if err := origDoc.ReadFromBytes(original); err != nil {
		return nil, fmt.Errorf("failed to parse original XML: %w", err)
	}

	// Parsear el documento generado
	genDoc := etree.NewDocument()
	if err := genDoc.ReadFromBytes(generated); err != nil {
		return nil, fmt.Errorf("failed to parse generated XML: %w", err)
	}

	// Comparar elementos raíz
	origRoot := origDoc.Root()
	genRoot := genDoc.Root()

	if origRoot == nil && genRoot == nil {
		return nil, nil // Ambos vacíos, sin diferencias
	}

	if origRoot == nil {
		return []XMLDifference{{
			Path:        "/",
			Type:        "structure",
			Description: "original has no root element, generated does",
			Expected:    "(none)",
			Got:         genRoot.Tag,
		}}, nil
	}

	if genRoot == nil {
		return []XMLDifference{{
			Path:        "/",
			Type:        "structure",
			Description: "original has root element, generated does not",
			Expected:    origRoot.Tag,
			Got:         "(none)",
		}}, nil
	}

	// Recolectar diferencias
	diffs := []XMLDifference{}
	compareElementsDetailed(origRoot, genRoot, "root", &diffs, opts)

	return diffs, nil
}

// compareElementsDetailed compara recursivamente dos elementos etree y recolecta todas las diferencias.
func compareElementsDetailed(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Verificar si se alcanzó el límite máximo de diferencias
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	// Comparar nombres de etiqueta
	if orig.Tag != gen.Tag {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "tag",
			Description: "tag name mismatch",
			Expected:    orig.Tag,
			Got:         gen.Tag,
		})
		return // No tiene sentido continuar comparando elementos distintos
	}

	// Comparar namespaces
	if orig.Space != gen.Space {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "namespace",
			Description: "namespace mismatch",
			Expected:    orig.Space,
			Got:         gen.Space,
		})
	}

	// Comparar atributos (sin importar el orden)
	compareAttributes(orig, gen, path, diffs, opts)

	// Comparar contenido de texto
	compareText(orig, gen, path, diffs, opts)

	// Comparar elementos hijos
	compareChildren(orig, gen, path, diffs, opts)
}

// compareAttributes compara los atributos de un elemento (sin importar el orden).
func compareAttributes(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Verificar límite antes de procesar
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	origAttrs := make(map[string]string)
	for _, attr := range orig.Attr {
		key := attr.Space + ":" + attr.Key
		origAttrs[key] = attr.Value
	}

	genAttrs := make(map[string]string)
	for _, attr := range gen.Attr {
		key := attr.Space + ":" + attr.Key
		genAttrs[key] = attr.Value
	}

	// Buscar atributos faltantes en el generado
	for key, origVal := range origAttrs {
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
		if genVal, exists := genAttrs[key]; !exists {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        "attribute",
				Description: fmt.Sprintf("attribute %q missing in generated", key),
				Expected:    origVal,
				Got:         "(missing)",
			})
		} else if genVal != origVal {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        "attribute",
				Description: fmt.Sprintf("attribute %q value differs", key),
				Expected:    origVal,
				Got:         genVal,
			})
		}
	}

	// Buscar atributos extra en el generado
	for key, genVal := range genAttrs {
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
		if _, exists := origAttrs[key]; !exists {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        "attribute",
				Description: fmt.Sprintf("attribute %q exists in generated but not in original", key),
				Expected:    "(none)",
				Got:         genVal,
			})
		}
	}
}

// compareText compara el contenido de texto de los elementos.
func compareText(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Verificar límite antes de procesar
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	origText := orig.Text()
	genText := gen.Text()

	// Aplicar recorte de whitespace si está habilitado
	if opts.IgnoreWhitespace {
		origText = strings.TrimSpace(origText)
		genText = strings.TrimSpace(genText)
	}

	if origText != genText {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "text",
			Description: "text content differs",
			Expected:    truncate(origText, 100),
			Got:         truncate(genText, 100),
		})
	}
}

// compareChildren compara los elementos hijos, con ordenamiento opcional.
func compareChildren(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Verificar límite antes de procesar
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	origChildren := orig.ChildElements()
	genChildren := gen.ChildElements()

	// Verificar si este tipo de elemento debe ordenarse antes de comparar
	shouldSort := contains(opts.SortElements, orig.Tag)

	if shouldSort {
		// Ordenar ambos hijos por nombre de etiqueta para comparación sin importar el orden
		sortElementsByTag(origChildren)
		sortElementsByTag(genChildren)
	}

	if len(origChildren) != len(genChildren) {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "structure",
			Description: "child element count mismatch",
			Expected:    fmt.Sprintf("%d children", len(origChildren)),
			Got:         fmt.Sprintf("%d children", len(genChildren)),
		})

		// Verificar límite después de agregar diff
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
	}

	// Comparar hijos en común
	minLen := len(origChildren)
	if len(genChildren) < minLen {
		minLen = len(genChildren)
	}

	for i := 0; i < minLen; i++ {
		childPath := fmt.Sprintf("%s/%s[%d]", path, origChildren[i].Tag, i)
		compareElementsDetailed(origChildren[i], genChildren[i], childPath, diffs, opts)

		// Verificar límite después de cada hijo
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
	}
}

// Funciones auxiliares

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func sortElementsByTag(elements []*etree.Element) {
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Tag < elements[j].Tag
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Funciones con compatibilidad hacia atrás que mantienen la API existente

// CompareXML compara dos slices de bytes XML para verificar equivalencia estructural.
// Usa la librería etree para parsear y comparar documentos XML.
// Maneja correctamente namespaces y atributos, e ignora whitespace no significativo.
//
// Retorna nil si el XML es estructuralmente equivalente, o un error con detalles si difiere.
// Esta es la API simple - usar CompareXMLWithDetails para obtener más información.
func CompareXML(original, generated []byte) error {
	return CompareXMLWithEtree(original, generated)
}

// CompareXMLWithEtree compara dos slices de bytes XML usando la librería etree.
// Provee mayor control sobre la comparación y puede manejar escenarios de namespace más complejos.
//
// Retorna nil si los árboles XML son equivalentes, o un error con detalles si difieren.
// Se detiene al primer error - usar CompareXMLWithDetails para recolectar todas las diferencias.
func CompareXMLWithEtree(original, generated []byte) error {
	// Parsear el documento original
	origDoc := etree.NewDocument()
	if err := origDoc.ReadFromBytes(original); err != nil {
		return fmt.Errorf("failed to parse original XML with etree: %w", err)
	}

	// Parsear el documento generado
	genDoc := etree.NewDocument()
	if err := genDoc.ReadFromBytes(generated); err != nil {
		return fmt.Errorf("failed to parse generated XML with etree: %w", err)
	}

	// Comparar elementos raíz
	origRoot := origDoc.Root()
	genRoot := genDoc.Root()

	if origRoot == nil && genRoot == nil {
		return nil // Ambos vacíos
	}
	if origRoot == nil || genRoot == nil {
		return fmt.Errorf("one document has no root element")
	}

	// Comparar recursivamente
	return compareElements(origRoot, genRoot, "root")
}

// compareElements compara recursivamente dos elementos etree.
// Es la implementación anterior que retorna al primer error.
func compareElements(orig, gen *etree.Element, path string) error {
	// Comparar nombres de etiqueta
	if orig.Tag != gen.Tag {
		return fmt.Errorf("%s: tag mismatch: %q vs %q", path, orig.Tag, gen.Tag)
	}

	// Comparar namespaces
	if orig.Space != gen.Space {
		return fmt.Errorf("%s: namespace mismatch: %q vs %q", path, orig.Space, gen.Space)
	}

	// Comparar atributos (sin importar el orden)
	origAttrs := make(map[string]string)
	for _, attr := range orig.Attr {
		key := attr.Space + ":" + attr.Key
		origAttrs[key] = attr.Value
	}

	genAttrs := make(map[string]string)
	for _, attr := range gen.Attr {
		key := attr.Space + ":" + attr.Key
		genAttrs[key] = attr.Value
	}

	if !cmp.Equal(origAttrs, genAttrs) {
		return fmt.Errorf("%s: attributes differ:\n%s", path, cmp.Diff(origAttrs, genAttrs))
	}

	// Comparar contenido de texto (ignorando whitespace no significativo)
	origText := strings.TrimSpace(orig.Text())
	genText := strings.TrimSpace(gen.Text())
	if origText != genText {
		return fmt.Errorf("%s: text content differs:\norig: %q\ngen:  %q", path, origText, genText)
	}

	// Comparar elementos hijos
	origChildren := orig.ChildElements()
	genChildren := gen.ChildElements()

	if len(origChildren) != len(genChildren) {
		return fmt.Errorf("%s: child count mismatch: %d vs %d", path, len(origChildren), len(genChildren))
	}

	for i := range origChildren {
		childPath := fmt.Sprintf("%s/%s[%d]", path, origChildren[i].Tag, i)
		if err := compareElements(origChildren[i], genChildren[i], childPath); err != nil {
			return err
		}
	}

	return nil
}

// NormalizeXML parsea y reformatea XML con indentación consistente.
// Útil para comparar archivos XML donde el formato difiere.
//
// Retorna los bytes XML normalizados y cualquier error de parseo.
func NormalizeXML(data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Formatear con indentación
	doc.Indent(2)

	// Escribir a bytes
	return doc.WriteToBytes()
}

// ParseToMap parsea XML en una estructura etree.Document.
// Útil cuando aún no se tienen structs tipados.
// Retorna el documento parseado o un error.
//
// Se trabaja con estructuras XML genéricas (etree.Document) en lugar de structs
// fuertemente tipados, lo que permite manejar cualquier XML sin definir structs primero.
func ParseToMap(data []byte) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	return doc, nil
}

// FormatDifferences retorna un resumen legible de las diferencias XML.
func FormatDifferences(diffs []XMLDifference) string {
	if len(diffs) == 0 {
		return "No se encontraron diferencias"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d difference(s):\n\n", len(diffs)))

	for i, diff := range diffs {
		b.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, diff.Path, diff.Type))
		b.WriteString(fmt.Sprintf("   %s\n", diff.Description))
		b.WriteString(fmt.Sprintf("   Expected: %s\n", diff.Expected))
		b.WriteString(fmt.Sprintf("   Got:      %s\n", diff.Got))
		if i < len(diffs)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
