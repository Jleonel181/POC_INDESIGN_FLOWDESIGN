// Package common contiene tipos compartidos usados en todos los paquetes de dominio IDML.
//
// Estos tipos fueron extraídos de pkg/idml para evitar dependencias circulares
// y proveer una base común para los paquetes document, spread, story y resources.
//
// Tipos clave:
//   - RawXMLElement: Comodín compatible hacia adelante para elementos XML desconocidos
//   - Properties: Contenedor de propiedades comunes con pares clave-valor Label
//   - GridDataInformation: Configuración de grilla compartida por Document y Spread
package common

import (
	"encoding/xml"
)

// RawXMLElement representa un elemento XML arbitrario que aún no fue modelado explícitamente.
// Permite compatibilidad hacia adelante preservando elementos desconocidos durante marshal/unmarshal.
//
// Ejemplo de uso en campos comodín:
//
//	OtherElements []RawXMLElement `xml:",any"`
//
// Captura elementos como KinsokuTable, MojikumiTable, TextVariable, etc.
// hasta que se agreguen definiciones de struct explícitas para ellos.
type RawXMLElement struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
}

// Properties representa el elemento contenedor Properties.
// Almacena metadatos y configuración como pares clave-valor en Label.
type Properties struct {
	XMLName xml.Name `xml:"Properties"`

	PathGeometry *PathGeometry `xml:"PathGeometry,omitempty"`

	// Label contiene pares clave-valor
	Label *Label `xml:"Label,omitempty"`

	// Comodín para otros hijos de Properties (AppliedMathMLSwatch, etc.)
	// que aún no están modelados explícitamente
	OtherElements []RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// Label representa un contenedor de pares clave-valor.
type Label struct {
	XMLName       xml.Name       `xml:"Label"`
	KeyValuePairs []KeyValuePair `xml:"KeyValuePair"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// KeyValuePair representa un único par clave-valor de metadatos.
type KeyValuePair struct {
	XMLName xml.Name `xml:"KeyValuePair"`
	Key     string   `xml:"Key,attr"`
	Value   string   `xml:"Value,attr"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PathGeometry representa información de geometría de un path.
type PathGeometry struct {
	GeometryPathType *GeometryPathType `xml:"GeometryPathType,omitempty"`
}

// GeometryPathType define un path geométrico con puntos.
type GeometryPathType struct {
	PathOpen       string          `xml:"PathOpen,attr,omitempty"`
	PathPointArray *PathPointArray `xml:"PathPointArray,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PathPointArray contiene un arreglo de puntos de path.
type PathPointArray struct {
	PathPoints []PathPointType `xml:"PathPointType"`
}

// PathPointType representa un único punto en un path con ancla y manejadores de dirección.
type PathPointType struct {
	Anchor         string `xml:"Anchor,attr"`
	LeftDirection  string `xml:"LeftDirection,attr,omitempty"`
	RightDirection string `xml:"RightDirection,attr,omitempty"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// GetAppliedFont extrae el valor de AppliedFont desde Properties.OtherElements.
// Retorna el nombre de la familia tipográfica (ej: "Polaris Condensed") o cadena vacía si no se encuentra.
//
// AppliedFont se almacena en el elemento Properties así:
//
//	<AppliedFont type="string">Polaris Condensed</AppliedFont>
func (p *Properties) GetAppliedFont() string {
	if p == nil {
		return ""
	}

	// Buscar el elemento <AppliedFont> en OtherElements
	for _, elem := range p.OtherElements {
		if elem.XMLName.Local == "AppliedFont" {
			// Extraer contenido de texto
			content := string(elem.Content)
			// Parsear el contenido para extraer solo el texto (ignorar XML anidado)
			// Para contenido de texto simple, esto funciona directamente
			if len(content) > 0 {
				// Encontrar texto entre > y <
				start := 0
				end := len(content)
				for i, ch := range content {
					if ch == '>' {
						start = i + 1
					} else if ch == '<' {
						end = i
						break
					}
				}
				if start < end {
					return content[start:end]
				}
				// Si no hay corchetes angulares, es texto plano
				return content
			}
		}
	}

	return ""
}

// GetBasedOn extrae el valor de BasedOn desde Properties.OtherElements.
// Retorna el ID del estilo padre o cadena vacía si no se encuentra.
//
// BasedOn se almacena en el elemento Properties así:
//
//	<BasedOn type="string">$ID/[No character style]</BasedOn>
func (p *Properties) GetBasedOn() string {
	if p == nil {
		return ""
	}

	// Buscar el elemento <BasedOn> en OtherElements
	for _, elem := range p.OtherElements {
		if elem.XMLName.Local == "BasedOn" {
			// Extraer contenido de texto
			content := string(elem.Content)
			// Parsear el contenido para extraer solo el texto (ignorar XML anidado)
			if len(content) > 0 {
				// Encontrar texto entre > y <
				start := 0
				end := len(content)
				for i, ch := range content {
					if ch == '>' {
						start = i + 1
					} else if ch == '<' {
						end = i
						break
					}
				}
				if start < end {
					return content[start:end]
				}
				// Si no hay corchetes angulares, es texto plano
				return content
			}
		}
	}

	return ""
}

// GridDataInformation contiene la configuración detallada de una grilla.
// Define el espaciado entre caracteres/líneas, alineación y configuración tipográfica.
type GridDataInformation struct {
	XMLName xml.Name `xml:"GridDataInformation"`

	// Tipografía
	FontStyle string `xml:"FontStyle,attr,omitempty"` // Estilo de fuente (ej: "Roman")
	PointSize string `xml:"PointSize,attr,omitempty"` // Tamaño de fuente en puntos

	// Espaciado de caracteres (Aki = espacio en japonés)
	CharacterAki string `xml:"CharacterAki,attr,omitempty"` // Espacio entre caracteres
	LineAki      string `xml:"LineAki,attr,omitempty"`      // Espacio entre líneas

	// Escala
	HorizontalScale string `xml:"HorizontalScale,attr,omitempty"` // Porcentaje de escala horizontal
	VerticalScale   string `xml:"VerticalScale,attr,omitempty"`   // Porcentaje de escala vertical

	// Alineación
	LineAlignment      string `xml:"LineAlignment,attr,omitempty"`      // Alineación de línea (ej: "LeftOrTopLineJustify")
	GridAlignment      string `xml:"GridAlignment,attr,omitempty"`      // Alineación de grilla (ej: "AlignEmCenter")
	CharacterAlignment string `xml:"CharacterAlignment,attr,omitempty"` // Alineación de carácter (ej: "AlignEmCenter")

	// Properties puede contener AppliedFont y otras configuraciones
	Properties *Properties `xml:"Properties,omitempty"`

	// Comodín para otros hijos de GridDataInformation
	OtherElements []RawXMLElement `xml:",any"`

	// OtherAttrs conserva los atributos que este tipo todavía no declara. Ver el
	// patrón OtherAttrs en ARCHITECTURE.md y docs/FIDELIDAD.md.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}
