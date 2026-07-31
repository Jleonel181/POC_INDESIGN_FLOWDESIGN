package spread

import (
	"bytes"
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

// Secuencias medidas en el Documento_Referencia. Son la verdad contra la que se compara:
// el orden en que InDesign escribió los elementos de página hijos directos.
var (
	secuenciaSpreadUce7 = []string{
		TagRectangle, TagRectangle, TagGroup,
		TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame,
		TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame,
		TagRectangle, TagRectangle, TagRectangle,
		TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame,
		TagRectangle, TagRectangle, TagRectangle,
		TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame,
		TagRectangle,
		TagTextFrame,
		TagGraphicLine, TagGroup,
	}

	secuenciaMasterU2dd = []string{
		TagGraphicLine, TagGraphicLine, TagGraphicLine,
		TagTextFrame, TagTextFrame, TagTextFrame, TagTextFrame,
		TagGraphicLine, TagGraphicLine,
		TagRectangle, TagRectangle,
		TagTextFrame,
		TagRectangle, TagRectangle,
		TagGraphicLine, TagGraphicLine, TagGraphicLine,
	}
)

// parseElementoInterno extrae del archivo el primer elemento con el nombre local dado y
// lo deserializa en un SpreadElement.
//
// Hace falta porque un Spread viene envuelto en <idPkg:Spread> y un MasterSpread en
// <idPkg:MasterSpread>, y aquí interesa el elemento de dentro.
func parseElementoInterno(t *testing.T, ruta, local string) *SpreadElement {
	t.Helper()
	data, err := os.ReadFile(ruta)
	if err != nil {
		t.Skipf("no encuentro %s: %v", ruta, err)
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := d.Token()
		if err != nil {
			t.Fatalf("no encontré el elemento <%s> en %s: %v", local, ruta, err)
		}
		start, ok := tok.(xml.StartElement)
		// El nombre local del envoltorio <idPkg:Spread> es también "Spread", así que hay
		// que exigir que no tenga namespace: el elemento de dentro es el que interesa.
		if !ok || start.Name.Local != local || start.Name.Space != "" {
			continue
		}
		var se SpreadElement
		if err := d.DecodeElement(&se, &start); err != nil {
			t.Fatalf("deserializando <%s> de %s: %v", local, ruta, err)
		}
		return &se
	}
}

// cuenta agrupa una secuencia de nombres de elemento por nombre.
func cuenta(seq []string) map[string]int {
	out := make(map[string]int, len(seq))
	for _, s := range seq {
		out[s]++
	}
	return out
}

// agrupadaPorTipo devuelve la permutación de una secuencia que resulta de emitir los
// elementos agrupados por tipo, en el orden en que SpreadElement declara sus campos. Es
// exactamente el defecto que esta tarea arregla, y sirve de control negativo.
func agrupadaPorTipo(seq []string) []string {
	n := cuenta(seq)
	out := make([]string, 0, len(seq))
	for _, kind := range childFieldOrder {
		for i := 0; i < n[kind]; i++ {
			out = append(out, kind)
		}
	}
	return out
}

// difierenEn cuenta en cuántas posiciones difieren dos secuencias de la misma longitud.
func difierenEn(a, b []string) int {
	d := 0
	for i := range a {
		if a[i] != b[i] {
			d++
		}
	}
	return d
}

// TestItems_OrdenDocumentalAlParsear comprueba el criterio 4 de la Tarea 7: la secuencia
// completa de elementos de página se conserva al parsear, con sus conteos por tipo.
func TestItems_OrdenDocumentalAlParsear(t *testing.T) {
	casos := []struct {
		nombre     string
		ruta       string
		local      string
		esperada   []string
		porTipo    map[string]int
		posiciones int
	}{
		{
			nombre:   "Spread_uce7",
			ruta:     "../../testdata/documento_referencia/Spreads/Spread_uce7.xml",
			local:    "Spread",
			esperada: secuenciaSpreadUce7,
			porTipo:  map[string]int{TagTextFrame: 20, TagRectangle: 9, TagGroup: 2, TagGraphicLine: 1},
			// La permutación agrupada por tipo difiere en 13 posiciones de las 32.
			posiciones: 13,
		},
		{
			nombre:   "MasterSpread_u2dd",
			ruta:     "../../testdata/documento_referencia/MasterSpreads/MasterSpread_u2dd.xml",
			local:    "MasterSpread",
			esperada: secuenciaMasterU2dd,
			porTipo:  map[string]int{TagGraphicLine: 8, TagTextFrame: 5, TagRectangle: 4},
			// Difiere en 12 posiciones de las 17.
			posiciones: 12,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			se := parseElementoInterno(t, c.ruta, c.local)
			obtenida := se.ItemTags()

			if len(obtenida) != len(c.esperada) {
				t.Fatalf("número de elementos de página: esperado %d, obtenido %d\n  esperado: %v\n  obtenido: %v",
					len(c.esperada), len(obtenida), c.esperada, obtenida)
			}
			for i := range c.esperada {
				if obtenida[i] != c.esperada[i] {
					t.Errorf("posición %d: esperado %q, obtenido %q\n  esperado: %v\n  obtenido: %v",
						i, c.esperada[i], obtenida[i], c.esperada, obtenida)
					break
				}
			}

			for kind, n := range c.porTipo {
				if got := cuenta(obtenida)[kind]; got != n {
					t.Errorf("%s: esperados %d, obtenidos %d", kind, n, got)
				}
			}

			// Control negativo: la secuencia leída NO es la agrupada por tipo. Sin esto,
			// un modelo que reagrupara pasaría los conteos y fallaría solo el orden, y
			// el test no diría que el defecto es la reagrupación.
			agrupada := agrupadaPorTipo(c.esperada)
			if d := difierenEn(c.esperada, agrupada); d != c.posiciones {
				t.Errorf("la permutación agrupada por tipo debería diferir en %d posiciones, difiere en %d",
					c.posiciones, d)
			}
			if d := difierenEn(obtenida, agrupada); d != c.posiciones {
				t.Errorf("lo parseado difiere de la agrupada por tipo en %d posiciones, se esperaban %d: se está reagrupando",
					d, c.posiciones)
			}
		})
	}
}

