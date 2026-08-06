package idmlgen

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/idml/idgen"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

var selfAttrRegex = regexp.MustCompile(`Self="([^"]+)"`)

// Generate crea un paquete IDML a partir del input validado.
func Generate(input *DocumentInput) (*idmlpkg.Package, error) {
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

	// Leer IDs de referencia dinámicamente del paquete generado.
	refs, err := readTemplateRefs(pkg)
	if err != nil {
		return nil, err
	}

	// Eliminar el TextFrame y Story placeholder de la plantilla.
	if refs.textFrameID != "" {
		_, _ = pkg.RemoveTextFrame(refs.spreadPath, refs.textFrameID, false)
	}
	if refs.storyPath != "" {
		_, _ = pkg.RemoveStory(refs.storyPath, false)
	}

	designmapDoc, err := pkg.Document()
	if err != nil {
		return nil, fmt.Errorf("error al leer designmap: %w", err)
	}

	// Limpiar la story placeholder del StoryList y de las refs.
	if refs.storyID != "" {
		designmapDoc.StoryList = strings.Replace(designmapDoc.StoryList, refs.storyID+" ", "", 1)
		designmapDoc.StoryList = strings.Replace(designmapDoc.StoryList, " "+refs.storyID, "", 1)
		designmapDoc.StoryList = strings.Replace(designmapDoc.StoryList, refs.storyID, "", 1)
		cleaned := make([]document.ResourceRef, 0, len(designmapDoc.Stories))
		for _, s := range designmapDoc.Stories {
			if s.Src != refs.storyPath {
				cleaned = append(cleaned, s)
			}
		}
		designmapDoc.Stories = cleaned
	}

	// Registrar todos los IDs que la plantilla ya contiene para evitar colisiones.
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
	type spreadGroup struct {
		pages       []PageSpec
		pageNumbers []int
	}

	var spreads []spreadGroup
	if doc.FacingPages && len(input.Pages) > 1 {
		// Portada sola
		spreads = append(spreads, spreadGroup{
			pages:       []PageSpec{input.Pages[0]},
			pageNumbers: []int{1},
		})
		// Pares
		for i := 1; i < len(input.Pages); i += 2 {
			if i+1 < len(input.Pages) {
				spreads = append(spreads, spreadGroup{
					pages:       []PageSpec{input.Pages[i], input.Pages[i+1]},
					pageNumbers: []int{i + 1, i + 2},
				})
			} else {
				// Contraportada sola
				spreads = append(spreads, spreadGroup{
					pages:       []PageSpec{input.Pages[i]},
					pageNumbers: []int{i + 1},
				})
			}
		}
	} else {
		for i, p := range input.Pages {
			spreads = append(spreads, spreadGroup{
				pages:       []PageSpec{p},
				pageNumbers: []int{i + 1},
			})
		}
	}

	// --- Generar cada spread ---
	for i, sg := range spreads {
		if i == 0 {
			// Primer spread: usa el de la plantilla.
			for _, frame := range sg.pages[0].Frames {
				if err := addTextFrame(pkg, reg, refs, frame, halfHeight); err != nil {
					return nil, err
				}
			}
			if len(doc.Guides) > 0 {
				if err := addGuidesToSpread(pkg, reg, refs, doc.Guides, refs.spreadPath); err != nil {
					return nil, err
				}
			}
		} else {
			// Spreads adicionales: construidos con structs tipados, guides incluidas.
			if _, err := addSpreadForPages(pkg, reg, refs, doc, sg.pages, sg.pageNumbers, halfHeight, doc.Guides); err != nil {
				return nil, err
			}
		}
	}

	return pkg, nil
}

