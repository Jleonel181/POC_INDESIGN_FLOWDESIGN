package idgen

import (
	"regexp"
	"sync"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// formatoSelf es el criterio 1: `u` más 1 a 8 dígitos hexadecimales en minúscula.
var formatoSelf = regexp.MustCompile(`^u[0-9a-f]{1,8}$`)

// TestGenerate_Formato comprueba el criterio 1 sobre una tirada larga.
func TestGenerate_Formato(t *testing.T) {
	r := New()
	for i := 0; i < 10000; i++ {
		id := r.Generate()
		if !formatoSelf.MatchString(id) {
			t.Fatalf("el identificador %d no cumple el formato: %q", i, id)
		}
	}
}

// TestGenerate_SinDuplicados comprueba el criterio 2: 10.000 identificadores, todos
// distintos. Es la cifra que pide la tarea.
func TestGenerate_SinDuplicados(t *testing.T) {
	const n = 10000
	r := New()
	vistos := make(map[string]int, n)

	for i := 0; i < n; i++ {
		id := r.Generate()
		if antes, dup := vistos[id]; dup {
			t.Fatalf("el identificador %q salió dos veces, en la iteración %d y en la %d", id, antes, i)
		}
		vistos[id] = i
	}
	if len(vistos) != n {
		t.Errorf("se esperaban %d identificadores distintos, hay %d", n, len(vistos))
	}
	if r.Len() != n {
		t.Errorf("el registro debería tener %d identificadores, tiene %d", n, r.Len())
	}
}

// TestGenerate_NoColisionaConLoRegistrado comprueba el criterio 2 en su parte difícil: los
// identificadores generados esquivan los que ya venían de un documento.
//
// Se registran a propósito los que el contador iba a producir, para forzar la colisión en
// lugar de esperar a que ocurra por casualidad.
func TestGenerate_NoColisionaConLoRegistrado(t *testing.T) {
	r := New()

	// Justo los 50 primeros que el contador quiere emitir: u1, u2, ... u32.
	ocupados := make(map[string]bool, 50)
	for i := 1; i <= 50; i++ {
		id := "u" + hexDe(uint64(i))
		if err := r.Register(id); err != nil {
			t.Fatalf("Register(%q): %v", id, err)
		}
		ocupados[id] = true
	}

	for i := 0; i < 200; i++ {
		id := r.Generate()
		if ocupados[id] {
			t.Fatalf("el identificador generado %q colisiona con uno registrado", id)
		}
		if !formatoSelf.MatchString(id) {
			t.Fatalf("tras esquivar colisiones el formato se rompió: %q", id)
		}
	}
}

// TestRegister_RechazaDuplicadoYVacio comprueba el criterio 4 y la validación de entrada.
func TestRegister_RechazaDuplicadoYVacio(t *testing.T) {
	r := New()

	if err := r.Register(""); err == nil {
		t.Error("Register(\"\") debería dar error")
	}
	if r.Len() != 0 {
		t.Errorf("un Register con error no debería dejar rastro, hay %d identificadores", r.Len())
	}

	if err := r.Register("uce7"); err != nil {
		t.Fatalf("el primer Register debería funcionar: %v", err)
	}
	if err := r.Register("uce7"); err == nil {
		t.Error("registrar dos veces el mismo identificador debería dar error")
	}
	if r.Len() != 1 {
		t.Errorf("el duplicado no debería contar, hay %d identificadores", r.Len())
	}
}

// TestRegister_AceptaLasFormasRealesDeSelf comprueba el criterio 3. Las cuatro formas están
// medidas en el Documento_Referencia; si Register exigiera el formato de Generate, no se
// podría registrar el contenido de un documento real.
func TestRegister_AceptaLasFormasRealesDeSelf(t *testing.T) {
	reales := []string{
		"uce7",                               // u + hexadecimal
		"Color/Black",                        // con barra
		"ObjectStyle/Naviga%3aStandard",      // con barra y escape
		"u1de1ColorGroupSwatch3",             // compuesto
		"di118FontnIvyEpic",                  // prefijo di
		"d",                                  // el del propio Document
		"dTextVariablenModificationDatensrc", // largo
	}
	r := New()
	for _, id := range reales {
		if err := r.Register(id); err != nil {
			t.Errorf("Register(%q) debería aceptarse, es una forma real: %v", id, err)
		}
	}
	if r.Len() != len(reales) {
		t.Errorf("se esperaban %d registrados, hay %d", len(reales), r.Len())
	}
}

// TestGenerate_ConcurrenteSinDuplicados comprueba el criterio 5. Con -race este test es lo
// que delata un acceso sin proteger.
func TestGenerate_ConcurrenteSinDuplicados(t *testing.T) {
	const goroutines = 16
	const porGoroutine = 500
	const total = goroutines * porGoroutine

	r := New()
	var wg sync.WaitGroup
	salida := make([][]string, goroutines)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			ids := make([]string, porGoroutine)
			for i := range ids {
				ids[i] = r.Generate()
			}
			salida[g] = ids
		}(g)
	}
	wg.Wait()

	vistos := make(map[string]struct{}, total)
	for _, ids := range salida {
		for _, id := range ids {
			if _, dup := vistos[id]; dup {
				t.Fatalf("el identificador %q se generó dos veces en concurrencia", id)
			}
			vistos[id] = struct{}{}
		}
	}
	if len(vistos) != total {
		t.Errorf("se esperaban %d identificadores distintos, hay %d", total, len(vistos))
	}
}

