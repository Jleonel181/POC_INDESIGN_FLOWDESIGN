package story

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Este archivo cubre la preservación de las instrucciones de proceso dentro de
// <Content>, que son los marcadores de carácter especial de InDesign.
//
// El defecto que arregla: `Content` se modelaba solo como datos de carácter, así que
// `<Content><?ACE 18?>.</Content>` se re-emitía como `<Content>.</Content>` y el
// marcador desaparecía del documento. Era la única pérdida de **contenido de texto**
// que quedaba en el corpus.

// TestContent_InstruccionDeProcesoSobreviveElCiclo cubre las cuatro formas que aparecen
// en el Documento_Referencia, más el caso de una instrucción entre dos textos.
func TestContent_InstruccionDeProcesoSobreviveElCiclo(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		wantText string
	}{
		{nombre: "instrucción antes del texto", entrada: `<?ACE 18?>.`, wantText: "."},
		{nombre: "instrucción sola", entrada: `<?ACE 18?>`, wantText: ""},
		{nombre: "instrucción y texto con espacios finales", entrada: `<?ACE 18?>.   `, wantText: ".   "},
		{nombre: "texto antes de la instrucción", entrada: ` <?ACE 18?>`, wantText: " "},
		{nombre: "instrucción entre dos textos", entrada: `antes<?ACE 18?>después`, wantText: "antesdespués"},
	}

	for _, tt := range casos {
		t.Run(tt.nombre, func(t *testing.T) {
			entrada := []byte(envuelveEnStory(`<Content>` + tt.entrada + `</Content>`))

			s, err := ParseStory(entrada)
			if err != nil {
				t.Fatalf("ParseStory falló: %v", err)
			}

			content := primerContent(t, s)
			if content.Text != tt.wantText {
				t.Errorf("Text = %q, esperado %q: Text debe seguir siendo el texto legible, sin el marcado", content.Text, tt.wantText)
			}

			out, err := MarshalStory(s)
			if err != nil {
				t.Fatalf("MarshalStory falló: %v", err)
			}

			if !strings.Contains(string(out), `<Content>`+tt.entrada+`</Content>`) {
				t.Errorf("el contenido no sobrevivió el ciclo.\nesperado contener: %s\nsalida:\n%s",
					`<Content>`+tt.entrada+`</Content>`, recorta(string(out)))
			}
		})
	}
}

// TestContent_TextoNormalNoCambia es el criterio de no-regresión: el camino del texto
// sin marcado tiene que seguir escapándose exactamente igual. Se toca el contenido de
// las 30 stories, que es lo más visible del documento.
func TestContent_TextoNormalNoCambia(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
		want    string
	}{
		{nombre: "acentos", entrada: `Miércoles, 15 de julio`, want: `Miércoles, 15 de julio`},
		{nombre: "ampersand", entrada: `pan &amp; circo`, want: `pan &amp; circo`},
		{nombre: "menor que", entrada: `3 &lt; 4`, want: `3 &lt; 4`},
		{nombre: "comillas y apóstrofos", entrada: `«cita» y apóstrofo’`, want: `«cita» y apóstrofo’`},
		{nombre: "vacío", entrada: ``, want: ``},
		{nombre: "solo espacios", entrada: `   `, want: `   `},
	}

	for _, tt := range casos {
		t.Run(tt.nombre, func(t *testing.T) {
			s, err := ParseStory([]byte(envuelveEnStory(`<Content>` + tt.entrada + `</Content>`)))
			if err != nil {
				t.Fatalf("ParseStory falló: %v", err)
			}
			out, err := MarshalStory(s)
			if err != nil {
				t.Fatalf("MarshalStory falló: %v", err)
			}
			if !strings.Contains(string(out), `<Content>`+tt.want+`</Content>`) {
				t.Errorf("el escapado cambió.\nesperado contener: %s\nsalida:\n%s",
					`<Content>`+tt.want+`</Content>`, recorta(string(out)))
			}
		})
	}
}

// TestContent_SiElTextoSeModificaGanaElTexto es la parte que evita un fallo peor que el
// original. Si se emitiera siempre el contenido literal, un llamador que cambie el
// texto vería su cambio descartado **en silencio**.
func TestContent_SiElTextoSeModificaGanaElTexto(t *testing.T) {
	s, err := ParseStory([]byte(envuelveEnStory(`<Content><?ACE 18?>viejo</Content>`)))
	if err != nil {
		t.Fatalf("ParseStory falló: %v", err)
	}

	primerContent(t, s).Text = "nuevo"

	out, err := MarshalStory(s)
	if err != nil {
		t.Fatalf("MarshalStory falló: %v", err)
	}

	if !strings.Contains(string(out), `<Content>nuevo</Content>`) {
		t.Errorf("al modificar Text debe ganar Text.\nsalida:\n%s", recorta(string(out)))
	}
	if strings.Contains(string(out), "viejo") {
		t.Error("se emitió el texto viejo: el contenido literal se usó cuando ya no era coherente con Text")
	}
}

// TestContent_MarcadoresDelCorpusSobreviven comprueba el caso real, sobre las cuatro
// stories del Documento_Referencia que llevan un marcador.
func TestContent_MarcadoresDelCorpusSobreviven(t *testing.T) {
	stories := []string{"Story_u35d.xml", "Story_u3a2.xml", "Story_u487.xml", "Story_u4cc.xml"}

	for _, nombre := range stories {
		t.Run(nombre, func(t *testing.T) {
			ruta := filepath.Join("../../testdata/documento_referencia/Stories", nombre)
			data, err := os.ReadFile(ruta)
			if err != nil {
				t.Skipf("story ausente: %v (ruta esperada: %s)", err, ruta)
			}

			antes := strings.Count(string(data), "<?ACE")
			if antes == 0 {
				t.Fatalf("%s debería contener al menos un marcador <?ACE: ¿cambió el corpus?", nombre)
			}

			s, err := ParseStory(data)
			if err != nil {
				t.Fatalf("ParseStory falló: %v", err)
			}
			out, err := MarshalStory(s)
			if err != nil {
				t.Fatalf("MarshalStory falló: %v", err)
			}

			if despues := strings.Count(string(out), "<?ACE"); despues != antes {
				t.Errorf("marcadores <?ACE: %d en la entrada, %d en la salida", antes, despues)
			}
		})
	}
}

// envuelveEnStory pone un fragmento de contenido dentro de la estructura mínima de una
// story, para poder pasarlo por ParseStory.
func envuelveEnStory(contenido string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Story Self="ue1">
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				` + contenido + `
			</CharacterStyleRange>
		</ParagraphStyleRange>
	</Story>
</idPkg:Story>`
}

// primerContent devuelve el primer Content de la story, para no repetir el descenso.
func primerContent(t *testing.T, s *Story) *Content {
	t.Helper()

	for i := range s.StoryElement.ParagraphStyleRanges {
		psr := &s.StoryElement.ParagraphStyleRanges[i]
		for j := range psr.CharacterStyleRanges {
			for _, child := range psr.CharacterStyleRanges[j].Children {
				if child.Content != nil {
					return child.Content
				}
			}
		}
	}
	t.Fatal("la story no contiene ningún Content")
	return nil
}

func recorta(s string) string {
	if len(s) > 600 {
		return s[:600] + "..."
	}
	return s
}
