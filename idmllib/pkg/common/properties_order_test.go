package common

import (
	"bytes"
	"encoding/xml"
	"testing"
)

// hijosDirectos devuelve los nombres de los hijos directos del primer elemento con el
// nombre local dado.
func hijosDirectos(t *testing.T, data []byte, local string) []string {
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

func mismaSecuencia(a, b []string) bool {
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

// TestProperties_ConservaElOrdenDeSusHijos es el criterio 1 de la Tarea 2e. Los dos primeros
// casos son los medidos en la especificación; el tercero es el que demuestra por qué
// reordenar los campos del struct no bastaba.
func TestProperties_ConservaElOrdenDeSusHijos(t *testing.T) {
	casos := []struct {
		nombre   string
		entrada  string
		esperado []string
	}{
		{
			// Properties de un Section del designmap de example.idml. Un hijo del comodín
			// y después el Label tipado.
			nombre:   "comodin antes de Label",
			entrada:  `<Properties><PageNumberStyle type="string">Arabic</PageNumberStyle><Label><KeyValuePair Key="k" Value="v"/></Label></Properties>`,
			esperado: []string{"PageNumberStyle", "Label"},
		},
		{
			// Properties de una Page de Spreads/Spread_u210.xml. Dos del comodín y después
			// el Label.
			nombre:   "dos comodines antes de Label",
			entrada:  `<Properties><PageColor type="enumeration">UseMasterColor</PageColor><Descriptor type="list"><ListItem type="string">a</ListItem></Descriptor><Label><KeyValuePair Key="k" Value="v"/></Label></Properties>`,
			esperado: []string{"PageColor", "Descriptor", "Label"},
		},
		{
			// La disposición contraria, del designmap: el Label tipado **antes** de un hijo
			// del comodín. Es la que se rompería si en lugar del registro de orden se
			// hubiera puesto el comodín primero en el struct. Medida 5 veces en el corpus.
			nombre:   "Label antes del comodin",
			entrada:  `<Properties><Label><KeyValuePair Key="k" Value="v"/></Label><AppliedMathMLSwatch type="object">n</AppliedMathMLSwatch></Properties>`,
			esperado: []string{"Label", "AppliedMathMLSwatch"},
		},
		{
			nombre:   "PathGeometry y Label, orden de los campos",
			entrada:  `<Properties><PathGeometry><GeometryPathType PathOpen="false"/></PathGeometry><Label><KeyValuePair Key="k" Value="v"/></Label></Properties>`,
			esperado: []string{"PathGeometry", "Label"},
		},
		{
			nombre:   "PathGeometry despues de un comodin",
			entrada:  `<Properties><PageColor type="enumeration">X</PageColor><PathGeometry><GeometryPathType PathOpen="false"/></PathGeometry></Properties>`,
			esperado: []string{"PageColor", "PathGeometry"},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			var p Properties
			if err := xml.Unmarshal([]byte(c.entrada), &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got := p.ChildTags(); !mismaSecuencia(got, c.esperado) {
				t.Errorf("orden registrado al parsear: esperado %v, obtenido %v", c.esperado, got)
			}

			salida, err := xml.Marshal(&p)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if got := hijosDirectos(t, salida, "Properties"); !mismaSecuencia(got, c.esperado) {
				t.Errorf("orden emitido: esperado %v, obtenido %v\n%s", c.esperado, got, salida)
			}
		})
	}
}

// TestProperties_ElOrdenNoEsElDeLosCampos es el control negativo. Sin él, un modelo que
// reagrupara podría pasar los casos de arriba por casualidad.
func TestProperties_ElOrdenNoEsElDeLosCampos(t *testing.T) {
	// Los campos se declaran PathGeometry, Label, OtherElements. Esta entrada trae el orden
	// contrario del todo.
	entrada := `<Properties><PageColor type="enumeration">X</PageColor><Label><KeyValuePair Key="k" Value="v"/></Label><PathGeometry><GeometryPathType PathOpen="false"/></PathGeometry></Properties>`
	esperado := []string{"PageColor", "Label", "PathGeometry"}
	porCampos := []string{"PathGeometry", "Label", "PageColor"}

	var p Properties
	if err := xml.Unmarshal([]byte(entrada), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	salida, err := xml.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := hijosDirectos(t, salida, "Properties")

	if mismaSecuencia(got, porCampos) {
		t.Errorf("se emitió en el orden de los campos %v: el registro de orden no se está usando", porCampos)
	}
	if !mismaSecuencia(got, esperado) {
		t.Errorf("orden emitido: esperado %v, obtenido %v", esperado, got)
	}
}

// TestProperties_DesdeCeroEmiteEnOrdenDeCampos comprueba que un Properties que nunca se
// parseó sigue emitiendo en el orden de los campos, el comportamiento anterior a la tarea.
func TestProperties_DesdeCeroEmiteEnOrdenDeCampos(t *testing.T) {
	p := Properties{
		Label:        &Label{KeyValuePairs: []KeyValuePair{{Key: "k", Value: "v"}}},
		PathGeometry: &PathGeometry{GeometryPathType: &GeometryPathType{PathOpen: "false"}},
	}
	salida, err := xml.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	esperado := []string{"PathGeometry", "Label"}
	if got := hijosDirectos(t, salida, "Properties"); !mismaSecuencia(got, esperado) {
		t.Errorf("esperado %v, obtenido %v\n%s", esperado, got, salida)
	}
}

// TestProperties_ConservaAtributos comprueba que el reparto de atributos sigue funcionando
// después de darle lectura propia al tipo, que es cuando la etiqueta `,any,attr` deja de
// aplicarse.
func TestProperties_ConservaAtributos(t *testing.T) {
	entrada := `<Properties atributo="valor" otro="2"><PageColor type="enumeration">X</PageColor></Properties>`

	var p Properties
	if err := xml.Unmarshal([]byte(entrada), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(p.OtherAttrs) != 2 {
		t.Fatalf("se esperaban 2 atributos en el comodín, hay %d: %v", len(p.OtherAttrs), p.OtherAttrs)
	}

	salida, err := xml.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, esperado := range []string{`atributo="valor"`, `otro="2"`} {
		if !bytes.Contains(salida, []byte(esperado)) {
			t.Errorf("falta %s en la salida: %s", esperado, salida)
		}
	}
}

// TestProperties_IdempotenciaAlParsearDosVeces comprueba que deserializar dos veces sobre el
// mismo valor no acumula hijos, atributos ni orden.
func TestProperties_IdempotenciaAlParsearDosVeces(t *testing.T) {
	entrada := []byte(`<Properties a="1"><PageColor type="enumeration">X</PageColor><Label><KeyValuePair Key="k" Value="v"/></Label></Properties>`)

	var p Properties
	if err := xml.Unmarshal(entrada, &p); err != nil {
		t.Fatalf("primera: %v", err)
	}
	tags1, attrs1 := len(p.ChildTags()), len(p.OtherAttrs)

	if err := xml.Unmarshal(entrada, &p); err != nil {
		t.Fatalf("segunda: %v", err)
	}
	if got := len(p.ChildTags()); got != tags1 {
		t.Errorf("el orden registrado pasó de %d a %d entradas: se está acumulando", tags1, got)
	}
	if got := len(p.OtherAttrs); got != attrs1 {
		t.Errorf("los atributos pasaron de %d a %d: se están acumulando", attrs1, got)
	}
	if got := len(p.OtherElements); got != 1 {
		t.Errorf("se esperaba 1 hijo en el comodín, hay %d", got)
	}
}

// TestProperties_ConservaElContenidoDeLosHijosNoModelados comprueba que el comodín no pierde
// el contenido interno, que es lo que hace que estos elementos se puedan reemitir literales.
func TestProperties_ConservaElContenidoDeLosHijosNoModelados(t *testing.T) {
	entrada := `<Properties><Descriptor type="list"><ListItem type="string">uno</ListItem><ListItem type="string">dos</ListItem></Descriptor></Properties>`

	var p Properties
	if err := xml.Unmarshal([]byte(entrada), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	salida, err := xml.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, esperado := range []string{"<ListItem", "uno", "dos", `type="list"`} {
		if !bytes.Contains(salida, []byte(esperado)) {
			t.Errorf("falta %q en la salida: %s", esperado, salida)
		}
	}
}