// addSpreadForPages crea un spread con 1 o 2 páginas y lo registra en el paquete.
// Usa los structs tipados de pkg/spread en vez de construir XML a mano.
func addSpreadForPages(pkg *idmlpkg.Package, reg *idgen.Registry, refs *templateRefs, doc DocumentSpec, pages []PageSpec, pageNumbers []int, halfHeight float64, guides []GuideSpec) (string, error) {
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
	// ponytail: columnGutter hardcodeado a 12pt (default de InDesign). Para exponerlo,
	// agregar un campo ColumnGutterMm a DocumentInput y convertirlo aquí con mmToPt.
	// Por ahora no se necesita porque el backend no lo parametriza.
	columnGutter := 12.0
	columnWidth := (usableWidth - float64(columns-1)*columnGutter) / float64(columns)

	var positions []string
	for i := 0; i < columns; i++ {
		start := float64(i) * (columnWidth + columnGutter)
		positions = append(positions, num(start), num(start+columnWidth))
	}
	columnsPositions := strings.Join(positions, " ")

	// Construir las páginas
	spreadPages := make([]spread.Page, 0, pageCount)
	for pi := range pages {
		pageID := reg.Generate()
		pageNum := pageNumbers[pi]

		var pageTransformX float64
		if pageCount == 2 && pi == 0 {
			pageTransformX = -pageWidthPt
		}

		// Guías de la grilla editorial en cada página
		pageGuides := make([]spread.Guide, 0, len(guides))
		for _, g := range guides {
			guideID := reg.Generate()
			orientation := "Vertical"
			if g.Orientation == "horizontal" {
				orientation = "Horizontal"
			}
			pageGuides = append(pageGuides, spread.Guide{
				Self:                    guideID,
				OverriddenPageItemProps: "",
				Orientation:             orientation,
				Location:                num(mmToPt(g.LocationMm)),
				FitToPage:               "true",
				ViewThreshold:           "5",
				Locked:                  "false",
				ItemLayer:               refs.layerID, PageIndex: "0",
				GuideType: "Ruler",
				GuideZone: "1",
				Properties: &common.Properties{
					OtherElements: []common.RawXMLElement{{
						XMLName: xml.Name{Local: "GuideColor"},
						Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "enumeration"}},
						Content: []byte("LightGray"),
					}},
				},
			})
		}

		spreadPages = append(spreadPages, spread.Page{
			Self:                   pageID,
			TabOrder:               "",
			AppliedMaster:          masterSpreadSelf(refs.masterSpread),
			OverrideList:           "",
			MasterPageTransform:    "1 0 0 1 0 0",
			Name:                   fmt.Sprintf("%d", pageNum),
			AppliedTrapPreset:      "TrapPreset/$ID/kDefaultTrapStyleName",
			GeometricBounds:        "0 0 " + num(pageHeightPt) + " " + num(pageWidthPt),
			ItemTransform:          "1 0 0 1 " + num(pageTransformX) + " " + num(-halfHeight),
			AppliedAlternateLayout: refs.sectionID,
			LayoutRule:             "UseMaster",
			SnapshotBlendingMode:   "IgnoreLayoutSnapshots",
			OptionalPage:           "false",
			GridStartingPoint:      "TopOutside",
			UseMasterGrid:          "true",
			Properties: &common.Properties{
				OtherElements: []common.RawXMLElement{{
					XMLName: xml.Name{Local: "PageColor"},
					Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "enumeration"}},
					Content: []byte("UseMasterColor"),
				}},
			},
			Guides: pageGuides,
			MarginPreference: &spread.MarginPreference{
				ColumnCount:      fmt.Sprintf("%d", columns),
				ColumnGutter:     num(columnGutter),
				Top:              num(mmToPt(doc.Margins.Top)),
				Bottom:           num(mmToPt(doc.Margins.Bottom)),
				Left:             num(marginLeft),
				Right:            num(marginRight),
				ColumnDirection:  "Horizontal",
				ColumnsPositions: columnsPositions,
			},
		})
	}

	// Construir los TextFrames
	textFrames := make([]spread.SpreadTextFrame, 0)
	for pi, page := range pages {
		var pageTransformX float64
		if pageCount == 2 && pi == 0 {
			pageTransformX = -pageWidthPt
		}

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

			textFrames = append(textFrames, spread.SpreadTextFrame{
				PageItemBase: spread.PageItemBase{
					Self:          frameID,
					Name:          frame.Name,
					Visible:       "true",
					ItemLayer:     refs.layerID,
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
				OtherElements: buildFrameExtras(frame.Options),
			})

			// Crear la story asociada
			if err := createAndRegisterStory(pkg, storyID, frame); err != nil {
				return "", err
			}
		}
	}

	// Ensamblar el spread completo
	sp := &spread.Spread{
		DOMVersion: "20.4",
		InnerSpread: spread.SpreadElement{
			Self:                    spreadID,
			PageTransitionType:      "None",
			PageTransitionDirection: "NotApplicable",
			PageTransitionDuration:  "Medium",
			ShowMasterItems:         "true",
			PageCount:               fmt.Sprintf("%d", pageCount),
			BindingLocation:         "0",
			SpreadHidden:            "false",
			AllowPageShuffle:        "true",
			ItemTransform:           "1 0 0 1 0 0",
			FlattenerOverride:       "Default",
			FlattenerPreference: &spread.FlattenerPreference{
				LineArtAndTextResolution:    "300",
				GradientAndMeshResolution:   "150",
				ClipComplexRegions:          "false",
				ConvertAllStrokesToOutlines: "false",
				ConvertAllTextToOutlines:    "false",
				Properties: &common.Properties{
					OtherElements: []common.RawXMLElement{{
						XMLName: xml.Name{Local: "RasterVectorBalance"},
						Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "double"}},
						Content: []byte("50"),
					}},
				},
			},
			Pages: spreadPages,
		},
	}

	// Agregar text frames al spread
	for i := range textFrames {
		sp.InnerSpread.Append(&textFrames[i])
	}

	// Serializar y registrar en el paquete
	data, err := spread.MarshalSpread(sp)
	if err != nil {
		return "", fmt.Errorf("error al serializar spread: %w", err)
	}

	spreadPath := "Spreads/Spread_" + spreadID + ".xml"
	pkg.SetFileData(spreadPath, data)

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

