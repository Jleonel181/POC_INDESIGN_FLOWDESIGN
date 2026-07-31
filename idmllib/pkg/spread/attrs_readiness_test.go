package spread

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/beevik/etree"
	"github.com/dimelords/idmllib/v2/internal/xmlutil"
)

// Este archivo comprueba que las funciones genéricas de atributos de la Tarea 4
// funcionan contra los tipos reales del modelo, antes de que la Tarea 5 las conecte.
//
// Vive en pkg/spread y no en internal/xmlutil porque necesita los tipos de aquí, y
// pkg/spread ya importa internal/xmlutil, así que no hay ciclo. Es la comprobación
// que evita descubrir en la Tarea 5 que la función no entiende la forma real de un
// elemento de página.

// TestAttrsGenericas_VenLosCamposHeredados es la comprobación que importa: Rectangle
// declara sus atributos propios y hereda seis de PageItemBase por embebido. Si las
// funciones genéricas no recorrieran el embebido, esos seis se tratarían como
// desconocidos y acabarían emitidos dos veces.
func TestAttrsGenericas_VenLosCamposHeredados(t *testing.T) {
	// Sin OtherAttrs todavía: eso lo agrega la Tarea 5. Lo que se comprueba aquí es
	// el reparto sobre los campos declarados, incluidos los heredados.
	var rect Rectangle

	heredados := []xml.Attr{
		{Name: xml.Name{Local: "Self"}, Value: "u2c4"},
		{Name: xml.Name{Local: "Name"}, Value: "$ID/"},
		{Name: xml.Name{Local: "ItemLayer"}, Value: "uba"},
		{Name: xml.Name{Local: "Visible"}, Value: "true"},
		{Name: xml.Name{Local: "GeometricBounds"}, Value: "0 0 10 10"},
		{Name: xml.Name{Local: "ItemTransform"}, Value: "1 0 0 1 0 0"},
	}
	propios := []xml.Attr{
		{Name: xml.Name{Local: "ContentType"}, Value: "GraphicType"},
		{Name: xml.Name{Local: "AppliedObjectStyle"}, Value: "ObjectStyle/$ID/[Normal Graphics Frame]"},
	}

	if err := xmlutil.UnmarshalAttrs(append(heredados, propios...), &rect); err != nil {
		t.Fatalf("UnmarshalAttrs sobre Rectangle falló: %v", err)
	}

	if rect.Self != "u2c4" {
		t.Errorf("Self heredado de PageItemBase = %q, esperado u2c4", rect.Self)
	}
	if rect.ItemLayer != "uba" || rect.GeometricBounds != "0 0 10 10" || rect.ItemTransform != "1 0 0 1 0 0" {
		t.Errorf("atributos heredados mal asignados: %+v", rect.PageItemBase)
	}
	if rect.ContentType != "GraphicType" {
		t.Errorf("ContentType propio = %q, esperado GraphicType", rect.ContentType)
	}

	// Y al emitir, cada nombre una sola vez.
	attrs, err := xmlutil.MarshalAttrs(rect)
	if err != nil {
		t.Fatalf("MarshalAttrs sobre Rectangle falló: %v", err)
	}
	vistos := map[string]int{}
	for _, a := range attrs {
		vistos[a.Name.Local]++
	}
	for name, n := range vistos {
		if n > 1 {
			t.Errorf("el atributo %q se emitió %d veces", name, n)
		}
	}
	for _, a := range append(heredados, propios...) {
		if vistos[a.Name.Local] == 0 {
			t.Errorf("el atributo %q no se emitió", a.Name.Local)
		}
	}
}

// TestAttrsGenericas_RectangleNoPierdeNingunAtributo es la prueba de que el comodín
// hace su trabajo sobre datos reales: el primer Rectangle del Documento_Referencia trae
// 33 atributos y los 33 sobreviven el ciclo.
//
// Antes de instalar el comodín este test comprobaba lo contrario, que se perdían 11, y
// estaba escrito para fallar cuando esa pérdida desapareciera. Fue lo que ocurrió.
func TestAttrsGenericas_RectangleNoPierdeNingunAtributo(t *testing.T) {
	const spreadPath = "../../testdata/documento_referencia/Spreads/Spread_uce7.xml"

	data, err := os.ReadFile(spreadPath)
	if err != nil {
		t.Skipf("Documento_Referencia ausente: %v (ruta esperada: %s)", err, spreadPath)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		t.Fatalf("no se pudo parsear el spread: %v", err)
	}
	inner := doc.Root().FindElement("Spread")
	if inner == nil {
		t.Fatal("el archivo no contiene el elemento Spread")
	}

	var primero *etree.Element
	for _, child := range inner.ChildElements() {
		if child.Tag == "Rectangle" {
			primero = child
			break
		}
	}
	if primero == nil {
		t.Skip("el spread no contiene ningún Rectangle")
	}

	attrs := make([]xml.Attr, 0, len(primero.Attr))
	for _, a := range primero.Attr {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: a.Space, Local: a.Key}, Value: a.Value})
	}

	var rect Rectangle
	if err := xmlutil.UnmarshalAttrs(attrs, &rect); err != nil {
		t.Fatalf("UnmarshalAttrs sobre el Rectangle del corpus falló: %v", err)
	}

	emitidos, err := xmlutil.MarshalAttrs(rect)
	if err != nil {
		t.Fatalf("MarshalAttrs falló: %v", err)
	}

	t.Logf("Rectangle del corpus: %d atributos en el archivo, %d emitidos", len(attrs), len(emitidos))

	if len(emitidos) != len(attrs) {
		t.Errorf("se emitieron %d atributos de los %d del archivo: el comodín debería conservarlos todos", len(emitidos), len(attrs))
	}

	// No basta con que cuadre el número: cada nombre y cada valor tiene que coincidir.
	// Un atributo perdido compensado por uno duplicado daría el mismo total.
	entrada := make(map[string]string, len(attrs))
	for _, a := range attrs {
		entrada[a.Name.Local] = a.Value
	}
	salida := make(map[string]string, len(emitidos))
	for _, a := range emitidos {
		if _, repetido := salida[a.Name.Local]; repetido {
			t.Errorf("el atributo %q se emitió más de una vez", a.Name.Local)
		}
		salida[a.Name.Local] = a.Value
	}
	for name, want := range entrada {
		got, ok := salida[name]
		if !ok {
			t.Errorf("el atributo %q no se emitió", name)
			continue
		}
		if got != want {
			t.Errorf("el atributo %q se emitió como %q, esperado %q", name, got, want)
		}
	}
	for name := range salida {
		if _, ok := entrada[name]; !ok {
			t.Errorf("se emitió el atributo %q, que no estaba en el archivo", name)
		}
	}
}
