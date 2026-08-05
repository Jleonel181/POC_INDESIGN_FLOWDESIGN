package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/internal/idmlgen"
	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
)

// TestGenerate_FacingPages3P verifica que Generate() produce un IDML coherente
// a partir del fixture facing_pages_3p.json. Es la red de seguridad antes de
// refactorizar: si este test falla tras un cambio, algo se rompió.
func TestGenerate_FacingPages3P(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "facing_pages_3p.json"))
	if err != nil {
		t.Fatalf("no se pudo leer fixture: %v", err)
	}

	var input idmlgen.DocumentInput
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}

	if err := idmlgen.Validate(&input); err != nil {
		t.Fatalf("validación falló: %v", err)
	}

	pkg, err := idmlgen.Generate(&input)
	if err != nil {
		t.Fatalf("generate falló: %v", err)
	}

	// --- Verificar estructura del documento ---

	doc, err := pkg.Document()
	if err != nil {
		t.Fatalf("no se pudo leer designmap: %v", err)
	}

	// Facing pages + 3 páginas → 2 spreads: [p1] [p2,p3]
	if got := len(doc.Spreads); got != 2 {
		t.Errorf("spreads: esperado 2, obtenido %d", got)
	}

	// --- Verificar stories ---
	// 4 frames en total → 4 stories (una por frame)
	// El fixture tiene: 1 (p1) + 2 (p2) + 1 (p3) = 4 frames
	expectedStoryRefs := 4
	if got := len(doc.Stories); got != expectedStoryRefs {
		t.Errorf("stories en designmap: esperado %d, obtenido %d", expectedStoryRefs, got)
	}

	// StoryList incluye las 4 stories + u98 (BackingStory del template)
	storyIDs := strings.Fields(doc.StoryList)
	if got := len(storyIDs); got != expectedStoryRefs+1 {
		t.Errorf("StoryList IDs: esperado %d (4 + BackingStory), obtenido %d", expectedStoryRefs+1, got)
	}

	// --- Verificar contenido de las stories ---
	expectedContents := []string{"Cabezote", "Pauta Deportes", "Pauta Publicidad", "Pauta Cultura"}
	foundContents := make(map[string]bool)

	for _, ref := range doc.Stories {
		st, err := pkg.Story(ref.Src)
		if err != nil {
			t.Errorf("no se pudo leer story %s: %v", ref.Src, err)
			continue
		}
		for _, psr := range st.StoryElement.ParagraphStyleRanges {
			for _, csr := range psr.CharacterStyleRanges {
				for _, child := range csr.Children {
					if child.Content != nil {
						text := child.Content.Text
						if text == "" {
							text = child.Content.Raw
						}
						foundContents[text] = true
					}
				}
			}
		}
	}

	for _, expected := range expectedContents {
		if !foundContents[expected] {
			t.Errorf("contenido %q no encontrado en ninguna story", expected)
		}
	}

	// --- Verificar TextFrames por spread ---
	// Spread 1 (portada): 1 frame
	// Spread 2 (par p2+p3): 3 frames (2 de p2, 1 de p3)

	for i, ref := range doc.Spreads {
		sp, err := pkg.Spread(ref.Src)
		if err != nil {
			t.Errorf("no se pudo leer spread %s: %v", ref.Src, err)
			continue
		}

		frameCount := len(sp.TextFrames())

		switch i {
		case 0:
			if frameCount != 1 {
				t.Errorf("spread[0] frames: esperado 1, obtenido %d", frameCount)
			}
		case 1:
			if frameCount != 3 {
				t.Errorf("spread[1] frames: esperado 3, obtenido %d", frameCount)
			}
		}
	}

	// --- Verificar que se puede escribir y releer sin error ---
	outPath := filepath.Join(t.TempDir(), "output.idml")
	if err := idmlpkg.Write(pkg, outPath); err != nil {
		t.Fatalf("Write falló: %v", err)
	}

	reread, err := idmlpkg.Read(outPath)
	if err != nil {
		t.Fatalf("Read del IDML generado falló: %v", err)
	}

	rereadDoc, err := reread.Document()
	if err != nil {
		t.Fatalf("designmap del reread falló: %v", err)
	}

	if got := len(rereadDoc.Spreads); got != 2 {
		t.Errorf("reread spreads: esperado 2, obtenido %d", got)
	}
	if got := len(rereadDoc.Stories); got != expectedStoryRefs {
		t.Errorf("reread stories: esperado %d, obtenido %d", expectedStoryRefs, got)
	}
}

