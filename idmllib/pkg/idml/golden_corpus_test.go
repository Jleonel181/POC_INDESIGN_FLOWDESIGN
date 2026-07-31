package idml

import (
	"os"
	"testing"
)

// TestCorpusPackages_RutasResolublesYDocumentadas comprueba lo que el Req 1,
// criterio 3 exige de cada elemento del corpus: que su ruta salga de una constante y
// que una variable de entorno pueda sustituirla.
//
// El descuido que evita es concreto: al agregar un origen es fácil olvidar la
// variable de entorno, y entonces ese origen no se puede apuntar a otra copia sin
// recompilar. Aquí falla en lugar de pasar inadvertido.
func TestCorpusPackages_RutasResolublesYDocumentadas(t *testing.T) {
	if len(corpusPackages) == 0 {
		t.Fatal("el corpus no declara ningún paquete")
	}

	seenOrigin := map[string]bool{}
	seenEnv := map[string]bool{}

	for _, cp := range corpusPackages {
		if cp.origin == "" {
			t.Error("un paquete del corpus no declara origen")
		}
		if seenOrigin[cp.origin] {
			t.Errorf("el origen %q está repetido: los resúmenes serían indistinguibles", cp.origin)
		}
		seenOrigin[cp.origin] = true

		if cp.def == "" {
			t.Errorf("el origen %q no declara ruta por defecto", cp.origin)
		}
		if cp.env == "" {
			t.Errorf("el origen %q no declara variable de entorno, así que no se puede apuntar a otra copia sin recompilar", cp.origin)
		}
		if seenEnv[cp.env] {
			t.Errorf("la variable %q está repetida: sustituiría dos orígenes a la vez", cp.env)
		}
		seenEnv[cp.env] = true

		// La variable de entorno tiene prioridad sobre la constante.
		t.Setenv(cp.env, "/ruta/de/prueba.idml")
		if got := corpusPath(cp.env, cp.def); got != "/ruta/de/prueba.idml" {
			t.Errorf("%s: %s no sustituye la ruta, se resolvió %q", cp.origin, cp.env, got)
		}
	}
}

// TestCorpusPackages_ElCorpusEstaCompleto avisa si alguno de los cinco elementos
// falta en la copia de trabajo. No falla: es un aviso, porque el arnés ya omite por
// separado lo que no está.
func TestCorpusPackages_ElCorpusEstaCompleto(t *testing.T) {
	missing := 0

	if _, err := os.Stat(corpusPath(EnvCorpusReferenceDir, CorpusReferenceDir)); err != nil {
		t.Logf("ausente: documento_referencia (%v)", err)
		missing++
	}
	for _, cp := range corpusPackages {
		if _, err := os.Stat(corpusPath(cp.env, cp.def)); err != nil {
			t.Logf("ausente: %s (%v)", cp.origin, err)
			missing++
		}
	}

	total := 1 + len(corpusPackages)
	t.Logf("corpus: %d de %d elementos presentes", total-missing, total)
	if missing > 0 {
		t.Logf("el arnés mide sobre menos documentos de los disponibles, así que un «cero diferencias» vale menos")
	}
}

// TestAggregateTallies_SumaPorCategoria cubre el criterio 4: el total agregado es la
// suma de los desgloses. Se prueba aparte porque en la ejecución real los números
// dependen del corpus, y aquí se fijan.
func TestAggregateTallies_SumaPorCategoria(t *testing.T) {
	a := newFidelityTally("a")
	a.clean, a.differing, a.unparsed, a.failed, a.truncated = 1, 2, 3, 4, 5
	a.byCategory["atributo-ausente"] = 10
	a.byCategory["orden-elementos-distinto"] = 1

	b := newFidelityTally("b")
	b.clean, b.differing, b.unparsed, b.failed, b.truncated = 10, 20, 30, 40, 50
	b.byCategory["atributo-ausente"] = 5
	b.byCategory["texto-distinto"] = 7

	total := aggregateTallies([]*fidelityTally{a, b})

	if total.clean != 11 || total.differing != 22 || total.unparsed != 33 || total.failed != 44 || total.truncated != 55 {
		t.Errorf("conteos de archivo mal sumados: %+v", total)
	}
	if total.total() != a.total()+b.total() {
		t.Errorf("total() = %d, esperado %d", total.total(), a.total()+b.total())
	}

	want := map[string]int{
		"atributo-ausente":         15,
		"orden-elementos-distinto": 1,
		"texto-distinto":           7,
	}
	for category, n := range want {
		if total.byCategory[category] != n {
			t.Errorf("categoría %q: %d, esperado %d", category, total.byCategory[category], n)
		}
	}
	if len(total.byCategory) != len(want) {
		t.Errorf("el total tiene %d categorías, esperadas %d", len(total.byCategory), len(want))
	}
}

// TestAggregateTallies_SinOrigenes comprueba el caso de corpus vacío.
func TestAggregateTallies_SinOrigenes(t *testing.T) {
	total := aggregateTallies(nil)
	if total.total() != 0 || len(total.byCategory) != 0 {
		t.Errorf("agregar cero orígenes debe dar un total vacío, dio %+v", total)
	}
}
