package xmlutil

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// baseAttrs imita a PageItemBase de pkg/spread: un struct embebido que aporta
// atributos. Es la forma que tienen los tipos reales, y la que rompería si las
// funciones no recorrieran los embebidos.
type baseAttrs struct {
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr,omitempty"`
}

// itemAttrs imita a Rectangle: embebe la base, declara algunos atributos propios y
// tiene el comodín.
type itemAttrs struct {
	baseAttrs

	ContentType string `xml:"ContentType,attr,omitempty"`
	Visible     string `xml:"Visible,attr,omitempty"`

	Ignored    string `xml:"-"`
	NotAnAttr  string `xml:"NotAnAttr"`
	OtherAttrs []xml.Attr
}

// typedAttrs sirve para las conversiones y sus errores. En el modelo real los 256
// campos de atributo son de tipo string, así que este struct es el único sitio donde
// las conversiones no triviales se ejercitan.
type typedAttrs struct {
	Count   int     `xml:"Count,attr"`
	Ratio   float64 `xml:"Ratio,attr"`
	Enabled bool    `xml:"Enabled,attr"`

	OtherAttrs []xml.Attr
}

// noWildcardAttrs no declara comodín: sus atributos no declarados se descartan sin
// que las funciones fallen.
type noWildcardAttrs struct {
	Self string `xml:"Self,attr"`
}

func attr(name, value string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: name}, Value: value}
}

func attrNames(attrs []xml.Attr) []string {
	names := make([]string, len(attrs))
	for i, a := range attrs {
		names[i] = attrKey(a.Name)
	}
	return names
}

// TestUnmarshalAttrs_RepartoEntreCamposYComodin es el caso central: lo declarado va a
// su campo, lo demás al comodín, y los atributos del struct embebido cuentan como
// declarados.
func TestUnmarshalAttrs_RepartoEntreCamposYComodin(t *testing.T) {
	var item itemAttrs
	err := UnmarshalAttrs([]xml.Attr{
		attr("Self", "u123"),
		attr("FillColor", "Color/Black"),
		attr("ContentType", "GraphicType"),
		attr("CornerRadius", "12"),
		attr("Name", "$ID/"),
	}, &item)
	if err != nil {
		t.Fatalf("UnmarshalAttrs falló: %v", err)
	}

	if item.Self != "u123" {
		t.Errorf("Self del struct embebido = %q, esperado u123: los campos del embebido deben contar como declarados", item.Self)
	}
	if item.Name != "$ID/" {
		t.Errorf("Name = %q, esperado $ID/", item.Name)
	}
	if item.ContentType != "GraphicType" {
		t.Errorf("ContentType = %q, esperado GraphicType", item.ContentType)
	}

	// Solo los dos no declarados, y en su orden de aparición.
	want := []string{"FillColor", "CornerRadius"}
	if got := attrNames(item.OtherAttrs); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("OtherAttrs = %v, esperado %v", got, want)
	}
	if item.OtherAttrs[0].Value != "Color/Black" {
		t.Errorf("el comodín debe guardar el valor literal, guardó %q", item.OtherAttrs[0].Value)
	}
}

// TestUnmarshalAttrs_CampoNoAtributoNoCapturaNada comprueba que un campo sin `,attr` y
// uno con `xml:"-"` no se confunden con atributos: sus nombres deben acabar en el
// comodín si aparecen como atributos en la entrada.
func TestUnmarshalAttrs_CampoNoAtributoNoCapturaNada(t *testing.T) {
	var item itemAttrs
	if err := UnmarshalAttrs([]xml.Attr{
		attr("Ignored", "x"),
		attr("NotAnAttr", "y"),
	}, &item); err != nil {
		t.Fatalf("UnmarshalAttrs falló: %v", err)
	}

	if item.Ignored != "" || item.NotAnAttr != "" {
		t.Errorf("un campo con xml:\"-\" o sin ,attr no debe recibir atributos: Ignored=%q NotAnAttr=%q", item.Ignored, item.NotAnAttr)
	}
	if got := attrNames(item.OtherAttrs); len(got) != 2 {
		t.Errorf("los dos atributos deben ir al comodín, fueron %v", got)
	}
}

