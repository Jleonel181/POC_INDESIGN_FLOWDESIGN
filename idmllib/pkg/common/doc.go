// Package common provee tipos compartidos usados en todos los paquetes de dominio IDML.
//
// Este paquete contiene tipos referenciados por múltiples paquetes de dominio
// (document, spread, story, resources) para evitar dependencias circulares y
// proveer una base estable para el sistema de tipos IDML.
//
// # Tipos compartidos
//
// RawXMLElement: Comodín compatible hacia adelante para elementos XML desconocidos.
// Usado en todo el código para preservar elementos que aún no están modelados.
//
// Properties: Contenedor común para metadatos y configuración almacenados como
// pares clave-valor en elementos Label.
//
// GridDataInformation: Configuración de grilla compartida entre los tipos
// Document (NamedGrid) y Spread (Page).
//
// # Uso
//
// Los paquetes de dominio importan common/ para acceder a estos tipos:
//
//	import "github.com/dimelords/idmllib/v2/pkg/common"
//
//	type MyType struct {
//	    Properties *common.Properties
//	    OtherElements []common.RawXMLElement
//	}
//
// # Arquitectura
//
// El paquete common es parte de la refactorización de arquitectura del Epic 5 que divide
// pkg/idml en paquetes específicos por dominio. Ver docs/EPIC-5-REFACTORING-ANALYSIS.md
// para decisiones de diseño detalladas.
package common
