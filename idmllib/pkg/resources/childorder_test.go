package resources

import (
	"os"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/dimelords/idmllib/v2/internal/testutil"
)

const (
	corpusReferenceStyles  = "../../testdata/documento_referencia/Resources/Styles.xml"
	corpusReferenceGraphic = "../../testdata/documento_referencia/Resources/Graphic.xml"
)

// childTagSequence devuelve los nombres de etiqueta de los hijos directos del
// elemento raíz, en orden.
func childTagSequence(t *testing.T, data []byte) []string {
	t.Helper()

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		t.Fatalf("no se pudo parsear el XML para leer la secuencia de hijos: %v", err)
	}
	root := doc.Root()
	if root == nil {
		t.Fatal("el XML no tiene elemento raíz")
	}

	tags := make([]string, 0, len(root.ChildElements()))
	for _, child := range root.ChildElements() {
		tags = append(tags, child.Tag)
	}
	return tags
}

func TestStylesChildOrder_OrdenDeCamposCubreTodasLasClases(t *testing.T) {
	testutil.AssertFieldOrderCovers(t, stylesChildOrder, []string{
		styleChildRootCharacterGroup, styleChildRootParagraphGroup,
		styleChildRootCellGroup, styleChildRootTableGroup, styleChildRootObjectGroup,
		styleChildTOCStyle, styleChildOther,
	})
}

func TestGraphicChildOrder_OrdenDeCamposCubreTodasLasClases(t *testing.T) {
	testutil.AssertFieldOrderCovers(t, graphicChildOrder, []string{
		graphicChildColor, graphicChildInk, graphicChildGradient, graphicChildSwatch,
		graphicChildPastedSmoothShade, graphicChildStrokeStyle, graphicChildOther,
	})
}

// TestStylesChildOrder_SecuenciaSobreviveElCiclo cubre el caso concreto que motivó la
// tarea en este archivo: InDesign pone TOCStyle en tercera posición y el orden de los
// campos del struct lo empujaba a la sexta.
func TestStylesChildOrder_SecuenciaSobreviveElCiclo(t *testing.T) {
	data, err := os.ReadFile(corpusReferenceStyles)
	if err != nil {
		t.Skipf("Styles.xml del Documento_Referencia ausente: %v (ruta esperada: %s)", err, corpusReferenceStyles)
	}

	styles, err := ParseStylesFile(data)
	if err != nil {
		t.Fatalf("ParseStylesFile falló: %v", err)
	}
	out, err := MarshalStylesFile(styles)
	if err != nil {
		t.Fatalf("MarshalStylesFile falló: %v", err)
	}

	want := childTagSequence(t, data)
	got := childTagSequence(t, out)

	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("la secuencia de hijos cambió al serializar\nesperada: %v\nobtenida: %v", want, got)
	}

	// La posición de TOCStyle es el síntoma concreto, así que se comprueba aparte:
	// un cambio en el orden de los campos que lo devolviera al final dejaría este
	// test en rojo con un mensaje que dice exactamente qué pasó.
	if len(got) < 3 || got[2] != styleChildTOCStyle {
		t.Errorf("TOCStyle debe quedar en la tercera posición, secuencia obtenida: %v", got)
	}
}

// TestGraphicChildOrder_SecuenciaSobreviveElCiclo cubre Graphic.xml, donde Gradient,
// Swatch y PastedSmoothShade venían intercalados y el orden de campos los reagrupaba.
func TestGraphicChildOrder_SecuenciaSobreviveElCiclo(t *testing.T) {
	data, err := os.ReadFile(corpusReferenceGraphic)
	if err != nil {
		t.Skipf("Graphic.xml del Documento_Referencia ausente: %v (ruta esperada: %s)", err, corpusReferenceGraphic)
	}

	graphic, err := ParseGraphicFile(data)
	if err != nil {
		t.Fatalf("ParseGraphicFile falló: %v", err)
	}
	out, err := MarshalGraphicFile(graphic)
	if err != nil {
		t.Fatalf("MarshalGraphicFile falló: %v", err)
	}

	want := childTagSequence(t, data)
	got := childTagSequence(t, out)

	if len(want) != 117 {
		t.Errorf("el Graphic.xml de referencia debería tener 117 hijos y tiene %d: ¿cambió el corpus?", len(want))
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("la secuencia de hijos cambió al serializar\nesperada: %v\nobtenida: %v", want, got)
	}
}

// TestGraphicChildOrder_DesdeCeroUsaOrdenDeCampos comprueba que un archivo construido
// a mano, sin parsear, sigue emitiendo en el orden de los campos.
func TestGraphicChildOrder_DesdeCeroUsaOrdenDeCampos(t *testing.T) {
	graphic := &GraphicFile{
		DOMVersion: "20.4",
		// Asignados al revés del orden de los campos a propósito.
		StrokeStyles: []StrokeStyle{{Self: "ss1", Name: "$ID/Solid"}},
		Colors:       []Color{{Self: "c1", Name: "Negro"}},
	}

	if graphic.childOrder.Recorded() {
		t.Fatal("un GraphicFile construido desde cero no debe tener orden registrado")
	}

	out, err := MarshalGraphicFile(graphic)
	if err != nil {
		t.Fatalf("MarshalGraphicFile falló: %v", err)
	}

	got := childTagSequence(t, out)
	if strings.Join(got, " ") != "Color StrokeStyle" {
		t.Errorf("sin registro se debe emitir en orden de campos (Color antes de StrokeStyle), obtenido: %v", got)
	}
}
