package idml

import (
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/resources"
)

// GetFontPostScriptName busca el nombre PostScript de una fuente en Fonts.xml.
// Útil para coincidencia precisa de fuentes en sistemas tipográficos como HarfBuzz.
//
// Parámetros:
//   - fontFamily: El nombre de la familia tipográfica (ej. "Helvetica", "Arial")
//   - fontStyle: El nombre del estilo de fuente (ej. "Regular", "Bold", "Italic")
//
// Devuelve el nombre PostScript (ej. "Helvetica-Bold") o cadena vacía si no se encuentra.
func (p *Package) GetFontPostScriptName(fontFamily, fontStyle string) (string, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return "", common.WrapError("idml", "get font postscript name", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	// Buscar entre las familias tipográficas
	for _, family := range fonts.FontFamilies {
		if family.Name == fontFamily {
			// Buscar entre las fuentes de esta familia
			for _, font := range family.Fonts {
				// Coincidir por FontStyleName
				if font.FontStyleName == fontStyle {
					return font.PostScriptName, nil
				}
			}
			// Familia encontrada pero estilo no coincide
			return "", common.WrapError("idml", "get font postscript name", fmt.Errorf("font style %q not found in family %q", fontStyle, fontFamily))
		}
	}

	// Familia no encontrada
	return "", common.WrapError("idml", "get font postscript name", fmt.Errorf("font family %q not found", fontFamily))
}

// GetFontByStyle obtiene una fuente específica de Fonts.xml por familia y estilo.
// Provee acceso a todos los metadatos de la fuente incluyendo Status, FontType, etc.
//
// Parámetros:
//   - fontFamily: El nombre de la familia tipográfica (ej. "Helvetica", "Arial")
//   - fontStyle: El nombre del estilo de fuente (ej. "Regular", "Bold", "Italic")
//
// Devuelve el struct Font o error si no se encuentra.
func (p *Package) GetFontByStyle(fontFamily, fontStyle string) (*resources.Font, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return nil, common.WrapError("idml", "get font by style", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	// Buscar entre las familias tipográficas
	for _, family := range fonts.FontFamilies {
		if family.Name == fontFamily {
			// Buscar entre las fuentes de esta familia
			for i := range family.Fonts {
				if family.Fonts[i].FontStyleName == fontStyle {
					return &family.Fonts[i], nil
				}
			}
			// Familia encontrada pero estilo no coincide
			return nil, common.WrapError("idml", "get font by style", fmt.Errorf("font style %q not found in family %q", fontStyle, fontFamily))
		}
	}

	// Familia no encontrada
	return nil, common.WrapError("idml", "get font by style", fmt.Errorf("font family %q not found", fontFamily))
}

// ListFontFamilies devuelve todos los nombres de familias tipográficas disponibles en el documento.
// Útil para descubrimiento y validación de fuentes.
func (p *Package) ListFontFamilies() ([]string, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return nil, common.WrapError("idml", "list font families", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	families := make([]string, 0, len(fonts.FontFamilies))
	for _, family := range fonts.FontFamilies {
		families = append(families, family.Name)
	}

	return families, nil
}

// ListFontStyles devuelve todos los estilos disponibles para una familia tipográfica dada.
// Útil para descubrir los pesos y estilos disponibles.
func (p *Package) ListFontStyles(fontFamily string) ([]string, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return nil, common.WrapError("idml", "list font styles", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	for _, family := range fonts.FontFamilies {
		if family.Name == fontFamily {
			styles := make([]string, 0, len(family.Fonts))
			for _, font := range family.Fonts {
				styles = append(styles, font.FontStyleName)
			}
			return styles, nil
		}
	}

	return nil, common.WrapError("idml", "list font styles", fmt.Errorf("font family %q not found", fontFamily))
}
