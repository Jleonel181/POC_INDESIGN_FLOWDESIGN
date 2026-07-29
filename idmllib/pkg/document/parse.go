package document

import (
	"encoding/xml"
	"io"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// ParseDocument parsea un archivo designmap.xml en un struct Document.
//
// Este parser de Fase 2 extrae datos estructurados del elemento Document
// mientras preserva todos los elementos desconocidos para compatibilidad hacia adelante.
//
// Con UnmarshalXML custom, las declaraciones de namespace ahora se manejan automáticamente.
func ParseDocument(data []byte) (*Document, error) {
	// Verificar que el input no sea nil
	if data == nil {
		return nil, common.Errorf("document", "parse document", "", "input data is nil")
	}

	// Verificar que el input no esté vacío
	if len(data) == 0 {
		return nil, common.Errorf("document", "parse document", "", "input data is empty")
	}

	var doc Document
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, common.WrapError("document", "parse document", err)
	}

	return &doc, nil
}

// MarshalDocument serializa un struct Document de vuelta a bytes XML.
// La salida incluye el encabezado XML estándar.
func MarshalDocument(doc *Document) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(doc, "", "\t")
}

// UnmarshalXML implementa deserialización XML custom para Document.
// Nos da control total sobre el manejo de namespaces y la lógica de parseo,
// eliminando la necesidad de parseo manual de strings para xmlns:idPkg.
func (d *Document) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	// Inicializar el documento
	d.XMLName = start.Name

	// Extraer atributos, incluyendo declaraciones de namespace
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "idPkg":
			// Manejar declaración de namespace xmlns:idPkg
			if attr.Name.Space == "xmlns" {
				d.Xmlns = attr.Value
			}
		case "DOMVersion":
			d.DOMVersion = attr.Value
		case "Self":
			d.Self = attr.Value
		case "Name":
			d.Name = attr.Value
		case "StoryList":
			d.StoryList = attr.Value
		case "ZeroPoint":
			d.ZeroPoint = attr.Value
		case "ActiveLayer":
			d.ActiveLayer = attr.Value
		case "CMYKProfile":
			d.CMYKProfile = attr.Value
		case "RGBProfile":
			d.RGBProfile = attr.Value
		case "SolidColorIntent":
			d.SolidColorIntent = attr.Value
		case "AfterBlendingIntent":
			d.AfterBlendingIntent = attr.Value
		case "DefaultImageIntent":
			d.DefaultImageIntent = attr.Value
		case "RGBPolicy":
			d.RGBPolicy = attr.Value
		case "CMYKPolicy":
			d.CMYKPolicy = attr.Value
		case "AccurateLABSpots":
			d.AccurateLABSpots = attr.Value
		case "AppliedMathMLFontSize":
			d.AppliedMathMLFontSize = attr.Value
		case "AppliedMathMLRgbColor":
			d.AppliedMathMLRgbColor = attr.Value
		case "PreferMathMLInEpubExport":
			d.PreferMathMLInEpubExport = attr.Value
		case "TintValue":
			d.TintValue = attr.Value
		}
	}

	// Parsear elementos hijo
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return common.WrapError("document", "parse document", err)
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if err := d.unmarshalChildElement(decoder, elem); err != nil {
				return err
			}
		case xml.EndElement:
			// Llegamos al final del elemento Document
			return nil
		}
	}

	return nil
}

