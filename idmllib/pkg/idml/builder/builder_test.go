package builder_test

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/idml/builder"
)

func TestBuilder_SinglePage(t *testing.T) {
	// A4 en puntos
	doc := builder.NewDocument(builder.DocumentOptions{
		Width:  595.28,
		Height: 841.89,
		Margins: builder.Margins{Top: 36, Bottom: 36, Left: 36, Right: 36},
		Columns: 1,
	})
	doc.AddPage(builder.PageOptions{
		Frames: []builder.FrameOptions{{
			Type:    "text",
			Name:    "Titulo",
			Bounds:  builder.Bounds{TopMm: 20, LeftMm: 20, BottomMm: 50, RightMm: 190},
			Content: "Hola Mundo",
		}},
	})

	pkg, err := doc.Build()
	if err != nil {
		t.Fatalf("Build() falló: %v", err)
	}

	// Criterio 7: tiene las entradas mínimas
	requiredPrefixes := []string{"mimetype", "META-INF/", "designmap.xml", "Resources/", "XML/", "MasterSpreads/", "Spreads/", "Stories/"}
	files := pkg.Files()
	for _, prefix := range requiredPrefixes {
		found := false
		for _, f := range files {
			if f == prefix || len(f) >= len(prefix) && f[:len(prefix)] == prefix {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("falta entrada con prefijo %q", prefix)
		}
	}

	// Criterio 8: roundtrip — re-abrir produce 0 diferencias
	var buf bytes.Buffer
	if err := idml.WriteTo(pkg, &buf); err != nil {
		t.Fatalf("WriteTo falló: %v", err)
	}

	pkg2, err := idml.ReadBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ReadBytes falló: %v", err)
	}

	doc2, err := pkg2.Document()
	if err != nil {
		t.Fatalf("Document() falló: %v", err)
	}

	if len(doc2.Spreads) == 0 {
		t.Error("no hay spreads en el roundtrip")
	}
	if len(doc2.Stories) == 0 {
		t.Error("no hay stories en el roundtrip")
	}

	// Verificar que el texto llegó
	for _, ref := range doc2.Stories {
		st, err := pkg2.Story(ref.Src)
		if err != nil {
			continue
		}
		text := st.ExtractText()
		if text == "Hola Mundo" {
			return // éxito
		}
	}
	t.Error("no se encontró el texto 'Hola Mundo' en ninguna story")
}

func TestBuilder_FacingPages(t *testing.T) {
	doc := builder.NewDocument(builder.DocumentOptions{
		Width:       595.28,
		Height:      841.89,
		Margins:     builder.Margins{Top: 36, Bottom: 36, Left: 36, Right: 36},
		FacingPages: true,
	})
	doc.AddPage(builder.PageOptions{
		Frames: []builder.FrameOptions{{Type: "text", Bounds: builder.Bounds{TopMm: 20, LeftMm: 20, BottomMm: 50, RightMm: 190}, Content: "P1"}},
	})
	doc.AddPage(builder.PageOptions{
		Frames: []builder.FrameOptions{{Type: "text", Bounds: builder.Bounds{TopMm: 20, LeftMm: 20, BottomMm: 50, RightMm: 190}, Content: "P2"}},
	})
	doc.AddPage(builder.PageOptions{
		Frames: []builder.FrameOptions{{Type: "text", Bounds: builder.Bounds{TopMm: 20, LeftMm: 20, BottomMm: 50, RightMm: 190}, Content: "P3"}},
	})

	pkg, err := doc.Build()
	if err != nil {
		t.Fatalf("Build() falló: %v", err)
	}

	// Facing pages + 3 páginas → 2 spreads: [p1] [p2,p3]
	dm, _ := pkg.Document()
	if got := len(dm.Spreads); got != 2 {
		t.Errorf("spreads: esperado 2, obtenido %d", got)
	}
}

func TestBuilder_InvalidOptions(t *testing.T) {
	// Criterio 9: error si opciones obligatorias faltan
	doc := builder.NewDocument(builder.DocumentOptions{
		Width:  0, // inválido
		Height: 841.89,
	})
	doc.AddPage(builder.PageOptions{})

	_, err := doc.Build()
	if err == nil {
		t.Error("esperaba error con Width=0")
	}
}

func TestBuilder_ZipMimetype(t *testing.T) {
	doc := builder.NewDocument(builder.DocumentOptions{
		Width:  595.28,
		Height: 841.89,
		Margins: builder.Margins{Top: 36, Bottom: 36, Left: 36, Right: 36},
	})
	doc.AddPage(builder.PageOptions{
		Frames: []builder.FrameOptions{{Type: "text", Bounds: builder.Bounds{TopMm: 20, LeftMm: 20, BottomMm: 50, RightMm: 190}, Content: "X"}},
	})

	pkg, err := doc.Build()
	if err != nil {
		t.Fatalf("Build() falló: %v", err)
	}

	var buf bytes.Buffer
	if err := idml.WriteTo(pkg, &buf); err != nil {
		t.Fatalf("WriteTo falló: %v", err)
	}

	// Verificar ZIP: mimetype primera entrada sin compresión
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader falló: %v", err)
	}
	if len(r.File) == 0 {
		t.Fatal("ZIP vacío")
	}
	if r.File[0].Name != "mimetype" {
		t.Errorf("primera entrada: %q, esperado 'mimetype'", r.File[0].Name)
	}
	if r.File[0].Method != zip.Store {
		t.Errorf("mimetype method: %d, esperado Store (0)", r.File[0].Method)
	}
}
