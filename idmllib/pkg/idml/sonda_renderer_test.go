package idml

// Sonda temporal: ¿se puede construir hoy un IDML con encabezado/folio usando solo
// lo que ya está cerrado (Tareas 3, 6b, 17 y el Append de la 8), sin las Tareas
// 13-16 ni 18-19?
//
// Mide tres cosas que el plan atribuye a tareas abiertas:
//   1. Agregar un TextFrame de folio a un spread generado desde plantilla.
//   2. Registrar su story en designmap.xml y en StoryList (Tarea 18, criterios 2 y 3).
//   3. Emitir el marcador de número de página automático `<?ACE 18?>` (Tarea 6b).

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/idml/idgen"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

func TestSonda_RendererConEncabezado(t *testing.T) {
	// 1. Documento base: A4, márgenes de 10mm (28.35pt). La Tarea 3 verificó que
	// esta salida abre en InDesign y en Affinity Publisher.
	opts := DefaultTemplateOptions()
	opts.Preset = PresetA4
	opts.Margins.Top = 28.35
	opts.Margins.Bottom = 28.35
	opts.Margins.Left = 28.35
	opts.Margins.Right = 28.35
	opts.ColumnCount = 5

	pkg, err := NewFromTemplate(opts)
	if err != nil {
		t.Fatalf("NewFromTemplate falló: %v", err)
	}

	// 2. IDs sin colisión con lo que la plantilla ya trae (Tarea 17, cerrada).
	reg := idgen.New()
	for _, id := range []string{"ud3", "ud8", "uf3", "ue1", "ub4"} {
		if err := reg.Register(id); err != nil {
			t.Fatalf("Register(%q) falló: %v", id, err)
		}
	}
	folioFrameID := reg.Generate()
	folioStoryID := reg.Generate()

	// 3. La story del folio: el marcador `<?ACE 18?>` va en Raw, porque Text no puede
	// representar una instrucción de proceso. Text vacío es el texto que se extrae de
	// ese Raw, y la coherencia entre ambos es lo que hace que MarshalXML reemita la
	// instrucción en lugar de descartarla.
	folioStory := &story.Story{
		DOMVersion: "20.4",
		StoryElement: story.StoryElement{
			Self:     folioStoryID,
			UserText: "true",
			ParagraphStyleRanges: []story.ParagraphStyleRange{{
				AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
				CharacterStyleRanges: []story.CharacterStyleRange{{
					AppliedCharacterStyle: "CharacterStyle/$ID/[No character style]",
					Children: []story.CharacterChild{{
						Content: &story.Content{Raw: "<?ACE 18?>"},
					}},
				}},
			}},
		},
	}

	storyPath := "Stories/Story_" + folioStoryID + ".xml"
	if err := pkg.AddStory(storyPath, folioStory, ValidationOptions{}); err != nil {
		t.Fatalf("AddStory falló: %v", err)
	}

	// 4. El marco del folio, al pie del área de contenido.
	// El origen vertical del spread está en el centro de la página, de ahí el
	// desplazamiento de media altura que ya aplica la plantilla.
	const (
		pageH   = 841.89
		pageW   = 595.276
		margen  = 28.35
		folioH  = 14.0
		utilW   = pageW - 2*margen
		centroX = margen + utilW/2
		// Pie: justo por encima del margen inferior.
		centroY = pageH - margen - folioH/2 - pageH/2
	)

	folioFrame := &spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{
			Self:          folioFrameID,
			Name:          "folio",
			Visible:       "true",
			ItemLayer:     "uba",
			ItemTransform: "1 0 0 1 " + num(centroX) + " " + num(centroY),
		},
		ParentStory:        folioStoryID,
		PreviousTextFrame:  "n",
		NextTextFrame:      "n",
		ContentType:        "TextType",
		AppliedObjectStyle: "ObjectStyle/$ID/[Normal Text Frame]",
		Properties: &common.Properties{
			PathGeometry: &common.PathGeometry{
				GeometryPathType: &common.GeometryPathType{
					PathOpen: "false",
					PathPointArray: &common.PathPointArray{
						PathPoints: puntosDeCaja(utilW/2, folioH/2),
					},
				},
			},
		},
		// TextFramePreference no está tipado todavía (Tarea 13, abierta), pero el
		// comodín OtherElements lo acepta como XML crudo y lo emite literal.
		OtherElements: []common.RawXMLElement{{
			XMLName: xml.Name{Local: "TextFramePreference"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "VerticalJustification"}, Value: "BottomAlign"},
				{Name: xml.Name{Local: "TextColumnCount"}, Value: "1"},
			},
		}},
	}

	if err := pkg.AddTextFrame(PathSpread, folioFrame, ValidationOptions{}); err != nil {
		t.Fatalf("AddTextFrame falló: %v", err)
	}

	// 5. Registrar la story en designmap.xml y en StoryList. Esto es lo que la
	// Tarea 18 va a envolver en Add*/Remove*; hoy hay que hacerlo a mano.
	doc, err := pkg.Document()
	if err != nil {
		t.Fatalf("Document() falló: %v", err)
	}
	doc.Stories = append(doc.Stories, document.ResourceRef{
		XMLName: xml.Name{
			Space: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
			Local: "Story",
		},
		Src: storyPath,
	})
	doc.StoryList = strings.TrimSpace(doc.StoryList + " " + folioStoryID)

	designmapBytes, err := document.MarshalDocumentWithMetadata(&document.DocumentWithMetadata{Document: doc})
	if err != nil {
		t.Fatalf("MarshalDocumentWithMetadata falló: %v", err)
	}
	pkg.setFileData(PathDesignmap, designmapBytes)
	pkg.invalidateCache(PathDesignmap)

	// 6. Escribir y comprobar que el resultado es un IDML válido.
	// Usamos WriteTo para escribir a un buffer en memoria — verifica que la nueva API
	// funciona correctamente (equivalente a escribir a stdout en un CLI).
	var buf bytes.Buffer
	if err := WriteTo(pkg, &buf); err != nil {
		t.Fatalf("WriteTo falló: %v", err)
	}
	datos := buf.Bytes()

	zr, err := zip.NewReader(bytes.NewReader(datos), int64(len(datos)))
	if err != nil {
		t.Fatalf("el resultado no es un ZIP legible: %v", err)
	}

	if zr.File[0].Name != "mimetype" {
		t.Errorf("primera entrada = %q, se esperaba mimetype", zr.File[0].Name)
	}
	if zr.File[0].Method != zip.Store {
		t.Error("mimetype tiene que ir sin comprimir")
	}

	contenido := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("abrir %s: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("leer %s: %v", f.Name, err)
		}
		contenido[f.Name] = string(b)
	}

	// El marcador de número automático sobrevivió.
	if got := contenido[storyPath]; !strings.Contains(got, "<?ACE 18?>") {
		t.Errorf("la story del folio no contiene <?ACE 18?>:\n%s", got)
	}

	// El marco está en el spread, con su nombre y su preferencia cruda.
	sp := contenido[PathSpread]
	for _, quiero := range []string{
		`Self="` + folioFrameID + `"`,
		`Name="folio"`,
		`ParentStory="` + folioStoryID + `"`,
		`VerticalJustification="BottomAlign"`,
	} {
		if !strings.Contains(sp, quiero) {
			t.Errorf("el spread no contiene %s", quiero)
		}
	}

	// La story quedó referenciada en el designmap y en StoryList.
	dm := contenido[PathDesignmap]
	if !strings.Contains(dm, storyPath) {
		t.Errorf("designmap no referencia %s", storyPath)
	}
	if !strings.Contains(dm, folioStoryID) {
		t.Errorf("designmap/StoryList no contiene %s", folioStoryID)
	}

	// Y el paquete se puede volver a abrir.
	reabierto, err := ReadBytes(datos)
	if err != nil {
		t.Fatalf("el IDML generado no se puede reabrir: %v", err)
	}
	st, err := reabierto.Story(storyPath)
	if err != nil {
		t.Fatalf("la story del folio no se recupera: %v", err)
	}
	if st.StoryElement.Self != folioStoryID {
		t.Errorf("Self de la story = %q, se esperaba %q", st.StoryElement.Self, folioStoryID)
	}

	t.Logf("IDML generado: %d bytes, %d entradas", len(datos), len(zr.File))
}

// puntosDeCaja devuelve los 4 puntos de un rectángulo centrado en el origen.
func puntosDeCaja(hw, hh float64) []common.PathPointType {
	esquinas := [4][2]float64{{-hw, -hh}, {-hw, hh}, {hw, hh}, {hw, -hh}}
	pts := make([]common.PathPointType, 0, 4)
	for _, c := range esquinas {
		a := num(c[0]) + " " + num(c[1])
		pts = append(pts, common.PathPointType{Anchor: a, LeftDirection: a, RightDirection: a})
	}
	return pts
}
