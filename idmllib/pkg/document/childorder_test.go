package document

import (
	"os"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/dimelords/idmllib/v2/internal/testutil"
)

// Rutas del corpus de fidelidad, relativas a este paquete. Se comparte la del
// Documento_Referencia con el arnés de pkg/idml.
const (
	corpusReferenceDesignmap = "../../testdata/documento_referencia/designmap.xml"
	corpusImagesDesignmap    = "../../testdata/imagenes_designmap.xml"
)

// childTagSequence devuelve los nombres de etiqueta de los hijos directos del
// elemento raíz, en orden. Es la secuencia que el ciclo tiene que conservar.
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
		name := child.Tag
		if child.Space != "" {
			name = child.Space + ":" + child.Tag
		}
		tags = append(tags, name)
	}
	return tags
}

// TestDocumentChildOrder_OrdenDeCamposCubreTodasLasClases evita el fallo silencioso
// que describe testutil.AssertFieldOrderCovers.
func TestDocumentChildOrder_OrdenDeCamposCubreTodasLasClases(t *testing.T) {
	testutil.AssertFieldOrderCovers(t, documentChildOrder, []string{
		childProperties, childLanguage,
		childRefGraphic, childRefFonts, childRefStyles, childRefPreferences, childRefTags,
		childRefMasterSpread, childRefSpread, childRefStory, childRefBackingStory,
		childLayer, childNumberingList, childNamedGrid, childSection, childDocumentUser,
		childColorGroup, childABullet, childAssignment, childTextVariable,
		childColor, childSwatch, childStrokeStyle,
		childRootCharStyleGroup, childRootParaStyleGroup, childRootObjStyleGroup,
		childTinDocumentData, childTransparencyDefault,
		childInlineSpread, childInlineStory,
		childOther,
	})
}

// TestDocumentChildOrder_SecuenciaSobreviveElCiclo es el test que da sentido a la
// tarea: el designmap del Documento_Referencia tiene 118 hijos intercalados y sin el
// registro de orden salían reagrupados por tipo.
func TestDocumentChildOrder_SecuenciaSobreviveElCiclo(t *testing.T) {
	data, err := os.ReadFile(corpusReferenceDesignmap)
	if err != nil {
		t.Skipf("designmap del Documento_Referencia ausente: %v (ruta esperada: %s)", err, corpusReferenceDesignmap)
	}

	// Se usa la vía con metadatos, que es la que emplea el paquete para leer y
	// escribir un designmap real: MarshalDocument por sí solo no restaura el prefijo
	// idPkg: de las referencias a recursos.
	doc, err := ParseDocumentWithMetadata(data)
	if err != nil {
		t.Fatalf("ParseDocumentWithMetadata falló: %v", err)
	}

	out, err := MarshalDocumentWithMetadata(doc)
	if err != nil {
		t.Fatalf("MarshalDocumentWithMetadata falló: %v", err)
	}

	want := childTagSequence(t, data)
	got := childTagSequence(t, out)

	if len(want) != 118 {
		t.Errorf("el designmap de referencia debería tener 118 hijos y tiene %d: ¿cambió el corpus?", len(want))
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("la secuencia de hijos cambió al serializar\nesperada: %s\nobtenida: %s",
			strings.Join(want, " "), strings.Join(got, " "))
	}
}

// TestDocumentChildOrder_DocumentoDesdeCeroUsaOrdenDeCampos comprueba que un
// documento que nunca se parseó sigue emitiendo en el orden de los campos, que es el
// comportamiento que tenía la librería antes del registro.
func TestDocumentChildOrder_DocumentoDesdeCeroUsaOrdenDeCampos(t *testing.T) {
	doc := &Document{
		DOMVersion: "20.4",
		Self:       "d",
		// A propósito en un orden que no es el de los campos: al no haber registro,
		// la salida tiene que seguir el orden de los campos y no el de asignación.
		TextVariables: []TextVariable{{Self: "tv1", Name: "var"}},
		Layers:        []Layer{{Self: "ul1", Name: "Capa"}},
	}

	if doc.childOrder.Recorded() {
		t.Fatal("un documento construido desde cero no debe tener orden registrado")
	}

	out, err := MarshalDocument(doc)
	if err != nil {
		t.Fatalf("MarshalDocument falló: %v", err)
	}

	got := childTagSequence(t, out)
	if strings.Join(got, " ") != "Layer TextVariable" {
		t.Errorf("sin registro se debe emitir en orden de campos (Layer antes de TextVariable), obtenido: %v", got)
	}
}

// TestDocumentChildOrder_HijoAgregadoTrasParsearNoSePierde cubre el techo declarado
// del registro: un hijo agregado sin actualizarlo se emite al final de su clase, pero
// se emite.
func TestDocumentChildOrder_HijoAgregadoTrasParsearNoSePierde(t *testing.T) {
	data, err := os.ReadFile(corpusReferenceDesignmap)
	if err != nil {
		t.Skipf("designmap del Documento_Referencia ausente: %v (ruta esperada: %s)", err, corpusReferenceDesignmap)
	}

	doc, err := ParseDocumentWithMetadata(data)
	if err != nil {
		t.Fatalf("ParseDocumentWithMetadata falló: %v", err)
	}

	before := len(childTagSequence(t, data))
	doc.Layers = append(doc.Layers, Layer{Self: "uNueva", Name: "Capa nueva"})

	out, err := MarshalDocumentWithMetadata(doc)
	if err != nil {
		t.Fatalf("MarshalDocumentWithMetadata falló: %v", err)
	}

	got := childTagSequence(t, out)
	if len(got) != before+1 {
		t.Fatalf("se emitieron %d hijos y se esperaban %d: la capa agregada no puede perderse", len(got), before+1)
	}
	if !strings.Contains(string(out), `Self="uNueva"`) {
		t.Error("la capa agregada tras parsear no aparece en la salida")
	}
}
