// cmd/idmlgen genera un archivo IDML a partir de una descripción de documento JSON
// recibida por stdin. Escribe el IDML en stdout (por defecto) o en la ruta indicada
// por -out.
//
// Este binario es agnóstico al dominio: no sabe de pautas, grillas editoriales, ni
// folios. Solo recibe páginas con marcos posicionados en milímetros.
//
// Códigos de salida:
//   - 0: generación exitosa
//   - 1: error de entrada (JSON inválido, validación fallida)
//   - 2: error de generación (fallo interno de idmllib)
//
// Ejemplo:
//
//	cat documento.json | idmlgen > edicion.idml
//	cat documento.json | idmlgen -out edicion.idml
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/idml/idgen"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"

	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
)

// --- Modelo de entrada (genérico, agnóstico al dominio) ---

type documentInput struct {
	Document documentSpec `json:"document"`
	Pages    []pageSpec   `json:"pages"`
}

type documentSpec struct {
	WidthMm     float64     `json:"widthMm"`
	HeightMm    float64     `json:"heightMm"`
	Margins     marginsSpec `json:"margins"`
	FacingPages bool        `json:"facingPages"`
	Columns     int         `json:"columns"`
	Guides      []guideSpec `json:"guides"`
}

type marginsSpec struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

type guideSpec struct {
	Orientation string  `json:"orientation"` // "vertical" o "horizontal"
	LocationMm  float64 `json:"locationMm"`  // posición en mm desde el borde de la página
}

type pageSpec struct {
	Frames []frameSpec `json:"frames"`
}

type frameSpec struct {
	Type    string       `json:"type"`
	Name    string       `json:"name"`
	Bounds  boundsSpec   `json:"bounds"`
	Content string       `json:"content"`
	Options frameOptions `json:"options"`
}

type boundsSpec struct {
	TopMm    float64 `json:"topMm"`
	LeftMm   float64 `json:"leftMm"`
	BottomMm float64 `json:"bottomMm"`
	RightMm  float64 `json:"rightMm"`
}

type frameOptions struct {
	VerticalJustification string `json:"verticalJustification"`
	ContentIsRaw          bool   `json:"contentIsRaw"`
}

// --- Conversión de unidades ---

const ptPerMm = 72.0 / 25.4

var selfAttrRegex = regexp.MustCompile(`Self="([^"]+)"`)

func mmToPt(mm float64) float64 { return mm * ptPerMm }

