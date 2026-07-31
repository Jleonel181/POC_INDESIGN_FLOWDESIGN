package xmlorder

// ChildOrder recuerda en qué orden aparecieron los hijos de un elemento al
// parsearlo, para poder reproducir ese orden al volver a serializarlo.
//
// El problema que resuelve: un struct de Go declara sus hijos en campos por tipo
// —una lista de capas, otra de secciones, un puntero a unas preferencias— y al
// serializar los emite en el orden en que están declarados los campos. Pero en el
// XML de entrada esos hijos venían intercalados en un orden que no está agrupado
// por tipo. Sin este registro, el ciclo parseo → serialización reagrupa los hijos
// y devuelve un documento distinto del que entró.
//
// Lo que NO es: un contenedor. El contenido de cada hijo sigue viviendo en su
// campo por tipo, que es su única fuente de verdad. Aquí solo se guarda la
// secuencia de clases. Mutar un hijo a través de su campo se refleja al
// serializar, porque Replay lee del modelo en el momento de emitir.
type ChildOrder struct {
	kinds []string
}

// ChildEmitter emite un hijo concreto. Es una función y no un valor porque cada
// tipo sabe cómo emitir el suyo: unos con encoder.Encode y otros con
// EncodeElement y un nombre de elemento explícito.
type ChildEmitter func() error

// Record anota que el siguiente hijo leído es de la clase indicada. Se llama una
// vez por hijo al parsear, en el mismo sitio donde ese hijo se guarda en su campo.
func (c *ChildOrder) Record(kind string) {
	c.kinds = append(c.kinds, kind)
}

// Recorded indica si hay orden registrado. Es falso en un modelo construido desde
// cero, que nunca se parseó.
func (c *ChildOrder) Recorded() bool {
	return len(c.kinds) > 0
}

// Reset olvida el orden registrado.
func (c *ChildOrder) Reset() {
	c.kinds = nil
}

// Kinds devuelve la secuencia registrada. Es para los tests y para diagnóstico: el
// orden de emisión sale de aquí, y poder leerlo permite comprobarlo sin serializar.
func (c ChildOrder) Kinds() []string {
	return c.kinds
}

// Replay emite los hijos en el orden registrado al parsear.
//
// children agrupa los hijos que el modelo tiene ahora, por clase y en el orden de
// su campo. fieldOrder es el orden en que están declarados los campos del struct,
// y tiene que nombrar **todas** las clases que children puede traer: una clase que
// falte en fieldOrder no se emitiría nunca si el registro no la menciona. Hay un
// test por tipo que lo comprueba.
//
// Cuando no hay orden registrado, el primer bucle no hace nada y la segunda pasada
// emite todo en el orden de los campos, que es el comportamiento de siempre para
// un modelo construido desde cero.
func (c *ChildOrder) Replay(fieldOrder []string, children map[string][]ChildEmitter) error {
	emitted := make(map[string]int, len(children))

	for _, kind := range c.kinds {
		n := emitted[kind]
		if n >= len(children[kind]) {
			// El registro menciona un hijo que el modelo ya no tiene: se eliminó
			// después de parsear. Se salta.
			continue
		}
		emitted[kind]++
		if err := children[kind][n](); err != nil {
			return err
		}
	}

	// Hijos que el modelo tiene y el registro no menciona, porque se agregaron
	// después de parsear. Se emiten al final, agrupados por clase y en el orden de
	// los campos.
	//
	// Esta pasada es lo que hace que el registro sea seguro: sin ella, agregar un
	// hijo sin actualizar el registro lo haría desaparecer al serializar, y perder
	// un elemento es mucho peor que emitirlo en una posición rara.
	for _, kind := range fieldOrder {
		for n := emitted[kind]; n < len(children[kind]); n++ {
			if err := children[kind][n](); err != nil {
				return err
			}
		}
	}

	return nil
}