// unmarshalChildElement maneja la deserialización de elementos hijo según su nombre y namespace.
func (d *Document) unmarshalChildElement(decoder *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if decoder == nil {
		return common.Errorf("document", "unmarshal child element", "", "decoder is nil")
	}

	// Manejar elementos del namespace idPkg (referencias a recursos)
	if start.Name.Space == "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" {
		return d.unmarshalResourceRef(decoder, start)
	}

	// Manejar elementos hijo regulares
	switch start.Name.Local {
	case "Properties":
		var props common.Properties
		if err := decoder.DecodeElement(&props, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Properties = &props

	case "Language":
		var lang Language
		if err := decoder.DecodeElement(&lang, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Languages = append(d.Languages, lang)

	case "Layer":
		var layer Layer
		if err := decoder.DecodeElement(&layer, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Layers = append(d.Layers, layer)

	case "NumberingList":
		var nl NumberingList
		if err := decoder.DecodeElement(&nl, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.NumberingLists = append(d.NumberingLists, nl)

	case "NamedGrid":
		var ng NamedGrid
		if err := decoder.DecodeElement(&ng, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.NamedGrids = append(d.NamedGrids, ng)

	case "Section":
		var section Section
		if err := decoder.DecodeElement(&section, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Sections = append(d.Sections, section)

	case "DocumentUser":
		var user DocumentUser
		if err := decoder.DecodeElement(&user, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.DocumentUsers = append(d.DocumentUsers, user)

	case "ColorGroup":
		var cg ColorGroup
		if err := decoder.DecodeElement(&cg, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.ColorGroups = append(d.ColorGroups, cg)

	case "ABullet":
		var bullet ABullet
		if err := decoder.DecodeElement(&bullet, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.ABullets = append(d.ABullets, bullet)

	case "Assignment":
		var assignment Assignment
		if err := decoder.DecodeElement(&assignment, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Assignments = append(d.Assignments, assignment)

	case "TextVariable":
		var tv TextVariable
		if err := decoder.DecodeElement(&tv, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.TextVariables = append(d.TextVariables, tv)

	// Recursos inline para IDMS (usados en snippets en lugar de archivos separados)
	case "Color":
		var color resources.Color
		if err := decoder.DecodeElement(&color, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Colors = append(d.Colors, color)

	case "Swatch":
		var swatch resources.Swatch
		if err := decoder.DecodeElement(&swatch, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Swatches = append(d.Swatches, swatch)

	case "StrokeStyle":
		var strokeStyle resources.StrokeStyle
		if err := decoder.DecodeElement(&strokeStyle, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.StrokeStyles = append(d.StrokeStyles, strokeStyle)

	case "RootCharacterStyleGroup":
		var group resources.CharacterStyleGroup
		if err := decoder.DecodeElement(&group, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.RootCharacterStyleGroup = &group

	case "RootParagraphStyleGroup":
		var group resources.ParagraphStyleGroup
		if err := decoder.DecodeElement(&group, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.RootParagraphStyleGroup = &group

	case "RootObjectStyleGroup":
		var group resources.ObjectStyleGroup
		if err := decoder.DecodeElement(&group, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.RootObjectStyleGroup = &group

	case "TinDocumentDataObject":
		var tin TinDocumentDataObject
		if err := decoder.DecodeElement(&tin, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.TinDocumentDataObject = &tin

	case "TransparencyDefaultContainerObject":
		var trans TransparencyDefaultContainerObject
		if err := decoder.DecodeElement(&trans, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.TransparencyDefaultContainerObject = &trans

	// Contenido inline para IDMS (spreads y stories)
	case "Spread":
		var spreadElem spread.SpreadElement
		if err := decoder.DecodeElement(&spreadElem, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.InlineSpreads = append(d.InlineSpreads, spreadElem)

	case "Story":
		var storyElem story.StoryElement
		if err := decoder.DecodeElement(&storyElem, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.InlineStories = append(d.InlineStories, storyElem)

	default:
		// Elemento desconocido - preservar como RawXMLElement
		var raw common.RawXMLElement
		if err := decoder.DecodeElement(&raw, &start); err != nil {
			return common.WrapErrorWithPath("document", "parse", start.Name.Local, err)
		}
		d.OtherElements = append(d.OtherElements, raw)
	}

	return nil
}

// unmarshalResourceRef maneja la deserialización de referencias a recursos del namespace idPkg.
func (d *Document) unmarshalResourceRef(decoder *xml.Decoder, start xml.StartElement) error {
	// Verificar que el decoder no sea nil
	if decoder == nil {
		return common.Errorf("document", "unmarshal resource ref", "", "decoder is nil")
	}

	var ref ResourceRef
	if err := decoder.DecodeElement(&ref, &start); err != nil {
		return common.WrapError("document", "parse document", err)
	}

	// Asignar al campo correspondiente según el nombre del elemento
	switch start.Name.Local {
	case "Graphic":
		d.GraphicResource = &ref
	case "Fonts":
		d.FontsResource = &ref
	case "Styles":
		d.StylesResource = &ref
	case "Preferences":
		d.PreferencesResource = &ref
	case "Tags":
		d.TagsResource = &ref
	case "MasterSpread":
		d.MasterSpreads = append(d.MasterSpreads, ref)
	case "Spread":
		d.Spreads = append(d.Spreads, ref)
	case "Story":
		d.Stories = append(d.Stories, ref)
	case "BackingStory":
		d.BackingStory = &ref
	default:
		// Referencia a recurso desconocida - se podría agregar a OtherElements si fuera necesario
		return common.Errorf("document", "parse document", "", "unknown resource reference type: %s", start.Name.Local)
	}

	return nil
}

// MarshalXML implementa serialización XML custom para Document.
// Nos da control total sobre el orden de atributos y las declaraciones de namespace.
func (d Document) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	// Crear el elemento de inicio del Document
	start.Name = xml.Name{Local: "Document"}

	// Construir atributos en un orden específico para consistencia
	var attrs []xml.Attr

	// Declaración de namespace primero (si está presente)
	if d.Xmlns != "" {
		attrs = append(attrs, xml.Attr{
			Name:  xml.Name{Space: "xmlns", Local: "idPkg"},
			Value: d.Xmlns,
		})
	}

	// Atributos principales
	if d.DOMVersion != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "DOMVersion"}, Value: d.DOMVersion})
	}
	if d.Self != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "Self"}, Value: d.Self})
	}
	if d.Name != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "Name"}, Value: d.Name})
	}

	// Gestión de stories
	if d.StoryList != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "StoryList"}, Value: d.StoryList})
	}

	// Atributos de layout
	if d.ZeroPoint != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "ZeroPoint"}, Value: d.ZeroPoint})
	}
	if d.ActiveLayer != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "ActiveLayer"}, Value: d.ActiveLayer})
	}

	// Atributos de gestión de color
	if d.CMYKProfile != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "CMYKProfile"}, Value: d.CMYKProfile})
	}
	if d.RGBProfile != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "RGBProfile"}, Value: d.RGBProfile})
	}
	if d.SolidColorIntent != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "SolidColorIntent"}, Value: d.SolidColorIntent})
	}
	if d.AfterBlendingIntent != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "AfterBlendingIntent"}, Value: d.AfterBlendingIntent})
	}
	if d.DefaultImageIntent != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "DefaultImageIntent"}, Value: d.DefaultImageIntent})
	}
	if d.RGBPolicy != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "RGBPolicy"}, Value: d.RGBPolicy})
	}
	if d.CMYKPolicy != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "CMYKPolicy"}, Value: d.CMYKPolicy})
	}
	if d.AccurateLABSpots != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "AccurateLABSpots"}, Value: d.AccurateLABSpots})
	}

	// Atributos MathML
	if d.AppliedMathMLFontSize != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "AppliedMathMLFontSize"}, Value: d.AppliedMathMLFontSize})
	}
	if d.AppliedMathMLRgbColor != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "AppliedMathMLRgbColor"}, Value: d.AppliedMathMLRgbColor})
	}
	if d.PreferMathMLInEpubExport != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "PreferMathMLInEpubExport"}, Value: d.PreferMathMLInEpubExport})
	}
	if d.TintValue != "" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "TintValue"}, Value: d.TintValue})
	}

	start.Attr = attrs

	// Escribir el elemento de inicio
	if err := encoder.EncodeToken(start); err != nil {
		return common.WrapError("document", "marshal document", err)
	}

	// Serializar elementos hijo en orden
	if err := d.marshalChildren(encoder); err != nil {
		return err
	}

	// Escribir el elemento de cierre
	if err := encoder.EncodeToken(start.End()); err != nil {
		return common.WrapError("document", "marshal document", err)
	}

	return encoder.Flush()
}

