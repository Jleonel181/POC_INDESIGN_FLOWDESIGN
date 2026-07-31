package spread

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// TextCapacityInfo contiene información de capacidad de text frame extraída desde TextFramePreference.
// Esto proporciona las dimensiones de texto realmente utilizables considerando columnas, gutters e insets.
type TextCapacityInfo struct {
	// ColumnWidth es el ancho de cada columna de texto en puntos
	ColumnWidth float64

	// ColumnCount es el número de columnas de texto
	ColumnCount int

	// ColumnGutter es el espacio entre columnas en puntos
	ColumnGutter float64

	// InsetSpacing define el espaciado interno [arriba, izquierda, abajo, derecha] en puntos
	InsetSpacing [4]float64

	// EffectiveWidth es el ancho total utilizable para texto en todas las columnas
	// Fórmula: (ColumnWidth × ColumnCount) + (ColumnGutter × (ColumnCount - 1))
	EffectiveWidth float64

	// GeometricWidth es el ancho total del frame desde GeometricBounds (para referencia)
	GeometricWidth float64
}

// TextCapacity extrae información de capacidad de text frame desde TextFramePreference.
// Retorna nil si TextFramePreference no se encuentra o el parseo falla.
//
// El TextCapacityInfo retornado proporciona dimensiones de texto precisas que InDesign usa internamente,
// lo cual es crítico para cálculos precisos de ajuste de texto. Esto considera:
// - Múltiples columnas y gutters
// - Insets internos del frame (espaciado)
// - Anchos de columna fijos vs. layouts flexibles
//
// Ejemplo:
//
//	capacity := frame.TextCapacity()
//	if capacity != nil {
//	    fmt.Printf("Ancho de texto utilizable: %.2fpt en %d columna(s)\n",
//	        capacity.EffectiveWidth, capacity.ColumnCount)
//	}
func (f *SpreadTextFrame) TextCapacity() *TextCapacityInfo {
	info := &TextCapacityInfo{
		ColumnCount: 1, // Default to single column
	}

	// Obtener bounds geométricos para referencia
	if bounds, err := f.Bounds(); err == nil {
		info.GeometricWidth = bounds.Width
	}

	// Parsear TextFramePreference desde OtherElements
	var foundPreference bool
	for _, elem := range f.OtherElements {
		if elem.XMLName.Local == "TextFramePreference" {
			foundPreference = true

			// Extract attributes from TextFramePreference
			for _, attr := range elem.Attrs {
				switch attr.Name.Local {
				case "TextColumnFixedWidth":
					// Este es el ancho real que InDesign usa para el layout de texto
					if width, err := strconv.ParseFloat(attr.Value, 64); err == nil {
						info.ColumnWidth = width
					}

				case "TextColumnCount":
					if count, err := strconv.Atoi(attr.Value); err == nil && count > 0 {
						info.ColumnCount = count
					}

				case "TextColumnGutter":
					if gutter, err := strconv.ParseFloat(attr.Value, 64); err == nil {
						info.ColumnGutter = gutter
					}
				}
			}

			// Parsear Properties anidadas para InsetSpacing desde elem.Content
			info.InsetSpacing = parseInsetSpacingFromContent(elem.Content)

			break
		}
	}

	if !foundPreference {
		return nil
	}

	// Calcular ancho efectivo
	// Si tenemos TextColumnFixedWidth, usarlo directamente
	if info.ColumnWidth > 0 {
		// Ancho efectivo = todas las columnas + gutters entre ellas
		info.EffectiveWidth = (info.ColumnWidth * float64(info.ColumnCount)) +
			(info.ColumnGutter * float64(info.ColumnCount-1))
	} else if info.GeometricWidth > 0 {
		// Alternativa: usar ancho geométrico menos insets
		usableWidth := info.GeometricWidth - info.InsetSpacing[1] - info.InsetSpacing[3] // left + right

		// Distribuir en columnas
		if info.ColumnCount > 1 {
			totalGutter := info.ColumnGutter * float64(info.ColumnCount-1)
			info.ColumnWidth = (usableWidth - totalGutter) / float64(info.ColumnCount)
		} else {
			info.ColumnWidth = usableWidth
		}

		info.EffectiveWidth = usableWidth
	}

	return info
}

// Estructuras helper para parsear XML de InsetSpacing
type propertiesContainer struct {
	InsetSpacing *insetSpacingList `xml:"InsetSpacing"`
}

type insetSpacingList struct {
	Items []insetListItem `xml:"ListItem"`
}

type insetListItem struct {
	Value string `xml:",chardata"`
}

// parseInsetSpacingFromContent extrae InsetSpacing desde contenido XML crudo.
// Formato de InsetSpacing: <Properties><InsetSpacing type="list"><ListItem>0</ListItem>...</InsetSpacing></Properties>
// Retorna [arriba, izquierda, abajo, derecha]
func parseInsetSpacingFromContent(content []byte) [4]float64 {
	var insets [4]float64

	if len(content) == 0 {
		return insets
	}

	// Parsear el contenido XML
	var props propertiesContainer
	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	if err := decoder.Decode(&props); err != nil {
		return insets
	}

	// Extraer valores desde ListItems
	if props.InsetSpacing != nil {
		for i, item := range props.InsetSpacing.Items {
			if i >= 4 {
				break
			}
			if val, err := strconv.ParseFloat(strings.TrimSpace(item.Value), 64); err == nil {
				insets[i] = val
			}
		}
	}

	return insets
}
