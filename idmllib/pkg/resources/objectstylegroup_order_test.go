package resources

import (
	"bytes"
	"encoding/xml"
	"testing"
)

// hijosDeGrupo devuelve los nombres de los hijos directos del primer elemento con el nombre
// local dado.
func hijosDeGrupo(t *testing.T, data []byte, local string) []string {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(data))
	dentro, prof := false, 0
	var out []string
	for {
		tok, err := d.Token()
		if err != nil {
			return out
		}
		switch e := tok.(type) {
		case xml.StartElement:
			if !dentro {
				if e.Name.Local == local {
					dentro, prof = true, 0
				}
				continue
			}
			if prof == 0 {
				out = append(out, e.Name.Local)
			}
			prof++
		case xml.EndElement:
			if !dentro {
				continue
			}
			if prof == 0 {
				return out
			}
			prof--
		}
	}
}

func igual(a, b []string) bool {
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

// TestObjectStyleGroup_ConservaElOrden es el criterio 2 de la Tarea 2e. Las dos primeras
// disposiciones están medidas en el corpus, y son las que demuestran que ningún orden fijo
// de los campos sirve para las dos.
func TestObjectStyleGroup_ConservaElOrden(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado []string
	}{
		{
			// La que está mal hoy: el grupo anidado delante de los estilos. Medida 2 veces.
			nombre: "grupo anidado antes de los estilos",
			entrada: `<RootObjectStyleGroup Self="u8a">` +
				`<ObjectStyleGroup Self="g1" Name="grupo"/>` +
				`<ObjectStyle Self="s1" Name="uno"/>` +
				`<ObjectStyle Self="s2" Name="dos"/>` +
				`</RootObjectStyleGroup>`,
			esperado: []string{"ObjectStyleGroup", "ObjectStyle", "ObjectStyle"},
		},
		{
			// La que ya salía bien porque coincide con el orden de los campos. Medida 2
			// veces. Tiene que seguir saliendo bien.
			nombre: "estilo antes del grupo anidado",
			entrada: `<RootObjectStyleGroup Self="u8a">` +
				`<ObjectStyle Self="s1" Name="uno"/>` +
				`<ObjectStyleGroup Self="g1" Name="grupo"/>` +
				`</RootObjectStyleGroup>`,
			esperado: []string{"ObjectStyle", "ObjectStyleGroup"},
		},
		{
			nombre: "intercalado",
			entrada: `<RootObjectStyleGroup Self="u8a">` +
				`<ObjectStyle Self="s1" Name="uno"/>` +
				`<ObjectStyleGroup Self="g1" Name="g1"/>` +
				`<ObjectStyle Self="s2" Name="dos"/>` +
				`<ObjectStyleGroup Self="g2" Name="g2"/>` +
				`</RootObjectStyleGroup>`,
			esperado: []string{"ObjectStyle", "ObjectStyleGroup", "ObjectStyle", "ObjectStyleGroup"},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			var g ObjectStyleGroup
			if err := xml.Unmarshal([]byte(c.entrada), &g); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got := g.ChildTags(); !igual(got, c.esperado) {
				t.Errorf("orden registrado: esperado %v, obtenido %v", c.esperado, got)
			}

			salida, err := xml.Marshal(&g)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if got := hijosDeGrupo(t, salida, "RootObjectStyleGroup"); !igual(got, c.esperado) {
				t.Errorf("orden emitido: esperado %v, obtenido %v\n%s", c.esperado, got, salida)
			}
		})
	}
}

