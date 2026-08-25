package idmlgen

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/idml/idgen"
	"github.com/dimelords/idmllib/v2/pkg/idml/images"
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

	// Inyectar MasterSpread desde plantilla externa si se indicó.
	// Se hace antes de registrar IDs del paquete fuente para que las stories y
	// elementos del master queden registrados y no colisionen con los generados.
	if input.MasterSpreadSource != nil {
		masterSelf, err := injectMasterSpreadFromTemplate(input.MasterSpreadSource, pkg)
		if err != nil {
			return nil, err
		}
		// Sobreescribir la referencia que usan las páginas como AppliedMaster.
		refs.masterSpread = "MasterSpreads/MasterSpread_" + masterSelf + ".xml"

		// Actualizar AppliedMaster en la página del primer spread (viene de la plantilla
		// con el master original hardcodeado).
		firstSpread, err := pkg.Spread(refs.spreadPath)
		if err == nil && len(firstSpread.InnerSpread.Pages) > 0 {
			for i := range firstSpread.InnerSpread.Pages {
				firstSpread.InnerSpread.Pages[i].AppliedMaster = masterSelf
			}
			if data, err := spread.MarshalSpread(firstSpread); err == nil {
				pkg.SetFileData(refs.spreadPath, data)
				pkg.InvalidateCache(refs.spreadPath)
			}
		}

		// Registrar IDs del master inyectado para evitar colisiones.
		for _, file := range pkg.Files() {
			if !strings.HasPrefix(file, "MasterSpreads/") && !strings.HasPrefix(file, "Stories/Story_") {
				continue
			}
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
				switch frame.Type {
				case "image":
					if err := addImageFrame(pkg, reg, refs, frame, halfHeight, input.BaseDir); err != nil {
						return nil, err
					}
				default: // "text"
					if err := addTextFrame(pkg, reg, refs, frame, halfHeight); err != nil {
						return nil, err
					}
				}
			}
			if len(doc.Guides) > 0 {
				if err := addGuidesToSpread(pkg, reg, refs, doc.Guides, refs.spreadPath); err != nil {
					return nil, err
				}
			}
		} else {
			// Spreads adicionales: construidos con structs tipados, guides incluidas.
			if _, err := addSpreadForPages(pkg, reg, refs, doc, sg.pages, sg.pageNumbers, halfHeight, doc.Guides, input.BaseDir); err != nil {
				return nil, err
			}
		}
	}

	return pkg, nil
}