// TestItems_OrdenDocumentalAlSerializar comprueba el criterio 5: los elementos de página
// se emiten en Orden_Documental, sin reagrupar por tipo.
func TestItems_OrdenDocumentalAlSerializar(t *testing.T) {
	se := parseElementoInterno(t, "../../testdata/documento_referencia/Spreads/Spread_uce7.xml", "Spread")

	data, err := xml.Marshal(se)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	emitida := hijosDirectos(t, data, "Spread")
	// Al serializar salen también FlattenerPreference y las Page, que no son elementos
	// de página; se filtran para comparar contra la secuencia de Items.
	var soloItems []string
	for _, h := range emitida {
		if h != tagFlattenerPreference && h != tagPage {
			soloItems = append(soloItems, h)
		}
	}

	if len(soloItems) != len(secuenciaSpreadUce7) {
		t.Fatalf("elementos de página emitidos: esperados %d, obtenidos %d",
			len(secuenciaSpreadUce7), len(soloItems))
	}
	for i := range secuenciaSpreadUce7 {
		if soloItems[i] != secuenciaSpreadUce7[i] {
			t.Fatalf("posición %d al emitir: esperado %q, obtenido %q",
				i, secuenciaSpreadUce7[i], soloItems[i])
		}
	}

	// El prefijo tiene que seguir siendo FlattenerPreference y luego la Page.
	if len(emitida) < 2 || emitida[0] != tagFlattenerPreference || emitida[1] != tagPage {
		t.Errorf("el prefijo emitido debería ser [FlattenerPreference Page], es %v", emitida[:min(2, len(emitida))])
	}
}