// createAndRegisterStory crea una story para un frame y la registra en el designmap.
func createAndRegisterStory(pkg *idmlpkg.Package, storyID string, frame FrameSpec) error {
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
		return fmt.Errorf("error al agregar Story %q: %w", frame.Name, err)
	}
	return registerStory(pkg, storyPath, storyID)
}

// addTextFrame agrega un TextFrame al primer spread (el de la plantilla) usando la API de pkg/idml.
func addTextFrame(pkg *idmlpkg.Package, reg *idgen.Registry, refs *templateRefs, frame FrameSpec, halfHeight float64) error {
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

	tf := &spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{
			Self:          frameID,
			Name:          frame.Name,
			Visible:       "true",
			ItemLayer:     refs.layerID,
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
		OtherElements: buildFrameExtras(frame.Options),
	}

	if err := pkg.AddTextFrame(refs.spreadPath, tf, idmlpkg.ValidationOptions{}); err != nil {
		return fmt.Errorf("error al agregar TextFrame %q: %w", frame.Name, err)
	}

	return createAndRegisterStory(pkg, storyID, frame)
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

func boxPoints(hw, hh float64) []common.PathPointType {
	corners := [4][2]float64{{-hw, -hh}, {-hw, hh}, {hw, hh}, {hw, -hh}}
	pts := make([]common.PathPointType, 0, 4)
	for _, c := range corners {
		a := num(c[0]) + " " + num(c[1])
		pts = append(pts, common.PathPointType{Anchor: a, LeftDirection: a, RightDirection: a})
	}
	return pts
}

func buildFrameExtras(opts FrameOptions) []common.RawXMLElement {
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

// addGuidesToSpread agrega guías a la primera página del spread indicado usando la API tipada.
// Carga el spread como struct, agrega las guías a Page.Guides, y re-serializa.
func addGuidesToSpread(pkg *idmlpkg.Package, reg *idgen.Registry, refs *templateRefs, guides []GuideSpec, spreadPath string) error {
	sp, err := pkg.Spread(spreadPath)
	if err != nil {
		return fmt.Errorf("error al cargar spread %s: %w", spreadPath, err)
	}

	if len(sp.InnerSpread.Pages) == 0 {
		return fmt.Errorf("el spread %s no tiene páginas", spreadPath)
	}

	for _, g := range guides {
		guideID := reg.Generate()
		orientation := "Vertical"
		if g.Orientation == "horizontal" {
			orientation = "Horizontal"
		}
		sp.InnerSpread.Pages[0].Guides = append(sp.InnerSpread.Pages[0].Guides, spread.Guide{
			Self:                    guideID,
			OverriddenPageItemProps: "",
			Orientation:             orientation,
			Location:                num(mmToPt(g.LocationMm)),
			FitToPage:               "true",
			ViewThreshold:           "5",
			Locked:                  "false",
			ItemLayer:               refs.layerID,
			PageIndex:               "0",
			GuideType:               "Ruler",
			GuideZone:               "1",
			Properties: &common.Properties{
				OtherElements: []common.RawXMLElement{{
					XMLName: xml.Name{Local: "GuideColor"},
					Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "enumeration"}},
					Content: []byte("LightGray"),
				}},
			},
		})
	}

	// Re-serializar y guardar
	data, err := spread.MarshalSpread(sp)
	if err != nil {
		return fmt.Errorf("error al serializar spread con guías: %w", err)
	}
	pkg.SetFileData(spreadPath, data)
	pkg.InvalidateCache(spreadPath)

	return nil
}