// addSpreadForPages crea un spread con 1 o 2 páginas y lo registra en el paquete.
// Usa los structs tipados de pkg/spread en vez de construir XML a mano.
func addSpreadForPages(pkg *idmlpkg.Package, reg *idgen.Registry, refs *templateRefs, doc DocumentSpec, pages []PageSpec, pageNumbers []int, halfHeight float64, guides []GuideSpec, baseDir string) (string, error) {
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

	// Construir los elementos de página (TextFrames para texto, Rectangles para imágenes)
	textFrames := make([]spread.SpreadTextFrame, 0)
	imageRects := make([]spread.Rectangle, 0)
	for pi, page := range pages {
		var pageTransformX float64
		if pageCount == 2 && pi == 0 {
			pageTransformX = -pageWidthPt
		}

		for _, frame := range page.Frames {
			frameID := reg.Generate()

			topPt := mmToPt(frame.Bounds.TopMm)
			leftPt := mmToPt(frame.Bounds.LeftMm)
			bottomPt := mmToPt(frame.Bounds.BottomMm)
			rightPt := mmToPt(frame.Bounds.RightMm)

			w := rightPt - leftPt
			h := bottomPt - topPt
			centerX := leftPt + w/2 + pageTransformX
			centerY := topPt + h/2 - halfHeight

			switch frame.Type {
			case "image":
				resolved, err := resolveImage(frame, w, h, baseDir)
				if err != nil {
					return "", err
				}
				imageID := reg.Generate()
				linkID := reg.Generate()
				rect := buildImageRectangle(frameID, imageID, linkID, frame.Name, refs.layerID, centerX, centerY, w, h, resolved)
				imageRects = append(imageRects, rect)

			default: // "text"
				storyID := reg.Generate()
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
	// Agregar image rectangles al spread
	for i := range imageRects {
		sp.InnerSpread.Append(&imageRects[i])
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

// extractAppliedMaster extrae el valor del atributo AppliedMaster de un MasterSpread
// parseado. En IDML real, AppliedMaster está en los <Page> hijos del master, no como
// atributo del <MasterSpread> mismo. Se toma del primer Page; si todos los Pages
// apuntan al mismo padre, basta con uno.
func extractAppliedMaster(ms *spread.MasterSpread) string {
	// Primero buscar en OtherAttrs por si algún template lo pone en el elemento raíz.
	for _, attr := range ms.InnerMasterSpread.OtherAttrs {
		if attr.Name.Local == "AppliedMaster" {
			return attr.Value
		}
	}
	// En IDML estándar, la referencia al padre está en los Pages del master.
	if len(ms.InnerMasterSpread.Pages) > 0 {
		return ms.InnerMasterSpread.Pages[0].AppliedMaster
	}
	return ""
}

// findMasterSpreadPathBySelf busca en el designmap del fuente la ruta del MasterSpread
// cuyo Self coincida con el ID dado. Retorna "" si no se encuentra.
func findMasterSpreadPathBySelf(srcPkg *idmlpkg.Package, srcDoc *document.Document, selfID string) string {
	for _, ref := range srcDoc.MasterSpreads {
		if masterSpreadSelf(ref.Src) == selfID {
			return ref.Src
		}
	}
	return ""
}

// masterChainEntry agrupa la ruta, el Self y los datos crudos de un master spread
// de la cadena de herencia, en el orden en que deben inyectarse (raíz primero).
type masterChainEntry struct {
	path string // "MasterSpreads/MasterSpread_ub0.xml"
	self string // "ub0"
	data []byte // XML crudo del archivo
}

// collectMasterChain recorre la cadena de herencia de un master spread hacia arriba
// (vía AppliedMaster) y retorna la lista ordenada de raíz (sin padre) a hoja (el
// master en startPath). Detecta ciclos con un mapa de visitados.
func collectMasterChain(srcPkg *idmlpkg.Package, srcDoc *document.Document, startPath string) ([]masterChainEntry, error) {
	var chain []masterChainEntry
	visited := make(map[string]bool)
	currentPath := startPath

	for {
		data, err := srcPkg.GetFileData(currentPath)
		if err != nil {
			return nil, fmt.Errorf("error al leer %s: %w", currentPath, err)
		}

		ms, err := spread.ParseMasterSpread(data)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: %w", currentPath, err)
		}

		self := ms.InnerMasterSpread.Self
		if visited[self] {
			return nil, fmt.Errorf("ciclo detectado en la cadena de master spreads: %s", self)
		}
		visited[self] = true

		chain = append(chain, masterChainEntry{path: currentPath, self: self, data: data})

		parentID := extractAppliedMaster(ms)
		if parentID == "" || parentID == "n" {
			break // raíz alcanzada
		}

		// Buscar la ruta del padre en el designmap del fuente.
		parentPath := findMasterSpreadPathBySelf(srcPkg, srcDoc, parentID)
		if parentPath == "" {
			return nil, fmt.Errorf("master padre %q referenciado por %s no encontrado", parentID, currentPath)
		}
		currentPath = parentPath
	}

	// Invertir: la cadena se construyó de hoja a raíz, se necesita raíz a hoja.
	slices.Reverse(chain)
	return chain, nil
}

// injectMasterSpreadFromTemplate abre un IDML plantilla, busca el MasterSpread con
// el nombre indicado, y lo inyecta en el paquete destino junto con sus stories
// asociadas. Retorna el Self del master spread inyectado, que es lo que las páginas
// usan como AppliedMaster.
//
// Lo que copia:
//   - El archivo MasterSpreads/MasterSpread_XXX.xml (byte-a-byte del original)
//   - Cada Story referenciada por los TextFrames del master (ParentStory)
//   - La referencia en el designmap (MasterSpreads + Stories + StoryList)
//
// Lo que NO copia: recursos (estilos, colores, fuentes). Si el master usa recursos
// que no existen en la plantilla base, InDesign los mostrará como "missing". Para el
// caso de uso actual esto es aceptable porque el backend genera sobre la misma base.
//
// ponytail: sin resolución de recursos. El upgrade es extraer los recursos usados del
// paquete fuente y mergearlos en el destino (análogo a lo que hace pkg/idms/exporter).
func injectMasterSpreadFromTemplate(src *MasterSpreadSource, pkg *idmlpkg.Package) (string, error) {
	if src == nil || src.TemplatePath == "" || src.MasterSpreadName == "" {
		return "", fmt.Errorf("masterSpreadSource: templatePath y masterSpreadName son obligatorios")
	}

	// Abrir el IDML fuente.
	srcPkg, err := idmlpkg.Read(src.TemplatePath)
	if err != nil {
		return "", fmt.Errorf("error al abrir plantilla %s: %w", src.TemplatePath, err)
	}

	// Buscar el MasterSpread por nombre en el designmap del fuente.
	srcDoc, err := srcPkg.Document()
	if err != nil {
		return "", fmt.Errorf("error al leer designmap de la plantilla: %w", err)
	}

	var masterPath string
	for _, ref := range srcDoc.MasterSpreads {
		data, err := srcPkg.GetFileData(ref.Src)
		if err != nil {
			continue
		}
		// Parsear para leer el Name del MasterSpread.
		ms, err := spread.ParseMasterSpread(data)
		if err != nil {
			continue
		}
		if ms.InnerMasterSpread.Name == src.MasterSpreadName {
			masterPath = ref.Src
			break
		}
	}
	if masterPath == "" {
		return "", fmt.Errorf("no se encontró MasterSpread con Name=%q en %s", src.MasterSpreadName, src.TemplatePath)
	}

	// Recolectar la cadena completa de herencia (raíz → hoja).
	chain, err := collectMasterChain(srcPkg, srcDoc, masterPath)
	if err != nil {
		return "", err
	}

	// Registrar en el designmap del destino.
	dstDoc, err := pkg.Document()
	if err != nil {
		return "", fmt.Errorf("error al leer designmap destino: %w", err)
	}

	// Inyectar toda la cadena, de raíz a hoja.
	for _, entry := range chain {
		// Saltar si ya existe en el destino (deduplicación).
		if _, err := pkg.GetFileData(entry.path); err == nil {
			continue
		}

		pkg.SetFileData(entry.path, entry.data)

		dstDoc.MasterSpreads = append(dstDoc.MasterSpreads, document.ResourceRef{
			XMLName: xml.Name{
				Space: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
				Local: "MasterSpread",
			},
			Src: entry.path,
		})

		// Parsear para copiar stories referenciadas por TextFrames del master.
		ms, err := spread.ParseMasterSpread(entry.data)
		if err != nil {
			continue // datos ya validados en collectMasterChain
		}
		for _, tf := range ms.InnerMasterSpread.TextFrames() {
			if tf.ParentStory == "" || tf.ParentStory == "n" {
				continue
			}
			storyPath := "Stories/Story_" + tf.ParentStory + ".xml"
			storyData, err := srcPkg.GetFileData(storyPath)
			if err != nil {
				// La story puede no existir si es un frame vacío; se omite.
				continue
			}

			// Sustituir placeholder {{fecha}} con la fecha real del folio.
			if src.FolioDate != "" {
				storyData = bytes.ReplaceAll(storyData, []byte("{{fecha}}"), []byte(src.FolioDate))
			}

			pkg.SetFileData(storyPath, storyData)

			dstDoc.Stories = append(dstDoc.Stories, document.ResourceRef{
				XMLName: xml.Name{
					Space: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
					Local: "Story",
				},
				Src: storyPath,
			})
			if dstDoc.StoryList == "" {
				dstDoc.StoryList = tf.ParentStory
			} else {
				dstDoc.StoryList = dstDoc.StoryList + " " + tf.ParentStory
			}
		}
	}

	// El Self retornado es el de la hoja (el master solicitado).
	return chain[len(chain)-1].self, nil
}

// resolveImage usa el resolvedor de imágenes para obtener los datos de una imagen.
func resolveImage(frame FrameSpec, widthPt, heightPt float64, baseDir string) (*images.ResolvedImage, error) {
	resolver := images.NewResolver(images.ResolverOptions{
		BaseDir: baseDir,
	})
	source := images.ImageSource{
		Path:   frame.ImagePath,
		Base64: frame.ImageBase64,
	}
	bounds := images.FrameBounds{
		Width:  widthPt,
		Height: heightPt,
	}
	resolved, err := resolver.Resolve(source, bounds)
	if err != nil {
		return nil, fmt.Errorf("frame %q: %w", frame.Name, err)
	}
	return resolved, nil
}

// buildImageRectangle construye un Rectangle con una Image embebida.
func buildImageRectangle(rectID, imageID, linkID, name, layerID string, centerX, centerY, w, h float64, img *images.ResolvedImage) spread.Rectangle {
	return spread.Rectangle{
		PageItemBase: spread.PageItemBase{
			Self:          rectID,
			Name:          name,
			Visible:       "true",
			ItemLayer:     layerID,
			ItemTransform: "1 0 0 1 " + num(centerX) + " " + num(centerY),
		},
		ContentType:        "GraphicType",
		AppliedObjectStyle: "ObjectStyle/$ID/[Normal Graphics Frame]",
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
		Image: &spread.Image{
			FrameContentBase: spread.FrameContentBase{
				Self:               imageID,
				Visible:            "true",
				AppliedObjectStyle: "ObjectStyle/$ID/[None]",
				ItemTransform:      "1 0 0 1 " + num(-w/2) + " " + num(-h/2),
			},
			ActualPpi:    img.ActualPpi,
			EffectivePpi: img.EffectivePpi,
			Properties: &common.Properties{
				OtherElements: buildImageProperties(img, w, h),
			},
			Link: &spread.Link{
				Self:             linkID,
				LinkResourceURI:  img.LinkResourceURI,
				StoredState:      img.StoredState,
				LinkResourceSize: img.LinkResourceSize,
			},
		},
	}
}

// addImageFrame agrega un Rectangle con Image embebida al primer spread (el de la plantilla).
func addImageFrame(pkg *idmlpkg.Package, reg *idgen.Registry, refs *templateRefs, frame FrameSpec, halfHeight float64, baseDir string) error {
	rectID := reg.Generate()
	imageID := reg.Generate()
	linkID := reg.Generate()

	topPt := mmToPt(frame.Bounds.TopMm)
	leftPt := mmToPt(frame.Bounds.LeftMm)
	bottomPt := mmToPt(frame.Bounds.BottomMm)
	rightPt := mmToPt(frame.Bounds.RightMm)

	w := rightPt - leftPt
	h := bottomPt - topPt
	centerX := leftPt + w/2
	centerY := topPt + h/2 - halfHeight

	resolved, err := resolveImage(frame, w, h, baseDir)
	if err != nil {
		return err
	}

	rect := buildImageRectangle(rectID, imageID, linkID, frame.Name, refs.layerID, centerX, centerY, w, h, resolved)
	return pkg.AddRectangle(refs.spreadPath, &rect, idmlpkg.ValidationOptions{})
}

// buildImageProperties construye los hijos de Properties para una Image embebida.
// El orden es: Profile, Contents (base64), GraphicBounds — exactamente como lo emite InDesign.
func buildImageProperties(img *images.ResolvedImage, w, h float64) []common.RawXMLElement {
	elements := []common.RawXMLElement{
		{
			XMLName: xml.Name{Local: "Profile"},
			Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "string"}},
			Content: []byte("$ID/None"),
		},
	}

	// Solo incluir Contents si hay datos embebidos (base64).
	if img.Contents != "" {
		elements = append(elements, common.RawXMLElement{
			XMLName: xml.Name{Local: "Contents"},
			Content: []byte(img.Contents),
		})
	}

	elements = append(elements, common.RawXMLElement{
		XMLName: xml.Name{Local: "GraphicBounds"},
		Attrs: []xml.Attr{
			{Name: xml.Name{Local: "Left"}, Value: "0"},
			{Name: xml.Name{Local: "Top"}, Value: "0"},
			{Name: xml.Name{Local: "Right"}, Value: num(w)},
			{Name: xml.Name{Local: "Bottom"}, Value: num(h)},
		},
	})

	return elements
}
