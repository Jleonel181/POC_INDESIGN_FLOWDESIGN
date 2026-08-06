package spread

import (
	"encoding/xml"
	"testing"
)

// TestAppend_EmiteEnOrdenDeAdicion es la razón de ser de Append, y el criterio 4 de la
// Tarea 19: los elementos salen en el orden en que se agregaron, no agrupados por tipo.
func TestAppend_EmiteEnOrdenDeAdicion(t *testing.T) {
	var se SpreadElement
	se.Self = "ue6"

	// Orden de adición deliberadamente intercalado, distinto del orden de los campos.
	agregados := []PageItem{
		&Rectangle{PageItemBase: PageItemBase{Self: "r1"}},
		&SpreadTextFrame{PageItemBase: PageItemBase{Self: "t1"}},
		&Rectangle{PageItemBase: PageItemBase{Self: "r2"}},
		&GraphicLine{PageItemBase: PageItemBase{Self: "l1"}},
		&SpreadTextFrame{PageItemBase: PageItemBase{Self: "t2"}},
	}
	for _, it := range agregados {
		if _, err := se.Append(it); err != nil {
			t.Fatalf("Append(%s): %v", it.GetSelf(), err)
		}
	}

	esperado := []string{TagRectangle, TagTextFrame, TagRectangle, TagGraphicLine, TagTextFrame}

	if got := se.ItemTags(); !mismaSecuencia(got, esperado) {
		t.Errorf("Items: esperado %v, obtenido %v", esperado, got)
	}

	data, err := xml.Marshal(&se)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := hijosDirectos(t, data, "Spread"); !mismaSecuencia(got, esperado) {
		t.Errorf("emitido: esperado %v, obtenido %v\n%s", esperado, got, data)
	}

	// Control negativo: el orden de adición NO es el orden de los campos, así que este
	// test detecta de verdad la reagrupación.
	porCampos := agrupadaPorTipo(esperado)
	if mismaSecuencia(esperado, porCampos) {
		t.Fatal("el caso de prueba es inútil: el orden de adición coincide con el de los campos")
	}
}

// TestAppend_DevuelveElPunteroCanonico comprueba que el puntero devuelto es el que se usa
// al emitir, así que modificar a través de él se ve en la salida.
func TestAppend_DevuelveElPunteroCanonico(t *testing.T) {
	var se SpreadElement
	guardado, err := se.Append(&Rectangle{PageItemBase: PageItemBase{Self: "r1"}})
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	rect, ok := guardado.(*Rectangle)
	if !ok {
		t.Fatalf("se esperaba *Rectangle, se obtuvo %T", guardado)
	}
	rect.Name = "modificado-despues-de-agregar"

	if se.rectangles[0].Name != "modificado-despues-de-agregar" {
		t.Error("el puntero devuelto no apunta al elemento guardado en el campo por tipo")
	}
	data, err := xml.Marshal(&se)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !contiene(data, "modificado-despues-de-agregar") {
		t.Errorf("la modificación hecha por el puntero devuelto no llegó al XML: %s", data)
	}
}

// TestAppend_GuardaCopiaNoElPunteroDelLlamador documenta la consecuencia del diseño: el
// contenido vive en el campo por tipo, así que el puntero que pasó el llamador queda
// desconectado. Está en un test para que el día que cambie, se vea.
func TestAppend_GuardaCopiaNoElPunteroDelLlamador(t *testing.T) {
	var se SpreadElement
	original := &Rectangle{PageItemBase: PageItemBase{Self: "r1"}}
	if _, err := se.Append(original); err != nil {
		t.Fatalf("Append: %v", err)
	}

	original.Name = "cambiado-en-el-original"

	if se.rectangles[0].Name == "cambiado-en-el-original" {
		t.Error("Append guardó el puntero del llamador; el contrato documentado dice que guarda una copia")
	}
}