// TestRegister_ConcurrenteUnSoloGanador comprueba que, si varias goroutines registran el
// mismo identificador, exactamente una lo consigue.
func TestRegister_ConcurrenteUnSoloGanador(t *testing.T) {
	const goroutines = 32
	r := New()

	var wg sync.WaitGroup
	exitos := make([]bool, goroutines)
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			exitos[g] = r.Register("uce7") == nil
		}(g)
	}
	wg.Wait()

	n := 0
	for _, ok := range exitos {
		if ok {
			n++
		}
	}
	if n != 1 {
		t.Errorf("exactamente una goroutine debería registrar el identificador, lo consiguieron %d", n)
	}
}

// TestPropiedades_Registry son las propiedades del registro con gopter, la herramienta que
// ya usa el repositorio.
func TestPropiedades_Registry(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("todo lo generado cumple el formato y es nuevo", prop.ForAll(
		func(n uint8) bool {
			r := New()
			vistos := make(map[string]struct{})
			for i := 0; i <= int(n); i++ {
				id := r.Generate()
				if !formatoSelf.MatchString(id) {
					return false
				}
				if _, dup := vistos[id]; dup {
					return false
				}
				vistos[id] = struct{}{}
			}
			return len(vistos) == int(n)+1
		},
		gen.UInt8(),
	))

	properties.Property("lo generado nunca coincide con lo registrado antes", prop.ForAll(
		func(ids []string) bool {
			r := New()
			registrados := make(map[string]struct{})
			for _, id := range ids {
				if id == "" {
					continue
				}
				if err := r.Register(id); err == nil {
					registrados[id] = struct{}{}
				}
			}
			for i := 0; i < 100; i++ {
				id := r.Generate()
				if _, choca := registrados[id]; choca {
					return false
				}
			}
			return true
		},
		// Identificadores con la misma forma que produce Generate, que es donde la
		// colisión es posible de verdad.
		gen.SliceOf(gen.RegexMatch(`^u[0-9a-f]{1,3}$`)),
	))

	properties.Property("registrar dos veces siempre falla la segunda", prop.ForAll(
		func(id string) bool {
			if id == "" {
				return true
			}
			r := New()
			if err := r.Register(id); err != nil {
				return false
			}
			return r.Register(id) != nil
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// hexDe formatea en hexadecimal minúscula, igual que Generate.
func hexDe(n uint64) string {
	const hex = "0123456789abcdef"
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{hex[n&0xf]}, b...)
		n >>= 4
	}
	return string(b)
}
