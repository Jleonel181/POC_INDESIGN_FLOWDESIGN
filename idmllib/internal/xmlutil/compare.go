package xmlutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/google/go-cmp/cmp"
)

// Categorías de diferencia. Las seis primeras son las que enumera el Requisito 1,
// criterio 2 del spec idml-generator-api, y están pensadas para poder contarlas y
// filtrarlas una por una: un criterio de aceptación del tipo «cero diferencias de
// categoría X» necesita que X sea un valor y no un matiz de la descripción.
//
// CategoryTag y CategoryNamespace quedan fuera de esas seis porque describen otra
// cosa —el nombre o el espacio de nombres del propio elemento— y perderlas sería
// perder información. Tras el emparejamiento por etiqueta de compareChildren,
// CategoryTag solo puede aparecer comparando los elementos raíz.
const (
	CategoryAttributeMissing = "atributo-ausente"
	CategoryAttributeValue   = "atributo-valor-distinto"
	CategoryAttributeExtra   = "atributo-sobrante"
	CategoryElementMissing   = "elemento-ausente"
	CategoryElementExtra     = "elemento-sobrante"
	CategoryElementOrder     = "orden-elementos-distinto"
	CategoryText             = "texto-distinto"
	CategoryTag              = "etiqueta-distinta"
	CategoryNamespace        = "namespace-distinto"
)

// XMLDifference representa una única diferencia encontrada durante la comparación de XML.
type XMLDifference struct {
	Path        string // Ruta tipo XPath al elemento (ej: "/root/Story[0]/ParagraphStyleRange[2]")
	Type        string // Categoría de la diferencia: una de las constantes Category*
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
			Type:        CategoryElementExtra,
			Description: "el original no tiene elemento raíz y el generado sí",
			Expected:    "(ninguno)",
			Got:         genRoot.Tag,
		}}, nil
	}

	if genRoot == nil {
		return []XMLDifference{{
			Path:        "/",
			Type:        CategoryElementMissing,
			Description: "el original tiene elemento raíz y el generado no",
			Expected:    origRoot.Tag,
			Got:         "(ninguno)",
		}}, nil
	}

	// Recolectar diferencias
	diffs := []XMLDifference{}
	compareElementsDetailed(origRoot, genRoot, "root", &diffs, opts)

	return diffs, nil
}

// compareElementsDetailed compara recursivamente dos elementos etree y recolecta todas las diferencias.
func compareElementsDetailed(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	if limitReached(diffs, opts) {
		return
	}

	// Comparar nombres de etiqueta
	if orig.Tag != gen.Tag {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        CategoryTag,
			Description: "el nombre del elemento no coincide",
			Expected:    orig.Tag,
			Got:         gen.Tag,
		})
		return // No tiene sentido continuar comparando elementos distintos
	}

	// Comparar namespaces
	if orig.Space != gen.Space {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        CategoryNamespace,
			Description: "el espacio de nombres no coincide",
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
	if limitReached(diffs, opts) {
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

	// Buscar atributos faltantes o con valor distinto en el generado.
	// Se recorre en orden para que la salida no dependa del recorrido del mapa.
	for _, key := range sortedKeys(origAttrs) {
		if limitReached(diffs, opts) {
			return
		}
		origVal := origAttrs[key]
		if genVal, exists := genAttrs[key]; !exists {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        CategoryAttributeMissing,
				Description: fmt.Sprintf("atributo %q ausente en el generado", key),
				Expected:    origVal,
				Got:         "(ausente)",
			})
		} else if genVal != origVal {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        CategoryAttributeValue,
				Description: fmt.Sprintf("el atributo %q tiene otro valor", key),
				Expected:    origVal,
				Got:         genVal,
			})
		}
	}

	// Buscar atributos extra en el generado
	for _, key := range sortedKeys(genAttrs) {
		if limitReached(diffs, opts) {
			return
		}
		if _, exists := origAttrs[key]; !exists {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        CategoryAttributeExtra,
				Description: fmt.Sprintf("atributo %q sobrante en el generado", key),
				Expected:    "(ninguno)",
				Got:         genAttrs[key],
			})
		}
	}
}

// compareText compara el contenido de texto de los elementos.
func compareText(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	if limitReached(diffs, opts) {
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
			Type:        CategoryText,
			Description: "el contenido de texto es distinto",
			Expected:    truncate(origText, 100),
			Got:         truncate(genText, 100),
		})
	}
}