// TestAppend_MantieneValidosLosPunterosDeItems es la guarda del riesgo real del diseño:
// hacer append a un campo por tipo puede reubicar el array de ese slice, y los punteros
// que Items ya tenía apuntarían al array viejo.
func TestAppend_MantieneValidosLosPunterosDeItems(t *testing.T) {
	var se SpreadElement

	// Suficientes elementos del mismo tipo para forzar al menos una reubicación del slice.
	const n = 64
	for i := 0; i < n; i++ {
		if _, err := se.Append(&Rectangle{PageItemBase: PageItemBase{Self: idDePrueba(i)}}); err != nil {
			t.Fatalf("Append %d: %v", i, err)
		}
	}
	if len(se.Items) != n {
		t.Fatalf("Items tiene %d elementos, se esperaban %d", len(se.Items), n)
	}

	// Cada puntero de Items tiene que apuntar al elemento correspondiente del campo, no a
	// una copia de un array viejo. Se comprueba mutando por Items y leyendo del campo.
	for i := range se.Items {
		rect, ok := se.Items[i].(*Rectangle)
		if !ok {
			t.Fatalf("Items[%d] es %T", i, se.Items[i])
		}
		rect.Name = "marca"
		if se.rectangles[i].Name != "marca" {
			t.Fatalf("Items[%d] no apunta a Rectangles[%d]: el append reubicó el slice y el puntero quedó viejo", i, i)
		}
	}
}

// TestAppend_TrasParsearAgregaAlFinal comprueba el criterio 4 de la Tarea 8 sobre un
// spread ya parseado: el elemento nuevo se emite en la última posición, que para InDesign
// es el elemento más al frente.
func TestAppend_TrasParsearAgregaAlFinal(t *testing.T) {
	se := parseElementoInterno(t, "../../testdata/documento_referencia/Spreads/Spread_uce7.xml", "Spread")
	antes := len(se.Items)

	if _, err := se.Append(&Oval{PageItemBase: PageItemBase{Self: "nuevo-oval"}}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	if len(se.Items) != antes+1 {
		t.Fatalf("Items pasó de %d a %d, se esperaba %d", antes, len(se.Items), antes+1)
	}
	if tags := se.ItemTags(); tags[len(tags)-1] != TagOval {
		t.Errorf("el elemento agregado debería ser el último, la secuencia acaba en %q", tags[len(tags)-1])
	}

	data, err := xml.Marshal(se)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	hijos := hijosDirectos(t, data, "Spread")
	if hijos[len(hijos)-1] != TagOval {
		t.Errorf("el XML debería acabar en <Oval>, acaba en <%s>", hijos[len(hijos)-1])
	}
	// Y el orden de los 32 que ya estaban no cambia.
	var soloItems []string
	for _, h := range hijos {
		if h != tagFlattenerPreference && h != tagPage {
			soloItems = append(soloItems, h)
		}
	}
	if !mismaSecuencia(soloItems[:len(secuenciaSpreadUce7)], secuenciaSpreadUce7) {
		t.Error("agregar un elemento cambió el orden de los que ya estaban")
	}
}

// TestAppend_RechazaLoQueNoPuedeGuardar comprueba la validación: un tipo sin campo donde
// guardarse da error en lugar de desaparecer en silencio.
func TestAppend_RechazaLoQueNoPuedeGuardar(t *testing.T) {
	var se SpreadElement

	if _, err := se.Append(nil); err == nil {
		t.Error("Append(nil) debería dar error")
	}

	// *PDF satisface PageItem pero SpreadElement no tiene campo para un PDF suelto.
	got, err := se.Append(&PDF{FrameContentBase: FrameContentBase{Self: "pdf1"}})
	if err == nil {
		t.Error("Append(*PDF) debería dar error: no hay campo donde guardarlo")
	}
	if got != nil {
		t.Errorf("con error, el elemento devuelto debería ser nil, es %T", got)
	}
	if len(se.Items) != 0 || len(se.OtherElements) != 0 {
		t.Error("un Append con error no debería dejar rastro en el modelo")
	}
}

// mismaSecuencia compara dos secuencias de nombres.
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

// contiene indica si data contiene la subcadena s.
func contiene(data []byte, s string) bool {
	return len(data) >= len(s) && indexOf(string(data), s) >= 0
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}

// idDePrueba genera un Self distinto por índice.
func idDePrueba(i int) string {
	const hex = "0123456789abcdef"
	return "u" + string([]byte{hex[(i>>4)&0xf], hex[i&0xf]})
}