// TestItems_ModeloDesdeCeroEmiteEnOrdenDeCampos comprueba que un SpreadElement que nunca
// se parseó sigue emitiendo en el orden de los campos, que es el comportamiento anterior
// a esta tarea. Es lo que permite que las 18 escrituras existentes no cambien.
func TestItems_ModeloDesdeCeroEmiteEnOrdenDeCampos(t *testing.T) {
	var se SpreadElement
	se.Self = "ue6"
	se.Rectangles = []Rectangle{{PageItemBase: PageItemBase{Self: "r1"}}}
	se.TextFrames = []SpreadTextFrame{{PageItemBase: PageItemBase{Self: "t1"}}}

	data, err := xml.Marshal(&se)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	hijos := hijosDirectos(t, data, "Spread")

	// Orden de los campos: TextFrames antes que Rectangles.
	esperado := []string{TagTextFrame, TagRectangle}
	if len(hijos) != len(esperado) {
		t.Fatalf("hijos emitidos: esperados %v, obtenidos %v", esperado, hijos)
	}
	for i := range esperado {
		if hijos[i] != esperado[i] {
			t.Fatalf("posición %d: esperado %q, obtenido %q (%v)", i, esperado[i], hijos[i], hijos)
		}
	}
	if len(se.Items) != 0 {
		t.Errorf("un modelo sin parsear no debería tener Items, tiene %d", len(se.Items))
	}
}

// TestItems_EscrituraEnCampoPorTipoSigueSurtiendoEfecto es la guarda del criterio 3.
//
// Reproduce lo que hace removeItemFromSpread de pkg/idml: parsear un spread y borrar un
// elemento del slice por tipo. Si el contenido se leyera de Items en lugar del campo, el
// borrado no llegaría al XML emitido y la pérdida sería silenciosa, porque
// TestRemoveTextFrame_Basic solo comprueba el slice en memoria.
func TestItems_EscrituraEnCampoPorTipoSigueSurtiendoEfecto(t *testing.T) {
	se := parseElementoInterno(t, "../../testdata/documento_referencia/Spreads/Spread_uce7.xml", "Spread")

	antes := cuenta(hijosDirectos(t, marshalOFallar(t, se), "Spread"))[TagTextFrame]
	if antes != 20 {
		t.Fatalf("preparación: se esperaban 20 TextFrame emitidos, hay %d", antes)
	}

	// Borrado por slice, el modismo exacto de removeItemFromSpread.
	se.TextFrames = append(se.TextFrames[:0], se.TextFrames[1:]...)

	despues := cuenta(hijosDirectos(t, marshalOFallar(t, se), "Spread"))[TagTextFrame]
	if despues != antes-1 {
		t.Errorf("tras borrar del slice se esperaban %d TextFrame emitidos, hay %d: la escritura al campo por tipo no llegó a la salida",
			antes-1, despues)
	}
}

// TestItems_PunterosNoCopias comprueba el criterio 1 en lo que se puede comprobar en esta
// fase: Items guarda punteros a los elementos de los campos por tipo, no copias, así que
// mutar a través de Items se ve en el campo y en la salida.
func TestItems_PunterosNoCopias(t *testing.T) {
	se := parseElementoInterno(t, "../../testdata/documento_referencia/Spreads/Spread_uce7.xml", "Spread")
	if len(se.Items) == 0 {
		t.Fatal("preparación: Items está vacío")
	}

	// El primer elemento del spread es un Rectangle; se le cambia el nombre a través de
	// Items y tiene que verse en el campo Rectangles y en el XML.
	rect, ok := se.Items[0].(*Rectangle)
	if !ok {
		t.Fatalf("se esperaba que Items[0] fuera *Rectangle, es %T", se.Items[0])
	}
	rect.Name = "marcado-por-la-prueba"

	if se.Rectangles[0].Name != "marcado-por-la-prueba" {
		t.Error("mutar a través de Items no se refleja en el campo Rectangles: Items guarda copias")
	}
	if !bytes.Contains(marshalOFallar(t, se), []byte("marcado-por-la-prueba")) {
		t.Error("la mutación hecha a través de Items no aparece en el XML emitido")
	}
}