// marshalChildren serializa todos los elementos hijo en el orden correcto.
func (d Document) marshalChildren(encoder *xml.Encoder) error {
	// 1. Properties
	if d.Properties != nil {
		if err := encoder.Encode(d.Properties); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 2. Idiomas
	for _, lang := range d.Languages {
		if err := encoder.Encode(lang); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 3. Referencias a recursos (namespace idPkg)
	if err := d.marshalResourceRefs(encoder); err != nil {
		return err
	}

	// 4. Capas
	for _, layer := range d.Layers {
		if err := encoder.Encode(layer); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 5. Listas de numeración
	for _, nl := range d.NumberingLists {
		if err := encoder.Encode(nl); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 6. Grillas con nombre
	for _, ng := range d.NamedGrids {
		if err := encoder.Encode(ng); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 7. Secciones
	for _, section := range d.Sections {
		if err := encoder.Encode(section); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 8. Usuarios del documento
	for _, user := range d.DocumentUsers {
		if err := encoder.Encode(user); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 9. Grupos de colores
	for _, cg := range d.ColorGroups {
		if err := encoder.Encode(cg); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 10. Viñetas
	for _, bullet := range d.ABullets {
		if err := encoder.Encode(bullet); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 11. Asignaciones
	for _, assignment := range d.Assignments {
		if err := encoder.Encode(assignment); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 12. Variables de texto
	for _, tv := range d.TextVariables {
		if err := encoder.Encode(tv); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 13. Contenido inline para IDMS (colores, muestras, estilos, spreads, stories)
	// Se usan en archivos IDMS (snippets) en lugar de referencias a recursos
	for _, color := range d.Colors {
		if err := encoder.Encode(color); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	for _, swatch := range d.Swatches {
		if err := encoder.Encode(swatch); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	for _, strokeStyle := range d.StrokeStyles {
		if err := encoder.Encode(strokeStyle); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// Serializar grupos de estilos raíz (los tags XML están definidos en los struct tags)
	if d.RootCharacterStyleGroup != nil {
		if err := encoder.Encode(d.RootCharacterStyleGroup); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.RootParagraphStyleGroup != nil {
		if err := encoder.Encode(d.RootParagraphStyleGroup); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.RootObjectStyleGroup != nil {
		if err := encoder.Encode(d.RootObjectStyleGroup); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	if d.TinDocumentDataObject != nil {
		if err := encoder.Encode(d.TinDocumentDataObject); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.TransparencyDefaultContainerObject != nil {
		if err := encoder.Encode(d.TransparencyDefaultContainerObject); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	for _, spread := range d.InlineSpreads {
		if err := encoder.Encode(spread); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	for _, story := range d.InlineStories {
		if err := encoder.Encode(story); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// 14. Elementos desconocidos restantes
	for _, elem := range d.OtherElements {
		if err := encoder.Encode(elem); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	return nil
}

// marshalResourceRefs serializa todos los elementos de referencia a recursos.
func (d Document) marshalResourceRefs(encoder *xml.Encoder) error {
	// Referencias a recursos individuales
	if d.GraphicResource != nil {
		if err := encoder.Encode(d.GraphicResource); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.FontsResource != nil {
		if err := encoder.Encode(d.FontsResource); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.StylesResource != nil {
		if err := encoder.Encode(d.StylesResource); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.PreferencesResource != nil {
		if err := encoder.Encode(d.PreferencesResource); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.TagsResource != nil {
		if err := encoder.Encode(d.TagsResource); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	// Referencias a recursos múltiples
	for _, ms := range d.MasterSpreads {
		if err := encoder.Encode(ms); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	for _, spread := range d.Spreads {
		if err := encoder.Encode(spread); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	for _, story := range d.Stories {
		if err := encoder.Encode(story); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}
	if d.BackingStory != nil {
		if err := encoder.Encode(d.BackingStory); err != nil {
			return common.WrapError("document", "marshal document", err)
		}
	}

	return nil
}