// TestMarshalAttrs_DeclaradosPrimeroYSinDuplicar cubre el Req 2 criterio 3.
func TestMarshalAttrs_DeclaradosPrimeroYSinDuplicar(t *testing.T) {
	item := itemAttrs{
		baseAttrs:   baseAttrs{Self: "u1", Name: "nombre"},
		ContentType: "TextType",
		OtherAttrs: []xml.Attr{
			attr("FillColor", "Color/Black"),
			// Coincide con un campo declarado: no debe emitirse dos veces, y debe
			// ganar el valor del campo.
			attr("ContentType", "ValorViejo"),
		},
	}

	attrs, err := MarshalAttrs(item)
	if err != nil {
		t.Fatalf("MarshalAttrs falló: %v", err)
	}

	want := []string{"Self", "Name", "ContentType", "FillColor"}
	if got := attrNames(attrs); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("orden de atributos = %v, esperado %v", got, want)
	}

	for _, a := range attrs {
		if a.Name.Local == "ContentType" && a.Value != "TextType" {
			t.Errorf("ante un nombre duplicado debe ganar el campo declarado, se emitió %q", a.Value)
		}
	}
}

// TestMarshalAttrs_OmiteVaciosSegunElTag comprueba la regla de omitempty, para no
// cambiar la forma de los documentos ya emitidos.
func TestMarshalAttrs_OmiteVaciosSegunElTag(t *testing.T) {
	attrs, err := MarshalAttrs(itemAttrs{})
	if err != nil {
		t.Fatalf("MarshalAttrs falló: %v", err)
	}

	// Self no lleva omitempty, así que se emite vacío. El resto sí lo lleva.
	want := []string{"Self"}
	if got := attrNames(attrs); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("con el struct vacío se esperaban %v, se emitieron %v", want, got)
	}
}

// TestMarshalAttrs_ComodinVacioNoAgregaNada cubre la última parte del criterio 3.
func TestMarshalAttrs_ComodinVacioNoAgregaNada(t *testing.T) {
	attrs, err := MarshalAttrs(itemAttrs{baseAttrs: baseAttrs{Self: "u1"}})
	if err != nil {
		t.Fatalf("MarshalAttrs falló: %v", err)
	}
	if len(attrs) != 1 {
		t.Errorf("con el comodín vacío solo debe emitirse Self, se emitieron %v", attrNames(attrs))
	}
}

// TestAttrs_Idempotencia cubre el Req 2 criterio 10, y es el test que detecta el fallo
// tramposo: preservar todo en la primera vuelta y duplicar o perder en la segunda.
func TestAttrs_Idempotencia(t *testing.T) {
	entrada := []xml.Attr{
		attr("Self", "u123"),
		attr("FillColor", "Color/Black"),
		attr("ContentType", "GraphicType"),
		attr("CornerRadius", "12"),
		attr("StrokeWeight", "1"),
	}

	var primero itemAttrs
	if err := UnmarshalAttrs(entrada, &primero); err != nil {
		t.Fatalf("primer ciclo, UnmarshalAttrs: %v", err)
	}
	salida1, err := MarshalAttrs(primero)
	if err != nil {
		t.Fatalf("primer ciclo, MarshalAttrs: %v", err)
	}

	var segundo itemAttrs
	if err := UnmarshalAttrs(salida1, &segundo); err != nil {
		t.Fatalf("segundo ciclo, UnmarshalAttrs: %v", err)
	}
	salida2, err := MarshalAttrs(segundo)
	if err != nil {
		t.Fatalf("segundo ciclo, MarshalAttrs: %v", err)
	}

	if strings.Join(attrNames(salida1), ",") != strings.Join(attrNames(salida2), ",") {
		t.Errorf("el segundo ciclo dio otro resultado\nprimero: %v\nsegundo: %v", attrNames(salida1), attrNames(salida2))
	}
	if len(salida2) != len(entrada) {
		t.Errorf("el segundo ciclo emitió %d atributos y la entrada tenía %d", len(salida2), len(entrada))
	}
	if strings.Join(attrNames(segundo.OtherAttrs), ",") != strings.Join(attrNames(primero.OtherAttrs), ",") {
		t.Errorf("el comodín cambió entre ciclos: %v vs %v", attrNames(primero.OtherAttrs), attrNames(segundo.OtherAttrs))
	}
}