// TestItems_IdempotenciaAlParsearDosVeces comprueba que deserializar dos veces sobre el
// mismo valor no acumula hijos ni duplica el orden registrado.
func TestItems_IdempotenciaAlParsearDosVeces(t *testing.T) {
	ruta := "../../testdata/documento_referencia/Spreads/Spread_uce7.xml"
	data, err := os.ReadFile(ruta)
	if err != nil {
		t.Skipf("no encuentro %s: %v", ruta, err)
	}

	decodificar := func(se *SpreadElement) {
		d := xml.NewDecoder(bytes.NewReader(data))
		for {
			tok, err := d.Token()
			if err != nil {
				t.Fatalf("no encontré <Spread>: %v", err)
			}
			// Mismo cuidado que en parseElementoInterno: el envoltorio idPkg:Spread
			// también tiene "Spread" como nombre local.
			if start, ok := tok.(xml.StartElement); ok && start.Name.Local == "Spread" && start.Name.Space == "" {
				if err := d.DecodeElement(se, &start); err != nil {
					t.Fatalf("decode: %v", err)
				}
				return
			}
		}
	}

	var se SpreadElement
	decodificar(&se)
	primera := len(se.Items)
	decodificar(&se)

	if len(se.Items) != primera {
		t.Errorf("tras deserializar dos veces Items pasó de %d a %d: se está acumulando", primera, len(se.Items))
	}
	if n := len(se.TextFrames); n != 20 {
		t.Errorf("tras deserializar dos veces se esperaban 20 TextFrame, hay %d", n)
	}
}

// TestItems_HijosNoModeladosConservanPosicion comprueba el criterio 6: un hijo que el
// modelo no declara se guarda como XML crudo y se re-emite en su posición.
func TestItems_HijosNoModeladosConservanPosicion(t *testing.T) {
	entrada := `<Spread Self="u1">` +
		`<Rectangle Self="r1"/>` +
		`<ElementoInventado Attr="v"><Dentro/></ElementoInventado>` +
		`<TextFrame Self="t1"/>` +
		`</Spread>`

	var se SpreadElement
	if err := xml.Unmarshal([]byte(entrada), &se); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(se.OtherElements) != 1 {
		t.Fatalf("se esperaba 1 hijo no modelado, hay %d", len(se.OtherElements))
	}
	// No es un elemento de página, así que no entra en Items.
	if got := se.ItemTags(); len(got) != 2 {
		t.Errorf("Items debería tener solo los 2 elementos de página, tiene %v", got)
	}

	salida := marshalOFallar(t, &se)
	hijos := hijosDirectos(t, salida, "Spread")
	esperado := []string{TagRectangle, "ElementoInventado", TagTextFrame}
	for i := range esperado {
		if i >= len(hijos) || hijos[i] != esperado[i] {
			t.Fatalf("orden emitido: esperado %v, obtenido %v", esperado, hijos)
		}
	}
	if !bytes.Contains(salida, []byte("<Dentro>")) && !bytes.Contains(salida, []byte("<Dentro/>")) {
		t.Errorf("el contenido interno del hijo no modelado se perdió: %s", salida)
	}
}

// marshalOFallar serializa y falla el test si hay error.
func marshalOFallar(t *testing.T, se *SpreadElement) []byte {
	t.Helper()
	data, err := xml.Marshal(se)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}

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

// min evita depender de la versión del lenguaje: go.mod declara go 1.21.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Comprobación de que la lista de clases de childFieldOrder cubre las siete que se
// pueblan por tipo, más FlattenerPreference y Page. Una clase que falte no se emitiría
// cuando el registro de orden no la menciona.
func TestChildFieldOrder_CubreTodasLasClases(t *testing.T) {
	necesarias := []string{
		tagFlattenerPreference, tagPage,
		TagTextFrame, TagRectangle, TagImage, TagOval, TagPolygon, TagGraphicLine, TagGroup,
	}
	for _, n := range necesarias {
		encontrada := false
		for _, k := range childFieldOrder {
			if k == n {
				encontrada = true
				break
			}
		}
		if !encontrada {
			t.Errorf("childFieldOrder no menciona %q, sus elementos se perderían al emitir un modelo sin orden registrado", n)
		}
	}
	if len(childFieldOrder) != len(necesarias) {
		t.Errorf("childFieldOrder tiene %d clases y se esperaban %d: %v",
			len(childFieldOrder), len(necesarias), strings.Join(childFieldOrder, ", "))
	}
}