// compareChildren compara los elementos hijos, con ordenamiento opcional.
//
// Distingue dos situaciones que antes se confundían. Si el conjunto de hijos no
// coincide, alguno falta o sobra y se reporta uno por uno. Si coincide pero la
// secuencia no, están todos y solo se han reordenado: eso es **una** diferencia
// en el padre, no una por cada posición desplazada.
//
// Después, los hijos se emparejan por (nombre de etiqueta, ordinal de aparición)
// y no por posición absoluta. Es lo que evita que un reordenamiento desalinee la
// comparación y produzca diferencias de atributo entre elementos que no se
// corresponden. Cuando las dos secuencias son iguales, este emparejamiento y el
// posicional son el mismo.
func compareChildren(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	if limitReached(diffs, opts) {
		return
	}

	origChildren := orig.ChildElements()
	genChildren := gen.ChildElements()

	// Verificar si este tipo de elemento debe ordenarse antes de comparar
	if contains(opts.SortElements, orig.Tag) {
		// Ordenar ambos hijos por nombre de etiqueta para comparación sin importar el orden
		sortElementsByTag(origChildren)
		sortElementsByTag(genChildren)
	}

	origTags := tagSequence(origChildren)
	genTags := tagSequence(genChildren)
	missing, extra := tagMultisetDiff(origTags, genTags)

	switch {
	case len(missing) > 0 || len(extra) > 0:
		for _, tag := range missing {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        CategoryElementMissing,
				Description: fmt.Sprintf("elemento hijo %q ausente en el generado", tag),
				Expected:    tag,
				Got:         "(ausente)",
			})
			if limitReached(diffs, opts) {
				return
			}
		}
		for _, tag := range extra {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        CategoryElementExtra,
				Description: fmt.Sprintf("elemento hijo %q sobrante en el generado", tag),
				Expected:    "(ninguno)",
				Got:         tag,
			})
			if limitReached(diffs, opts) {
				return
			}
		}

	case !equalStrings(origTags, genTags):
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        CategoryElementOrder,
			Description: fmt.Sprintf("los %d elementos hijos son los mismos pero en otro orden", len(origTags)),
			Expected:    truncate(strings.Join(origTags, ", "), 200),
			Got:         truncate(strings.Join(genTags, ", "), 200),
		})
		if limitReached(diffs, opts) {
			return
		}
	}

	for _, pair := range pairChildrenByTag(origChildren, genChildren) {
		childPath := fmt.Sprintf("%s/%s[%d]", path, pair.orig.Tag, pair.origIndex)
		compareElementsDetailed(pair.orig, pair.gen, childPath, diffs, opts)
		if limitReached(diffs, opts) {
			return
		}
	}
}

// Funciones auxiliares

// limitReached indica si ya se alcanzó el máximo de diferencias a recolectar.
// MaxDifferences igual a 0 significa sin límite.
func limitReached(diffs *[]XMLDifference, opts *CompareOptions) bool {
	return opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences
}

// tagSequence devuelve los nombres de etiqueta de los elementos, en su orden.
func tagSequence(elements []*etree.Element) []string {
	tags := make([]string, len(elements))
	for i, e := range elements {
		tags[i] = e.Tag
	}
	return tags
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// tagMultisetDiff compara las dos secuencias como multiconjuntos, es decir sin
// mirar el orden pero sí las repeticiones, y devuelve qué etiquetas faltan en la
// segunda y cuáles sobran. Una etiqueta con 3 apariciones en el original y 1 en el
// generado aparece 2 veces en missing.
//
// Ambas listas salen ordenadas alfabéticamente para que el reporte sea estable
// entre ejecuciones.
func tagMultisetDiff(origTags, genTags []string) (missing, extra []string) {
	balance := make(map[string]int, len(origTags))
	for _, tag := range origTags {
		balance[tag]++
	}
	for _, tag := range genTags {
		balance[tag]--
	}

	tags := make([]string, 0, len(balance))
	for tag := range balance {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	for _, tag := range tags {
		for i := 0; i < balance[tag]; i++ {
			missing = append(missing, tag)
		}
		for i := 0; i > balance[tag]; i-- {
			extra = append(extra, tag)
		}
	}
	return missing, extra
}

// childPair es un hijo del original con el hijo del generado que le corresponde.
// origIndex es la posición del hijo en el original, y se conserva para que la ruta
// del reporte siga señalando el documento de entrada.
type childPair struct {
	orig, gen *etree.Element
	origIndex int
}

// pairChildrenByTag empareja la n-ésima aparición de una etiqueta en el original
// con la n-ésima aparición de esa misma etiqueta en el generado. Los hijos del
// original sin pareja se omiten: su ausencia ya se reportó como elemento ausente.
func pairChildrenByTag(origChildren, genChildren []*etree.Element) []childPair {
	byTag := make(map[string][]*etree.Element, len(genChildren))
	for _, e := range genChildren {
		byTag[e.Tag] = append(byTag[e.Tag], e)
	}

	seen := make(map[string]int, len(origChildren))
	pairs := make([]childPair, 0, len(origChildren))
	for i, e := range origChildren {
		n := seen[e.Tag]
		seen[e.Tag]++
		if n < len(byTag[e.Tag]) {
			pairs = append(pairs, childPair{orig: e, gen: byTag[e.Tag][n], origIndex: i})
		}
	}
	return pairs
}

// sortedKeys devuelve las claves de un mapa ordenadas alfabéticamente.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

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
	b.WriteString(fmt.Sprintf("Se encontraron %d diferencia(s):\n\n", len(diffs)))

	for i, diff := range diffs {
		b.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, diff.Path, diff.Type))
		b.WriteString(fmt.Sprintf("   %s\n", diff.Description))
		b.WriteString(fmt.Sprintf("   Esperado:  %s\n", diff.Expected))
		b.WriteString(fmt.Sprintf("   Obtenido:  %s\n", diff.Got))
		if i < len(diffs)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