// TestUnmarshalAttrs_ReutilizarElDestinoNoAcumula es el motivo por el que
// UnmarshalAttrs vacía el comodín antes de rellenarlo. Sin eso, decodificar dos veces
// sobre el mismo struct duplicaría los atributos.
func TestUnmarshalAttrs_ReutilizarElDestinoNoAcumula(t *testing.T) {
	entrada := []xml.Attr{attr("Self", "u1"), attr("FillColor", "Color/Black")}

	var item itemAttrs
	for i := 0; i < 3; i++ {
		if err := UnmarshalAttrs(entrada, &item); err != nil {
			t.Fatalf("vuelta %d: %v", i, err)
		}
	}

	if len(item.OtherAttrs) != 1 {
		t.Errorf("tras tres decodificaciones el comodín tiene %d atributos, esperado 1: %v", len(item.OtherAttrs), attrNames(item.OtherAttrs))
	}
}

// TestAttrs_CientoVeintiochoAtributos cubre el Req 2 criterio 1: al menos 128
// atributos por elemento, con su orden intacto.
func TestAttrs_CientoVeintiochoAtributos(t *testing.T) {
	const n = 128

	entrada := make([]xml.Attr, 0, n)
	for i := 0; i < n; i++ {
		entrada = append(entrada, attr("Attr"+strconv.Itoa(i), "v"+strconv.Itoa(i)))
	}

	var item itemAttrs
	if err := UnmarshalAttrs(entrada, &item); err != nil {
		t.Fatalf("UnmarshalAttrs con %d atributos: %v", n, err)
	}
	if len(item.OtherAttrs) != n {
		t.Fatalf("el comodín guardó %d atributos de %d", len(item.OtherAttrs), n)
	}

	// El orden de aparición se conserva, no solo el conjunto.
	for i, a := range item.OtherAttrs {
		wantName := "Attr" + strconv.Itoa(i)
		if a.Name.Local != wantName || a.Value != "v"+strconv.Itoa(i) {
			t.Fatalf("posición %d: %s=%q, esperado %s", i, a.Name.Local, a.Value, wantName)
		}
	}

	salida, err := MarshalAttrs(item)
	if err != nil {
		t.Fatalf("MarshalAttrs: %v", err)
	}
	// Self se emite además, por no llevar omitempty.
	if len(salida) != n+1 {
		t.Errorf("se emitieron %d atributos, esperados %d", len(salida), n+1)
	}
}

// TestUnmarshalAttrs_Conversiones comprueba los tipos que no son string.
func TestUnmarshalAttrs_Conversiones(t *testing.T) {
	var typed typedAttrs
	err := UnmarshalAttrs([]xml.Attr{
		attr("Count", "42"),
		attr("Ratio", "1.5"),
		attr("Enabled", "true"),
	}, &typed)
	if err != nil {
		t.Fatalf("UnmarshalAttrs falló: %v", err)
	}

	if typed.Count != 42 || typed.Ratio != 1.5 || !typed.Enabled {
		t.Errorf("conversión incorrecta: %+v", typed)
	}

	salida, err := MarshalAttrs(typed)
	if err != nil {
		t.Fatalf("MarshalAttrs falló: %v", err)
	}
	want := map[string]string{"Count": "42", "Ratio": "1.5", "Enabled": "true"}
	for _, a := range salida {
		if want[a.Name.Local] != a.Value {
			t.Errorf("%s se emitió como %q, esperado %q", a.Name.Local, a.Value, want[a.Name.Local])
		}
	}
}