func num(v float64) string {
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// --- Validación ---

func validate(input *documentInput) error {
	doc := input.Document
	if doc.WidthMm <= 0 {
		return fmt.Errorf("document.widthMm: debe ser > 0, se recibió %v", doc.WidthMm)
	}
	if doc.HeightMm <= 0 {
		return fmt.Errorf("document.heightMm: debe ser > 0, se recibió %v", doc.HeightMm)
	}
	if len(input.Pages) == 0 {
		return fmt.Errorf("pages: debe tener al menos una página")
	}
	for i, page := range input.Pages {
		for j, frame := range page.Frames {
			if frame.Type == "" {
				return fmt.Errorf("pages[%d].frames[%d].type: es obligatorio", i, j)
			}
			if frame.Type != "text" {
				return fmt.Errorf("pages[%d].frames[%d].type: solo se soporta \"text\", se recibió %q", i, j, frame.Type)
			}
			b := frame.Bounds
			if b.BottomMm <= b.TopMm {
				return fmt.Errorf("pages[%d].frames[%d].bounds: bottomMm (%v) debe ser > topMm (%v)", i, j, b.BottomMm, b.TopMm)
			}
			if b.RightMm <= b.LeftMm {
				return fmt.Errorf("pages[%d].frames[%d].bounds: rightMm (%v) debe ser > leftMm (%v)", i, j, b.RightMm, b.LeftMm)
			}
		}
	}
	return nil
}

// --- Generación ---

func generate(input *documentInput) (*idmlpkg.Package, error) {
	doc := input.Document

	pageWidthPt := mmToPt(doc.WidthMm)
	pageHeightPt := mmToPt(doc.HeightMm)

	columns := doc.Columns
	if columns <= 0 {
		columns = 1
	}

	opts := &idmlpkg.TemplateOptions{
		DOMVersion:  "20.4",
		Preset:      idmlpkg.PresetCustom,
		Orientation: "Portrait",
		ColumnCount: columns,
		CustomDimensions: &idmlpkg.PageDimensions{
			Width:  pageWidthPt,
			Height: pageHeightPt,
		},
	}
	opts.Margins.Top = mmToPt(doc.Margins.Top)
	opts.Margins.Bottom = mmToPt(doc.Margins.Bottom)
	opts.Margins.Left = mmToPt(doc.Margins.Left)
	opts.Margins.Right = mmToPt(doc.Margins.Right)

	pkg, err := idmlpkg.NewFromTemplate(opts)
	if err != nil {
		return nil, fmt.Errorf("error al crear documento base: %w", err)
	}

	// Eliminar el TextFrame de la plantilla (uf3) y su story huérfana (ue1).
	_, _ = pkg.RemoveTextFrame(idmlpkg.PathSpread, "uf3", false)
	_, _ = pkg.RemoveStory("Stories/Story_ue1.xml", false)

	designmapDoc, err := pkg.Document()
	if err != nil {
		return nil, fmt.Errorf("error al leer designmap: %w", err)
	}
	designmapDoc.StoryList = strings.Replace(designmapDoc.StoryList, "ue1 ", "", 1)
	designmapDoc.StoryList = strings.Replace(designmapDoc.StoryList, " ue1", "", 1)
	designmapDoc.StoryList = strings.Replace(designmapDoc.StoryList, "ue1", "", 1)
	cleaned := make([]document.ResourceRef, 0, len(designmapDoc.Stories))
	for _, s := range designmapDoc.Stories {
		if s.Src != "Stories/Story_ue1.xml" {
			cleaned = append(cleaned, s)
		}
	}
	designmapDoc.Stories = cleaned

	// Registrar todos los IDs que la plantilla ya contiene para evitar colisiones.
	// Se escanean dinámicamente del paquete generado — agnóstico a qué plantilla se usó.
	reg := idgen.New()
	for _, file := range pkg.Files() {
		data, err := pkg.GetFileData(file)
		if err != nil {
			continue
		}
		for _, match := range selfAttrRegex.FindAllSubmatch(data, -1) {
			if len(match) > 1 {
				_ = reg.Register(string(match[1]))
			}
		}
	}

	halfHeight := pageHeightPt / 2

	// --- Agrupar páginas en spreads ---
	// Si facingPages: página 1 sola, luego pares (2-3, 4-5...), posible última sola.
	// Si no: cada página es su propio spread.
	type spreadGroup struct {
		pages       []pageSpec
		pageNumbers []int
	}

	var spreads []spreadGroup
	if doc.FacingPages && len(input.Pages) > 1 {
		// Portada sola
		spreads = append(spreads, spreadGroup{
			pages:       []pageSpec{input.Pages[0]},
			pageNumbers: []int{1},
		})
		// Pares
		for i := 1; i < len(input.Pages); i += 2 {
			if i+1 < len(input.Pages) {
				spreads = append(spreads, spreadGroup{
					pages:       []pageSpec{input.Pages[i], input.Pages[i+1]},
					pageNumbers: []int{i + 1, i + 2},
				})
			} else {
				// Contraportada sola
				spreads = append(spreads, spreadGroup{
					pages:       []pageSpec{input.Pages[i]},
					pageNumbers: []int{i + 1},
				})
			}
		}
	} else {
		for i, p := range input.Pages {
			spreads = append(spreads, spreadGroup{
				pages:       []pageSpec{p},
				pageNumbers: []int{i + 1},
			})
		}
	}

	// --- Generar cada spread ---
	for i, sg := range spreads {
		if i == 0 {
			// Primer spread: usa el de la plantilla.
			for _, frame := range sg.pages[0].Frames {
				if err := addTextFrame(pkg, reg, frame, halfHeight); err != nil {
					return nil, err
				}
			}
			if len(doc.Guides) > 0 {
				if err := addGuides(pkg, reg, doc.Guides, idmlpkg.PathSpread); err != nil {
					return nil, err
				}
			}
		} else {
			spreadPath, err := addSpreadForPages(pkg, reg, doc, sg.pages, sg.pageNumbers, halfHeight)
			if err != nil {
				return nil, err
			}
			if len(doc.Guides) > 0 {
				if err := addGuides(pkg, reg, doc.Guides, spreadPath); err != nil {
					return nil, err
				}
			}
		}
	}

	return pkg, nil
}

// addSpreadForPages crea un spread con 1 o 2 páginas y lo registra en el paquete.
func addSpreadForPages(pkg *idmlpkg.Package, reg *idgen.Registry, doc documentSpec, pages []pageSpec, pageNumbers []int, halfHeight float64) (string, error) {
	spreadID := reg.Generate()
	pageWidthPt := mmToPt(doc.WidthMm)
	pageHeightPt := mmToPt(doc.HeightMm)
	pageCount := len(pages)

	columns := doc.Columns
	if columns <= 0 {
		columns = 1
	}
	marginLeft := mmToPt(doc.Margins.Left)
	marginRight := mmToPt(doc.Margins.Right)
	usableWidth := pageWidthPt - marginLeft - marginRight
	columnGutter := 12.0
	columnWidth := (usableWidth - float64(columns-1)*columnGutter) / float64(columns)

	var positions []string
	for i := 0; i < columns; i++ {
		start := float64(i) * (columnWidth + columnGutter)
		positions = append(positions, num(start), num(start+columnWidth))
	}
	columnsPositions := strings.Join(positions, " ")

	var spreadXML strings.Builder
	fmt.Fprintf(&spreadXML, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Spread Self="%s" PageTransitionType="None" PageTransitionDirection="NotApplicable" PageTransitionDuration="Medium" ShowMasterItems="true" PageCount="%d" BindingLocation="0" SpreadHidden="false" AllowPageShuffle="true" ItemTransform="1 0 0 1 0 0" FlattenerOverride="Default">
		<FlattenerPreference LineArtAndTextResolution="300" GradientAndMeshResolution="150" ClipComplexRegions="false" ConvertAllStrokesToOutlines="false" ConvertAllTextToOutlines="false">
			<Properties>
				<RasterVectorBalance type="double">50</RasterVectorBalance>
			</Properties>
		</FlattenerPreference>
`, spreadID, pageCount)

	// Emitir cada página del spread.
	for pi, page := range pages {
		pageID := reg.Generate()
		pageNum := pageNumbers[pi]

		// Calcular ItemTransform de la página.
		// Facing pages con 2 páginas: izquierda X=-pageWidth, derecha X=0.
		// Página sola: X=0.
		var pageTransformX float64
		if pageCount == 2 {
			if pi == 0 {
				pageTransformX = -pageWidthPt // izquierda
			} else {
				pageTransformX = 0 // derecha
			}
		} else {
			pageTransformX = 0
		}

		fmt.Fprintf(&spreadXML, `		<Page Self="%s" TabOrder="" AppliedMaster="ub4" OverrideList="" MasterPageTransform="1 0 0 1 0 0" Name="%d" AppliedTrapPreset="TrapPreset/$ID/kDefaultTrapStyleName" GeometricBounds="0 0 %s %s" ItemTransform="1 0 0 1 %s %s" AppliedAlternateLayout="ub6" LayoutRule="UseMaster" SnapshotBlendingMode="IgnoreLayoutSnapshots" OptionalPage="false" GridStartingPoint="TopOutside" UseMasterGrid="true">
			<Properties>
				<PageColor type="enumeration">UseMasterColor</PageColor>
			</Properties>
			<MarginPreference ColumnCount="%d" ColumnGutter="%s" Top="%s" Bottom="%s" Left="%s" Right="%s" ColumnDirection="Horizontal" ColumnsPositions="%s" />
		</Page>
`,
			pageID, pageNum,
			num(pageHeightPt), num(pageWidthPt),
			num(pageTransformX), num(-halfHeight),
			columns, num(columnGutter),
			num(mmToPt(doc.Margins.Top)), num(mmToPt(doc.Margins.Bottom)),
			num(marginLeft), num(marginRight),
			columnsPositions)

		// TextFrames de esta página.
		// El desplazamiento X de los frames debe sumar pageTransformX para estar en la página correcta.
		for _, frame := range page.Frames {
			frameID := reg.Generate()
			storyID := reg.Generate()

			topPt := mmToPt(frame.Bounds.TopMm)
			leftPt := mmToPt(frame.Bounds.LeftMm)
			bottomPt := mmToPt(frame.Bounds.BottomMm)
			rightPt := mmToPt(frame.Bounds.RightMm)

			w := rightPt - leftPt
			h := bottomPt - topPt
			centerX := leftPt + w/2 + pageTransformX
			centerY := topPt + h/2 - halfHeight

			vJust := ""
			if frame.Options.VerticalJustification != "" {
				vJust = fmt.Sprintf(` VerticalJustification="%s"`, frame.Options.VerticalJustification)
			}

			fmt.Fprintf(&spreadXML, `		<TextFrame Self="%s" Name="%s" ItemLayer="uba" Visible="true" ItemTransform="1 0 0 1 %s %s" ParentStory="%s" PreviousTextFrame="n" NextTextFrame="n" ContentType="TextType" OverriddenPageItemProps="" AppliedObjectStyle="ObjectStyle/$ID/[Normal Text Frame]" ParentInterfaceChangeCount="" TargetInterfaceChangeCount="" LastUpdatedInterfaceChangeCount="">
			<Properties>
				<PathGeometry>
					<GeometryPathType PathOpen="false">
						<PathPointArray>
							<PathPointType Anchor="%s %s" LeftDirection="%s %s" RightDirection="%s %s" />
							<PathPointType Anchor="%s %s" LeftDirection="%s %s" RightDirection="%s %s" />
							<PathPointType Anchor="%s %s" LeftDirection="%s %s" RightDirection="%s %s" />
							<PathPointType Anchor="%s %s" LeftDirection="%s %s" RightDirection="%s %s" />
						</PathPointArray>
					</GeometryPathType>
				</PathGeometry>
			</Properties>
			<TextFramePreference TextColumnCount="1" TextColumnFixedWidth="0" TextColumnMaxWidth="0"%s><Properties><InsetSpacing type="list"><ListItem type="unit">0</ListItem><ListItem type="unit">0</ListItem><ListItem type="unit">0</ListItem><ListItem type="unit">0</ListItem></InsetSpacing></Properties></TextFramePreference>
			<TextWrapPreference Inverse="false" ApplyToMasterPageOnly="false" TextWrapSide="BothSides" TextWrapMode="None"><Properties><TextWrapOffset Top="0" Left="0" Bottom="0" Right="0" /></Properties></TextWrapPreference>
		</TextFrame>
`,
				frameID, frame.Name, num(centerX), num(centerY), storyID,
				num(-w/2), num(-h/2), num(-w/2), num(-h/2), num(-w/2), num(-h/2),
				num(-w/2), num(h/2), num(-w/2), num(h/2), num(-w/2), num(h/2),
				num(w/2), num(h/2), num(w/2), num(h/2), num(w/2), num(h/2),
				num(w/2), num(-h/2), num(w/2), num(-h/2), num(w/2), num(-h/2),
				vJust)

			// Story
			var children []story.CharacterChild
			if frame.Content != "" {
				if frame.Options.ContentIsRaw {
					children = []story.CharacterChild{{Content: &story.Content{Raw: frame.Content}}}
				} else {
					children = []story.CharacterChild{{Content: &story.Content{Text: frame.Content}}}
				}
			}

			st := &story.Story{
				DOMVersion: "20.4",
				StoryElement: story.StoryElement{
					Self:     storyID,
					UserText: "true",
					ParagraphStyleRanges: []story.ParagraphStyleRange{{
						AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
						CharacterStyleRanges: []story.CharacterStyleRange{{
							AppliedCharacterStyle: "CharacterStyle/$ID/[No character style]",
							Children:              children,
						}},
					}},
				},
			}
			storyPath := "Stories/Story_" + storyID + ".xml"
			if err := pkg.AddStory(storyPath, st, idmlpkg.ValidationOptions{}); err != nil {
				return "", fmt.Errorf("error al agregar Story %q: %w", frame.Name, err)
			}
			if err := registerStory(pkg, storyPath, storyID); err != nil {
				return "", err
			}
		}
	}

	spreadXML.WriteString("\t</Spread>\n</idPkg:Spread>\n")

	spreadPath := "Spreads/Spread_" + spreadID + ".xml"
	pkg.SetFileData(spreadPath, []byte(spreadXML.String()))

	dmDoc, err := pkg.Document()
	if err != nil {
		return "", err
	}
	dmDoc.Spreads = append(dmDoc.Spreads, document.ResourceRef{
		XMLName: xml.Name{
			Space: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
			Local: "Spread",
		},
		Src: spreadPath,
	})

	return spreadPath, nil
}

func addTextFrame(pkg *idmlpkg.Package, reg *idgen.Registry, frame frameSpec, halfHeight float64) error {
	frameID := reg.Generate()
	storyID := reg.Generate()

	topPt := mmToPt(frame.Bounds.TopMm)
	leftPt := mmToPt(frame.Bounds.LeftMm)
	bottomPt := mmToPt(frame.Bounds.BottomMm)
	rightPt := mmToPt(frame.Bounds.RightMm)

	w := rightPt - leftPt
	h := bottomPt - topPt
	centerX := leftPt + w/2
	centerY := topPt + h/2 - halfHeight

	extras := buildFrameExtras(frame.Options)

	tf := &spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{
			Self:          frameID,
			Name:          frame.Name,
			Visible:       "true",
			ItemLayer:     "uba",
			ItemTransform: "1 0 0 1 " + num(centerX) + " " + num(centerY),
		},
		ParentStory:        storyID,
		PreviousTextFrame:  "n",
		NextTextFrame:      "n",
		ContentType:        "TextType",
		AppliedObjectStyle: "ObjectStyle/$ID/[Normal Text Frame]",
		Properties: &common.Properties{
			PathGeometry: &common.PathGeometry{
				GeometryPathType: &common.GeometryPathType{
					PathOpen: "false",
					PathPointArray: &common.PathPointArray{
						PathPoints: boxPoints(w/2, h/2),
					},
				},
			},
		},
		OtherElements: extras,
	}

	if err := pkg.AddTextFrame(idmlpkg.PathSpread, tf, idmlpkg.ValidationOptions{}); err != nil {
		return fmt.Errorf("error al agregar TextFrame %q: %w", frame.Name, err)
	}

	// Construir el contenido de la story.
	var children []story.CharacterChild
	if frame.Content != "" {
		if frame.Options.ContentIsRaw {
			// Contenido con instrucciones de proceso (ej: <?ACE 18?> para auto page number).
			children = []story.CharacterChild{{Content: &story.Content{Raw: frame.Content}}}
		} else {
			children = []story.CharacterChild{{Content: &story.Content{Text: frame.Content}}}
		}
	}

	st := &story.Story{
		DOMVersion: "20.4",
		StoryElement: story.StoryElement{
			Self:     storyID,
			UserText: "true",
			ParagraphStyleRanges: []story.ParagraphStyleRange{{
				AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
				CharacterStyleRanges: []story.CharacterStyleRange{{
					AppliedCharacterStyle: "CharacterStyle/$ID/[No character style]",
					Children:              children,
				}},
			}},
		},
	}
	storyPath := "Stories/Story_" + storyID + ".xml"
	if err := pkg.AddStory(storyPath, st, idmlpkg.ValidationOptions{}); err != nil {
		return fmt.Errorf("error al agregar Story %q: %w", frame.Name, err)
	}

	return registerStory(pkg, storyPath, storyID)
}

func registerStory(pkg *idmlpkg.Package, storyPath, storyID string) error {
	doc, err := pkg.Document()
	if err != nil {
		return fmt.Errorf("error al leer designmap: %w", err)
	}
	doc.Stories = append(doc.Stories, document.ResourceRef{
		XMLName: xml.Name{
			Space: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
			Local: "Story",
		},
		Src: storyPath,
	})
	if doc.StoryList == "" {
		doc.StoryList = storyID
	} else {
		doc.StoryList = doc.StoryList + " " + storyID
	}
	return nil
}

// --- Helpers de estructura IDML ---

func boxPoints(hw, hh float64) []common.PathPointType {
	corners := [4][2]float64{{-hw, -hh}, {-hw, hh}, {hw, hh}, {hw, -hh}}
	pts := make([]common.PathPointType, 0, 4)
	for _, c := range corners {
		a := num(c[0]) + " " + num(c[1])
		pts = append(pts, common.PathPointType{Anchor: a, LeftDirection: a, RightDirection: a})
	}
	return pts
}

func buildFrameExtras(opts frameOptions) []common.RawXMLElement {
	var tfpAttrs []xml.Attr

	tfpAttrs = append(tfpAttrs,
		xml.Attr{Name: xml.Name{Local: "TextColumnCount"}, Value: "1"},
		xml.Attr{Name: xml.Name{Local: "TextColumnFixedWidth"}, Value: "0"},
		xml.Attr{Name: xml.Name{Local: "TextColumnMaxWidth"}, Value: "0"},
	)

	if opts.VerticalJustification != "" {
		tfpAttrs = append(tfpAttrs,
			xml.Attr{Name: xml.Name{Local: "VerticalJustification"}, Value: opts.VerticalJustification},
		)
	}

	return []common.RawXMLElement{
		{
			XMLName: xml.Name{Local: "TextFramePreference"},
			Attrs:   tfpAttrs,
			Content: []byte(`<Properties><InsetSpacing type="list"><ListItem type="unit">0</ListItem><ListItem type="unit">0</ListItem><ListItem type="unit">0</ListItem><ListItem type="unit">0</ListItem></InsetSpacing></Properties>`),
		},
		{
			XMLName: xml.Name{Local: "TextWrapPreference"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "Inverse"}, Value: "false"},
				{Name: xml.Name{Local: "ApplyToMasterPageOnly"}, Value: "false"},
				{Name: xml.Name{Local: "TextWrapSide"}, Value: "BothSides"},
				{Name: xml.Name{Local: "TextWrapMode"}, Value: "None"},
			},
			Content: []byte(`<Properties><TextWrapOffset Top="0" Left="0" Bottom="0" Right="0" /></Properties>`),
		},
	}
}

func addGuides(pkg *idmlpkg.Package, reg *idgen.Registry, guides []guideSpec, spreadPath string) error {
	spreadData, err := pkg.GetFileData(spreadPath)
	if err != nil {
		return fmt.Errorf("error al leer bytes del spread: %w", err)
	}

	var guideXML strings.Builder
	for _, g := range guides {
		id := reg.Generate()
		orientation := "Vertical"
		if g.Orientation == "horizontal" {
			orientation = "Horizontal"
		}
		locationPt := mmToPt(g.LocationMm)

		fmt.Fprintf(&guideXML,
			"\t\t<Guide Self=%q OverriddenPageItemProps=\"\" Orientation=%q Location=%q FitToPage=\"true\" ViewThreshold=\"5\" Locked=\"false\" ItemLayer=\"uba\" PageIndex=\"0\" GuideType=\"Ruler\" GuideZone=\"1\">\n\t\t\t<Properties>\n\t\t\t\t<GuideColor type=\"enumeration\">LightGray</GuideColor>\n\t\t\t</Properties>\n\t\t</Guide>\n",
			id, orientation, num(locationPt))
	}

	// Insertar antes del cierre </Spread>
	content := string(spreadData)
	closeTag := "\t</Spread>"
	idx := strings.LastIndex(content, closeTag)
	if idx == -1 {
		return fmt.Errorf("no se encontró </Spread> en el spread")
	}

	newContent := content[:idx] + guideXML.String() + content[idx:]
	pkg.SetFileData(spreadPath, []byte(newContent))
	pkg.InvalidateCache(spreadPath)

	return nil
}

// --- main ---

func main() {
	outPath := flag.String("out", "", "ruta de salida del archivo IDML (por defecto: stdout)")
	flag.Parse()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: no se pudo leer stdin: %v\n", err)
		os.Exit(1)
	}

	var input documentInput
	if err := json.Unmarshal(data, &input); err != nil {
		if synErr, ok := err.(*json.SyntaxError); ok {
			fmt.Fprintf(os.Stderr, "error: JSON inválido en posición %d: %v\n", synErr.Offset, err)
		} else {
			fmt.Fprintf(os.Stderr, "error: JSON inválido: %v\n", err)
		}
		os.Exit(1)
	}

	if err := validate(&input); err != nil {
		fmt.Fprintf(os.Stderr, "error de validación: %v\n", err)
		os.Exit(1)
	}

	pkg, err := generate(&input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error de generación: %v\n", err)
		os.Exit(2)
	}

	if *outPath != "" {
		if err := idmlpkg.Write(pkg, *outPath); err != nil {
			fmt.Fprintf(os.Stderr, "error al escribir %s: %v\n", *outPath, err)
			os.Exit(2)
		}
	} else {
		if err := idmlpkg.WriteTo(pkg, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "error al escribir a stdout: %v\n", err)
			os.Exit(2)
		}
	}
}
