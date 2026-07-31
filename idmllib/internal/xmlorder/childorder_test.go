package xmlorder

import (
	"strings"
	"testing"
)

// emitterFor construye un emisor que anota su etiqueta en el registro compartido.
func emitterFor(log *[]string, label string) ChildEmitter {
	return func() error {
		*log = append(*log, label)
		return nil
	}
}

func TestChildOrder_ReplayRespetaElOrdenRegistrado(t *testing.T) {
	var order ChildOrder
	// Entrada: a b a c, con dos "a" intercaladas, que es el caso que el orden de
	// campos destruiría al agrupar.
	order.Record("a")
	order.Record("b")
	order.Record("a")
	order.Record("c")

	var got []string
	children := map[string][]ChildEmitter{
		"a": {emitterFor(&got, "a0"), emitterFor(&got, "a1")},
		"b": {emitterFor(&got, "b0")},
		"c": {emitterFor(&got, "c0")},
	}

	if err := order.Replay([]string{"a", "b", "c"}, children); err != nil {
		t.Fatalf("Replay falló: %v", err)
	}

	want := "a0 b0 a1 c0"
	if strings.Join(got, " ") != want {
		t.Errorf("orden emitido: %q, esperado %q", strings.Join(got, " "), want)
	}
}

func TestChildOrder_SinRegistroUsaElOrdenDeCampos(t *testing.T) {
	var order ChildOrder
	if order.Recorded() {
		t.Error("un ChildOrder recién creado no debe tener orden registrado")
	}

	var got []string
	children := map[string][]ChildEmitter{
		"a": {emitterFor(&got, "a0"), emitterFor(&got, "a1")},
		"b": {emitterFor(&got, "b0")},
	}

	if err := order.Replay([]string{"a", "b"}, children); err != nil {
		t.Fatalf("Replay falló: %v", err)
	}

	want := "a0 a1 b0"
	if strings.Join(got, " ") != want {
		t.Errorf("orden emitido: %q, esperado %q", strings.Join(got, " "), want)
	}
}

func TestChildOrder_HijoAgregadoDespuesNoSePierde(t *testing.T) {
	var order ChildOrder
	order.Record("a")
	order.Record("b")

	// El modelo tiene una "a" más de las que el registro menciona: alguien la agregó
	// después de parsear. Tiene que emitirse igualmente.
	var got []string
	children := map[string][]ChildEmitter{
		"a": {emitterFor(&got, "a0"), emitterFor(&got, "a1")},
		"b": {emitterFor(&got, "b0")},
	}

	if err := order.Replay([]string{"a", "b"}, children); err != nil {
		t.Fatalf("Replay falló: %v", err)
	}

	if strings.Join(got, " ") != "a0 b0 a1" {
		t.Errorf("orden emitido: %q, esperado %q", strings.Join(got, " "), "a0 b0 a1")
	}
	if len(got) != 3 {
		t.Errorf("se emitieron %d hijos, esperados 3: un hijo agregado tras el parseo no puede perderse", len(got))
	}
}

func TestChildOrder_HijoEliminadoDespuesNoRompe(t *testing.T) {
	var order ChildOrder
	order.Record("a")
	order.Record("a")
	order.Record("b")

	// El modelo se quedó con una sola "a": la otra se eliminó tras parsear.
	var got []string
	children := map[string][]ChildEmitter{
		"a": {emitterFor(&got, "a0")},
		"b": {emitterFor(&got, "b0")},
	}

	if err := order.Replay([]string{"a", "b"}, children); err != nil {
		t.Fatalf("Replay falló: %v", err)
	}

	if strings.Join(got, " ") != "a0 b0" {
		t.Errorf("orden emitido: %q, esperado %q", strings.Join(got, " "), "a0 b0")
	}
}

func TestChildOrder_ResetOlvidaElOrden(t *testing.T) {
	var order ChildOrder
	order.Record("a")
	if !order.Recorded() {
		t.Fatal("Recorded debe ser true tras Record")
	}
	order.Reset()
	if order.Recorded() {
		t.Error("Recorded debe ser false tras Reset")
	}
}