// TestUnmarshalAttrs_ErrorDeConversion cubre el Req 2 criterio 9: el error identifica
// el atributo, el valor y el tipo, y el campo se queda sin asignar.
func TestUnmarshalAttrs_ErrorDeConversion(t *testing.T) {
	tests := []struct {
		name  string
		attrs []xml.Attr
		field string
	}{
		{name: "entero", attrs: []xml.Attr{attr("Count", "muchos")}, field: "Count"},
		{name: "flotante", attrs: []xml.Attr{attr("Ratio", "1,5")}, field: "Ratio"},
		{name: "booleano", attrs: []xml.Attr{attr("Enabled", "quizá")}, field: "Enabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typed := typedAttrs{Count: 7, Ratio: 7, Enabled: true}
			err := UnmarshalAttrs(tt.attrs, &typed)
			if err == nil {
				t.Fatal("se esperaba un error de conversión")
			}

			msg := err.Error()
			for _, want := range []string{tt.field, tt.attrs[0].Value} {
				if !strings.Contains(msg, want) {
					t.Errorf("el error debe nombrar %q, dice: %s", want, msg)
				}
			}
			if !strings.Contains(msg, "int") && !strings.Contains(msg, "float") && !strings.Contains(msg, "bool") {
				t.Errorf("el error debe nombrar el tipo esperado, dice: %s", msg)
			}
		})
	}
}

// TestUnmarshalAttrs_SinComodinNoFalla comprueba que un struct que no declara
// OtherAttrs sigue funcionando: descarta lo no declarado, como antes.
func TestUnmarshalAttrs_SinComodinNoFalla(t *testing.T) {
	var nw noWildcardAttrs
	if err := UnmarshalAttrs([]xml.Attr{attr("Self", "u1"), attr("Desconocido", "x")}, &nw); err != nil {
		t.Fatalf("un struct sin comodín no debe hacer fallar UnmarshalAttrs: %v", err)
	}
	if nw.Self != "u1" {
		t.Errorf("Self = %q, esperado u1", nw.Self)
	}
}

// TestAttrs_NamespaceSePreserva comprueba que un atributo con prefijo, incluida una
// declaración de namespace, sobrevive el ciclo. Si se perdiera, el elemento se
// re-emitiría sin su namespace.
func TestAttrs_NamespaceSePreserva(t *testing.T) {
	entrada := []xml.Attr{
		{Name: xml.Name{Space: "xmlns", Local: "idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
		{Name: xml.Name{Space: "idPkg", Local: "Algo"}, Value: "valor"},
	}

	var item itemAttrs
	if err := UnmarshalAttrs(entrada, &item); err != nil {
		t.Fatalf("UnmarshalAttrs falló: %v", err)
	}
	if len(item.OtherAttrs) != 2 {
		t.Fatalf("el comodín debe guardar los dos atributos con prefijo, guardó %v", attrNames(item.OtherAttrs))
	}

	salida, err := MarshalAttrs(item)
	if err != nil {
		t.Fatalf("MarshalAttrs falló: %v", err)
	}
	for i, a := range salida[1:] { // salida[0] es Self, sin omitempty
		if a.Name.Space != entrada[i].Name.Space || a.Name.Local != entrada[i].Name.Local {
			t.Errorf("el prefijo no sobrevivió: %v, esperado %v", a.Name, entrada[i].Name)
		}
	}
}

// TestAttrs_DestinoInvalido comprueba los errores de uso, con mensajes que dicen qué
// se recibió.
func TestAttrs_DestinoInvalido(t *testing.T) {
	var nilPtr *itemAttrs

	tests := []struct {
		name string
		dest any
	}{
		{name: "nil", dest: nil},
		{name: "puntero nil", dest: nilPtr},
		{name: "no es struct", dest: new(int)},
		{name: "cadena", dest: "no soy un struct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UnmarshalAttrs(nil, tt.dest); err == nil {
				t.Error("UnmarshalAttrs debe rechazar este destino")
			}
			if _, err := MarshalAttrs(tt.dest); err == nil {
				t.Error("MarshalAttrs debe rechazar este origen")
			}
		})
	}
}

// TestAttrs_PunteroYValor comprueba que MarshalAttrs acepta las dos formas, porque los
// MarshalXML del repositorio tienen receptores de los dos tipos.
func TestAttrs_PunteroYValor(t *testing.T) {
	item := itemAttrs{baseAttrs: baseAttrs{Self: "u1"}}

	porValor, err := MarshalAttrs(item)
	if err != nil {
		t.Fatalf("por valor: %v", err)
	}
	porPuntero, err := MarshalAttrs(&item)
	if err != nil {
		t.Fatalf("por puntero: %v", err)
	}

	if fmt.Sprint(porValor) != fmt.Sprint(porPuntero) {
		t.Errorf("valor y puntero dan resultados distintos: %v vs %v", porValor, porPuntero)
	}
}
