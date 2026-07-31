// Package idgen genera identificadores únicos para los elementos de un documento IDML.
//
// En IDML casi todo elemento lleva un atributo `Self` que lo identifica, y otros elementos
// lo referencian por ese valor: un marco de texto apunta a su story con `ParentStory`, una
// página a su maqueta con `AppliedMaster`. Dos elementos con el mismo `Self` producen un
// documento que InDesign no puede resolver, así que generar identificadores nuevos sin
// chocar con los que ya trae un documento es un requisito, no una comodidad.
//
// El registro sirve para las dos cosas a la vez: se le dan los identificadores que ya
// existen en el paquete abierto con Register, y luego Generate produce nuevos que no
// colisionan ni con esos ni entre sí.
package idgen

import (
	"strconv"
	"sync"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Registry lleva la cuenta de los identificadores en uso y produce nuevos.
//
// Es seguro usarlo desde varias goroutines: todas sus operaciones toman un mutex interno.
// El valor cero no sirve, hay que construirlo con New.
type Registry struct {
	mu sync.Mutex

	// enUso son todos los identificadores conocidos, tanto los registrados desde un
	// documento existente como los ya generados. Es un conjunto: solo importa la
	// pertenencia.
	enUso map[string]struct{}

	// siguiente es el contador secuencial del que sale el próximo identificador. Empieza
	// en 1 para que el primero sea "u1" y no "u0".
	siguiente uint64
}

// New crea un registro vacío.
func New() *Registry {
	return &Registry{
		enUso:     make(map[string]struct{}),
		siguiente: 1,
	}
}

// Register anota un identificador que ya está en uso, típicamente leído de un documento
// que se acaba de abrir, para que Generate no lo vuelva a producir.
//
// No valida la forma del identificador a propósito. Los `Self` que escribe InDesign tienen
// varias formas y solo una es la que produce Generate: medidos en el Documento_Referencia,
// 468 son `u` más hexadecimal, 346 llevan barra como `Color/Black` u
// `ObjectStyle/Naviga%3aStandard`, y el resto son compuestos como `u1de1ColorGroupSwatch3`
// o `di118FontnIvyEpic`. Exigir un formato aquí haría imposible registrar los
// identificadores de un documento real, que es justo para lo que existe este método.
//
// Devuelve error si el identificador está vacío o si ya estaba registrado. Lo segundo es el
// criterio 4 de la tarea, y tiene un motivo: un duplicado en la entrada significa que el
// documento de origen ya era ambiguo, o que el llamador está registrando dos veces, y las
// dos cosas se quieren saber antes de generar nada a partir de ahí.
func (r *Registry) Register(id string) error {
	if id == "" {
		return common.Errorf("idgen", "register id", "", "el identificador está vacío")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, yaEstaba := r.enUso[id]; yaEstaba {
		return common.Errorf("idgen", "register id", id, "el identificador ya estaba registrado")
	}
	r.enUso[id] = struct{}{}
	return nil
}

// Registered indica si un identificador está en uso, ya sea porque se registró o porque se
// generó.
func (r *Registry) Registered(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, está := r.enUso[id]
	return está
}

// Generate devuelve un identificador nuevo con la forma `u` seguido de 1 a 8 dígitos
// hexadecimales en minúscula, como los `uce7` y `u1bc8` que escribe InDesign.
//
// El identificador devuelto queda registrado, así que nunca se repite. Si el contador cae
// en un valor que ya estaba registrado desde un documento existente, se salta y sigue: por
// eso el resultado no es necesariamente consecutivo respecto al anterior.
//
// ponytail: el contador es de 64 bits pero el formato admite 8 dígitos hexadecimales, así
// que a partir de 0xffffffff identificadores generados la forma dejaría de cumplirse. Son
// 4.294.967.295, y el conjunto `enUso` necesitaría cientos de gigabytes mucho antes de
// llegar ahí, así que el límite lo pone la memoria y no este contador. La vía de mejora, si
// algún día un caso de uso se acerca, es pasar a 1 a 10 dígitos, que es lo que InDesign
// acepta de hecho, o devolver error al agotarse.
func (r *Registry) Generate() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		id := "u" + strconv.FormatUint(r.siguiente, 16)
		r.siguiente++

		if _, ocupado := r.enUso[id]; ocupado {
			continue
		}
		r.enUso[id] = struct{}{}
		return id
	}
}

// Len devuelve cuántos identificadores hay en uso. Es para diagnóstico y para los tests.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.enUso)
}