// TestObjectStyleGroup_ConservaElNombreDeLaRaiz es la guarda de la trampa de encoding/xml:
// ignora el campo XMLName en un tipo que implementa Marshaler y cae en el nombre del tipo de
// Go, que aquí es «ObjectStyleGroup». Si eso pasara, la raíz se emitiría con el nombre del
// anidado y el archivo de estilos quedaría irreconocible.
func TestObjectStyleGroup_ConservaElNombreDeLaRaiz(t *testing.T) {
	entrada := `<RootObjectStyleGroup Self="u8a"><ObjectStyle Self="s1" Name="uno"/></RootObjectStyleGroup>`

	var g ObjectStyleGroup
	if err := xml.Unmarshal([]byte(entrada), &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g.XMLName.Local != "RootObjectStyleGroup" {
		t.Fatalf("XMLName debería ser RootObjectStyleGroup, es %q", g.XMLName.Local)
	}

	salida, err := xml.Marshal(&g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(salida, []byte("<RootObjectStyleGroup")) {
		t.Errorf("la raíz debería emitirse como <RootObjectStyleGroup>: %s", salida)
	}
	if bytes.HasPrefix(salida, []byte("<ObjectStyleGroup")) {
		t.Errorf("la raíz se emitió con el nombre del anidado: %s", salida)
	}
}

// TestObjectStyleGroup_OrdenEnCadaNivelDeAnidamiento es la parte del criterio 2 que dice «y
// lo hace en cada nivel de anidamiento».
func TestObjectStyleGroup_OrdenEnCadaNivelDeAnidamiento(t *testing.T) {
	entrada := `<RootObjectStyleGroup Self="r">` +
		`<ObjectStyleGroup Self="n1" Name="nivel1">` +
		`<ObjectStyleGroup Self="n2" Name="nivel2">` +
		`<ObjectStyle Self="s3" Name="tres"/>` +
		`</ObjectStyleGroup>` +
		`<ObjectStyle Self="s2" Name="dos"/>` +
		`</ObjectStyleGroup>` +
		`<ObjectStyle Self="s1" Name="uno"/>` +
		`</RootObjectStyleGroup>`

	var g ObjectStyleGroup
	if err := xml.Unmarshal([]byte(entrada), &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Nivel 1: grupo anidado y luego estilo.
	if got, esp := g.ChildTags(), []string{"ObjectStyleGroup", "ObjectStyle"}; !igual(got, esp) {
		t.Errorf("raíz: esperado %v, obtenido %v", esp, got)
	}
	// Nivel 2: dentro del anidado, otro grupo y luego estilo.
	if len(g.NestedGroups) != 1 {
		t.Fatalf("se esperaba 1 grupo anidado, hay %d", len(g.NestedGroups))
	}
	n1 := &g.NestedGroups[0]
	if got, esp := n1.ChildTags(), []string{"ObjectStyleGroup", "ObjectStyle"}; !igual(got, esp) {
		t.Errorf("nivel 1: esperado %v, obtenido %v", esp, got)
	}

	// Y el XML emitido conserva las dos secuencias.
	salida, err := xml.Marshal(&g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got, esp := hijosDeGrupo(t, salida, "RootObjectStyleGroup"), []string{"ObjectStyleGroup", "ObjectStyle"}; !igual(got, esp) {
		t.Errorf("emitido en la raíz: esperado %v, obtenido %v\n%s", esp, got, salida)
	}
	// El primer <ObjectStyleGroup> de la salida es el nivel 1.
	if got, esp := hijosDeGrupo(t, salida, "ObjectStyleGroup"), []string{"ObjectStyleGroup", "ObjectStyle"}; !igual(got, esp) {
		t.Errorf("emitido en el nivel 1: esperado %v, obtenido %v\n%s", esp, got, salida)
	}
	// Y los tres estilos siguen ahí.
	for _, self := range []string{`Self="s1"`, `Self="s2"`, `Self="s3"`} {
		if !bytes.Contains(salida, []byte(self)) {
			t.Errorf("se perdió el estilo %s: %s", self, salida)
		}
	}
}

// TestObjectStyleGroup_DesdeCeroEmiteEnOrdenDeCampos comprueba que un grupo construido a
// mano sigue emitiendo estilos antes que grupos anidados, el comportamiento anterior.
func TestObjectStyleGroup_DesdeCeroEmiteEnOrdenDeCampos(t *testing.T) {
	g := ObjectStyleGroup{
		Self:         "u8a",
		NestedGroups: []ObjectStyleGroup{{Self: "g1", Name: "grupo"}},
		ObjectStyles: []ObjectStyle{{Self: "s1", Name: "uno"}},
	}
	salida, err := xml.Marshal(&g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Sin XMLName ni nombre explícito, el nombre cae en el del tipo de Go, que para un
	// grupo construido a mano es el del anidado. Es correcto: quien quiera la raíz usa
	// EncodeElement con el nombre, que es lo que hace StylesFile.
	esperado := []string{"ObjectStyle", "ObjectStyleGroup"}
	if got := hijosDeGrupo(t, salida, "ObjectStyleGroup"); !igual(got, esperado) {
		t.Errorf("esperado %v, obtenido %v\n%s", esperado, got, salida)
	}
}

// TestObjectStyleGroup_ConservaAtributos comprueba que el reparto de atributos sigue
// funcionando tras darle lectura propia al tipo, incluidos los que el modelo no declara.
func TestObjectStyleGroup_ConservaAtributos(t *testing.T) {
	entrada := `<RootObjectStyleGroup Self="u8a" Name="raiz" NoDeclarado="valor"/>`

	var g ObjectStyleGroup
	if err := xml.Unmarshal([]byte(entrada), &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g.Self != "u8a" || g.Name != "raiz" {
		t.Errorf("los atributos declarados no se leyeron: Self=%q Name=%q", g.Self, g.Name)
	}
	if len(g.OtherAttrs) != 1 || g.OtherAttrs[0].Name.Local != "NoDeclarado" {
		t.Errorf("el atributo no declarado no llegó al comodín: %v", g.OtherAttrs)
	}

	salida, err := xml.Marshal(&g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, esp := range []string{`Self="u8a"`, `Name="raiz"`, `NoDeclarado="valor"`} {
		if !bytes.Contains(salida, []byte(esp)) {
			t.Errorf("falta %s en la salida: %s", esp, salida)
		}
	}
}

// TestObjectStyleGroup_IdempotenciaAlParsearDosVeces comprueba que deserializar dos veces no
// acumula hijos ni orden.
func TestObjectStyleGroup_IdempotenciaAlParsearDosVeces(t *testing.T) {
	entrada := []byte(`<RootObjectStyleGroup Self="u8a"><ObjectStyleGroup Self="g1"/><ObjectStyle Self="s1" Name="uno"/></RootObjectStyleGroup>`)

	var g ObjectStyleGroup
	if err := xml.Unmarshal(entrada, &g); err != nil {
		t.Fatalf("primera: %v", err)
	}
	n1 := len(g.ChildTags())

	if err := xml.Unmarshal(entrada, &g); err != nil {
		t.Fatalf("segunda: %v", err)
	}
	if got := len(g.ChildTags()); got != n1 {
		t.Errorf("el orden registrado pasó de %d a %d entradas: se está acumulando", n1, got)
	}
	if len(g.ObjectStyles) != 1 || len(g.NestedGroups) != 1 {
		t.Errorf("los hijos se acumularon: %d estilos, %d grupos", len(g.ObjectStyles), len(g.NestedGroups))
	}
}
