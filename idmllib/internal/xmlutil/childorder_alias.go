package xmlutil

import "github.com/dimelords/idmllib/v2/internal/xmlorder"

// El registro de orden documental vive en internal/xmlorder y no aquí por un ciclo de
// importación: `pkg/common` lo necesita para el elemento `Properties`, y este paquete
// importa `pkg/common` para sus helpers de error, así que `pkg/common` no puede importar
// este paquete.
//
// `internal/xmlorder` no importa nada fuera de la biblioteca estándar, así que puede
// usarlo cualquiera. Estos dos alias existen para que los 24 sitios que ya usaban
// `xmlutil.ChildOrder` y `xmlutil.ChildEmitter` sigan funcionando sin cambiarlos. La
// documentación del mecanismo está en el paquete de destino.
type (
	// ChildOrder es xmlorder.ChildOrder.
	ChildOrder = xmlorder.ChildOrder

	// ChildEmitter es xmlorder.ChildEmitter.
	ChildEmitter = xmlorder.ChildEmitter
)