// TestGenerate_SinglePage verifica el camino sin facing pages (cada página = 1 spread).
func TestGenerate_SinglePage(t *testing.T) {
	input := idmlgen.DocumentInput{
		Document: idmlgen.DocumentSpec{
			WidthMm:     210,
			HeightMm:    297,
			Margins:     idmlgen.MarginsSpec{Top: 10, Bottom: 10, Left: 10, Right: 10},
			FacingPages: false,
			Columns:     1,
		},
		Pages: []idmlgen.PageSpec{
			{
				Frames: []idmlgen.FrameSpec{
					{
						Type:    "text",
						Name:    "Marco único",
						Bounds:  idmlgen.BoundsSpec{TopMm: 10, LeftMm: 10, BottomMm: 287, RightMm: 200},
						Content: "Contenido de prueba",
					},
				},
			},
		},
	}

	if err := idmlgen.Validate(&input); err != nil {
		t.Fatalf("validación falló: %v", err)
	}

	pkg, err := idmlgen.Generate(&input)
	if err != nil {
		t.Fatalf("generate falló: %v", err)
	}

	doc, err := pkg.Document()
	if err != nil {
		t.Fatalf("designmap falló: %v", err)
	}

	// 1 página, sin facing → 1 spread
	if got := len(doc.Spreads); got != 1 {
		t.Errorf("spreads: esperado 1, obtenido %d", got)
	}

	// 1 story
	if got := len(doc.Stories); got != 1 {
		t.Errorf("stories: esperado 1, obtenido %d", got)
	}
}

// TestValidate_Errors verifica que Validate rechaza inputs inválidos.
func TestValidate_Errors(t *testing.T) {
	cases := []struct {
		name  string
		input idmlgen.DocumentInput
	}{
		{
			name: "ancho cero",
			input: idmlgen.DocumentInput{
				Document: idmlgen.DocumentSpec{WidthMm: 0, HeightMm: 297},
				Pages:    []idmlgen.PageSpec{{Frames: []idmlgen.FrameSpec{{Type: "text", Bounds: idmlgen.BoundsSpec{TopMm: 0, LeftMm: 0, BottomMm: 10, RightMm: 10}}}}},
			},
		},
		{
			name: "alto cero",
			input: idmlgen.DocumentInput{
				Document: idmlgen.DocumentSpec{WidthMm: 210, HeightMm: 0},
				Pages:    []idmlgen.PageSpec{{Frames: []idmlgen.FrameSpec{{Type: "text", Bounds: idmlgen.BoundsSpec{TopMm: 0, LeftMm: 0, BottomMm: 10, RightMm: 10}}}}},
			},
		},
		{
			name: "sin páginas",
			input: idmlgen.DocumentInput{
				Document: idmlgen.DocumentSpec{WidthMm: 210, HeightMm: 297},
				Pages:    []idmlgen.PageSpec{},
			},
		},
		{
			name: "frame sin tipo",
			input: idmlgen.DocumentInput{
				Document: idmlgen.DocumentSpec{WidthMm: 210, HeightMm: 297},
				Pages:    []idmlgen.PageSpec{{Frames: []idmlgen.FrameSpec{{Bounds: idmlgen.BoundsSpec{TopMm: 0, LeftMm: 0, BottomMm: 10, RightMm: 10}}}}},
			},
		},
		{
			name: "bounds invertidos vertical",
			input: idmlgen.DocumentInput{
				Document: idmlgen.DocumentSpec{WidthMm: 210, HeightMm: 297},
				Pages:    []idmlgen.PageSpec{{Frames: []idmlgen.FrameSpec{{Type: "text", Bounds: idmlgen.BoundsSpec{TopMm: 50, LeftMm: 0, BottomMm: 10, RightMm: 10}}}}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := idmlgen.Validate(&tc.input); err == nil {
				t.Errorf("esperaba error de validación para %q, no obtuvo ninguno", tc.name)
			}
		})
	}
}
