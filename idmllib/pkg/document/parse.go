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

	// Manejar elementos hijo regulares.
	//
	// Cada rama registra su clase en d.childOrder justo donde guarda el hijo en su
	// campo. Tenerlo pegado al append es lo que evita que las dos cosas se
	// desincronicen al agregar un elemento nuevo al modelo.
	switch start.Name.Local {
	case "Properties":
		var props common.Properties
		if err := decoder.DecodeElement(&props, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Properties = &props
		d.childOrder.Record(childProperties)

	case "Language":
		var lang Language
		if err := decoder.DecodeElement(&lang, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Languages = append(d.Languages, lang)
		d.childOrder.Record(childLanguage)

	case "Layer":
		var layer Layer
		if err := decoder.DecodeElement(&layer, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Layers = append(d.Layers, layer)
		d.childOrder.Record(childLayer)

	case "NumberingList":
		var nl NumberingList
		if err := decoder.DecodeElement(&nl, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.NumberingLists = append(d.NumberingLists, nl)
		d.childOrder.Record(childNumberingList)

	case "NamedGrid":
		var ng NamedGrid
		if err := decoder.DecodeElement(&ng, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.NamedGrids = append(d.NamedGrids, ng)
		d.childOrder.Record(childNamedGrid)

	case "Section":
		var section Section
		if err := decoder.DecodeElement(&section, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Sections = append(d.Sections, section)
		d.childOrder.Record(childSection)

	case "DocumentUser":
		var user DocumentUser
		if err := decoder.DecodeElement(&user, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.DocumentUsers = append(d.DocumentUsers, user)
		d.childOrder.Record(childDocumentUser)

	case "ColorGroup":
		var cg ColorGroup
		if err := decoder.DecodeElement(&cg, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.ColorGroups = append(d.ColorGroups, cg)
		d.childOrder.Record(childColorGroup)

	case "ABullet":
		var bullet ABullet
		if err := decoder.DecodeElement(&bullet, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.ABullets = append(d.ABullets, bullet)
		d.childOrder.Record(childABullet)

	case "Assignment":
		var assignment Assignment
		if err := decoder.DecodeElement(&assignment, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Assignments = append(d.Assignments, assignment)
		d.childOrder.Record(childAssignment)

	case "TextVariable":
		var tv TextVariable
		if err := decoder.DecodeElement(&tv, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.TextVariables = append(d.TextVariables, tv)
		d.childOrder.Record(childTextVariable)

	// Recursos inline para IDMS (usados en snippets en lugar de archivos separados)
	case "Color":
		var color resources.Color
		if err := decoder.DecodeElement(&color, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Colors = append(d.Colors, color)
		d.childOrder.Record(childColor)

	case "Swatch":
		var swatch resources.Swatch
		if err := decoder.DecodeElement(&swatch, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.Swatches = append(d.Swatches, swatch)
		d.childOrder.Record(childSwatch)

	case "StrokeStyle":
		var strokeStyle resources.StrokeStyle
		if err := decoder.DecodeElement(&strokeStyle, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.StrokeStyles = append(d.StrokeStyles, strokeStyle)
		d.childOrder.Record(childStrokeStyle)

	case "RootCharacterStyleGroup":
		var group resources.CharacterStyleGroup
		if err := decoder.DecodeElement(&group, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.RootCharacterStyleGroup = &group
		d.childOrder.Record(childRootCharStyleGroup)

	case "RootParagraphStyleGroup":
		var group resources.ParagraphStyleGroup
		if err := decoder.DecodeElement(&group, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.RootParagraphStyleGroup = &group
		d.childOrder.Record(childRootParaStyleGroup)

	case "RootObjectStyleGroup":
		var group resources.ObjectStyleGroup
		if err := decoder.DecodeElement(&group, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.RootObjectStyleGroup = &group
		d.childOrder.Record(childRootObjStyleGroup)

	case "TinDocumentDataObject":
		var tin TinDocumentDataObject
		if err := decoder.DecodeElement(&tin, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.TinDocumentDataObject = &tin
		d.childOrder.Record(childTinDocumentData)

	case "TransparencyDefaultContainerObject":
		var trans TransparencyDefaultContainerObject
		if err := decoder.DecodeElement(&trans, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.TransparencyDefaultContainerObject = &trans
		d.childOrder.Record(childTransparencyDefault)

	// Contenido inline para IDMS (spreads y stories)
	case "Spread":
		var spreadElem spread.SpreadElement
		if err := decoder.DecodeElement(&spreadElem, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.InlineSpreads = append(d.InlineSpreads, spreadElem)
		d.childOrder.Record(childInlineSpread)

	case "Story":
		var storyElem story.StoryElement
		if err := decoder.DecodeElement(&storyElem, &start); err != nil {
			return common.WrapError("document", "parse document", err)
		}
		d.InlineStories = append(d.InlineStories, storyElem)
		d.childOrder.Record(childInlineStory)

	default:
		// Elemento desconocido - preservar como RawXMLElement
		var raw common.RawXMLElement
		if err := decoder.DecodeElement(&raw, &start); err != nil {
			return common.WrapErrorWithPath("document", "parse", start.Name.Local, err)
		}
		d.OtherElements = append(d.OtherElements, raw)
		d.childOrder.Record(childOther)
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
		d.childOrder.Record(childRefGraphic)
	case "Fonts":
		d.FontsResource = &ref
		d.childOrder.Record(childRefFonts)
	case "Styles":
		d.StylesResource = &ref
		d.childOrder.Record(childRefStyles)
	case "Preferences":
		d.PreferencesResource = &ref
		d.childOrder.Record(childRefPreferences)
	case "Tags":
		d.TagsResource = &ref
		d.childOrder.Record(childRefTags)
	case "MasterSpread":
		d.MasterSpreads = append(d.MasterSpreads, ref)
		d.childOrder.Record(childRefMasterSpread)
	case "Spread":
		d.Spreads = append(d.Spreads, ref)
		d.childOrder.Record(childRefSpread)
	case "Story":
		d.Stories = append(d.Stories, ref)
		d.childOrder.Record(childRefStory)
	case "BackingStory":
		d.BackingStory = &ref
		d.childOrder.Record(childRefBackingStory)
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

// marshalChildren serializa los elementos hijo en el orden en que venían al
// parsear, o en el orden de los campos del struct cuando no hay orden registrado.
func (d Document) marshalChildren(encoder *xml.Encoder) error {
	return d.childOrder.Replay(documentChildOrder, d.childrenByKind(encoder))
}

// childrenByKind agrupa los hijos que el documento tiene ahora, por clase y en el
// orden de su campo, cada uno envuelto en la función que lo emite.
//
// Los emisores toman la dirección del elemento dentro de su campo y no una copia,
// así que un hijo mutado a través de su campo se serializa con el valor nuevo. Se
// construye una vez por serialización, con los campos ya en su estado final.
func (d Document) childrenByKind(encoder *xml.Encoder) map[string][]xmlutil.ChildEmitter {
	children := make(map[string][]xmlutil.ChildEmitter, len(documentChildOrder))

	encode := func(child any) xmlutil.ChildEmitter {
		return func() error {
			if err := encoder.Encode(child); err != nil {
				return common.WrapError("document", "marshal document", err)
			}
			return nil
		}
	}
	one := func(kind string, child any, present bool) {
		if present {
			children[kind] = []xmlutil.ChildEmitter{encode(child)}
		}
	}

	one(childProperties, d.Properties, d.Properties != nil)
	children[childLanguage] = emitters(d.Languages, encode)

	// Referencias a recursos del namespace idPkg
	one(childRefGraphic, d.GraphicResource, d.GraphicResource != nil)
	one(childRefFonts, d.FontsResource, d.FontsResource != nil)
	one(childRefStyles, d.StylesResource, d.StylesResource != nil)
	one(childRefPreferences, d.PreferencesResource, d.PreferencesResource != nil)
	one(childRefTags, d.TagsResource, d.TagsResource != nil)
	children[childRefMasterSpread] = emitters(d.MasterSpreads, encode)
	children[childRefSpread] = emitters(d.Spreads, encode)
	children[childRefStory] = emitters(d.Stories, encode)
	one(childRefBackingStory, d.BackingStory, d.BackingStory != nil)

	children[childLayer] = emitters(d.Layers, encode)
	children[childNumberingList] = emitters(d.NumberingLists, encode)
	children[childNamedGrid] = emitters(d.NamedGrids, encode)
	children[childSection] = emitters(d.Sections, encode)
	children[childDocumentUser] = emitters(d.DocumentUsers, encode)
	children[childColorGroup] = emitters(d.ColorGroups, encode)
	children[childABullet] = emitters(d.ABullets, encode)
	children[childAssignment] = emitters(d.Assignments, encode)
	children[childTextVariable] = emitters(d.TextVariables, encode)

	// Contenido inline de un IDMS, que en IDML vive en archivos aparte
	children[childColor] = emitters(d.Colors, encode)
	children[childSwatch] = emitters(d.Swatches, encode)
	children[childStrokeStyle] = emitters(d.StrokeStyles, encode)
	one(childRootCharStyleGroup, d.RootCharacterStyleGroup, d.RootCharacterStyleGroup != nil)
	one(childRootParaStyleGroup, d.RootParagraphStyleGroup, d.RootParagraphStyleGroup != nil)
	one(childRootObjStyleGroup, d.RootObjectStyleGroup, d.RootObjectStyleGroup != nil)
	one(childTinDocumentData, d.TinDocumentDataObject, d.TinDocumentDataObject != nil)
	one(childTransparencyDefault, d.TransparencyDefaultContainerObject, d.TransparencyDefaultContainerObject != nil)
	children[childInlineSpread] = emitters(d.InlineSpreads, encode)
	children[childInlineStory] = emitters(d.InlineStories, encode)

	children[childOther] = emitters(d.OtherElements, encode)

	return children
}

// emitters construye un emisor por elemento del slice, tomando la dirección de
// cada uno para no serializar copias.
func emitters[T any](items []T, encode func(any) xmlutil.ChildEmitter) []xmlutil.ChildEmitter {
	if len(items) == 0 {
		return nil
	}
	out := make([]xmlutil.ChildEmitter, len(items))
	for i := range items {
		out[i] = encode(&items[i])
	}
	return out
}
