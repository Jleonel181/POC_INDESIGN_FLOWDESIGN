"""
Builds raw IDML XML content for each required file in the IDML package.
IDML units are points (pt). 1 mm = 2.834645669 pt.

Key findings from real InDesign IDML:
- DOMVersion must be "21.4" (InDesign 2026)
- META-INF/container.xml is required
- TextFrame position is encoded in ItemTransform (tx, ty) + PathGeometry PathPointArray
- GeometricBounds on TextFrame is NOT used for positioning
- StoryList in designmap lists story Self IDs (one per frame)
"""
from lxml import etree
from ..domain.entities import LayoutDocument, SpreadPage, Frame

DOM_VERSION = "21.4"
MM_TO_PT = 2.834645669
IDPKG = "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"
NSMAP = {"idPkg": IDPKG}


def _q(local: str) -> str:
    return f"{{{IDPKG}}}{local}"


def _root(local: str) -> etree._Element:
    root = etree.Element(_q(local), nsmap=NSMAP)
    root.set("DOMVersion", DOM_VERSION)
    return root


def _pt(mm_val: float) -> str:
    return f"{mm_val * MM_TO_PT:.4f}"


def _story_id(frame: Frame) -> str:
    return f"stf{frame.pauta_id}"


def _frame_id(frame: Frame) -> str:
    return f"fr{frame.pauta_id}"


def build_mimetype() -> bytes:
    return b"application/vnd.adobe.indesign-idml-package"


def build_container() -> bytes:
    root = etree.Element(
        "container",
        version="1.0",
        nsmap={None: "urn:oasis:names:tc:opendocument:xmlns:container"},
    )
    rootfiles = etree.SubElement(root, "rootfiles")
    rf = etree.SubElement(rootfiles, "rootfile")
    rf.set("full-path", "designmap.xml")
    rf.set("media-type", "text/xml")
    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_designmap(doc: LayoutDocument, spread_names: list) -> bytes:
    # Collect all story IDs (one per frame across all pages)
    all_story_ids = [
        _story_id(frame)
        for page in doc.pages
        for frame in page.frames
    ]

    root = etree.Element("Document", nsmap=NSMAP)
    root.set("DOMVersion", DOM_VERSION)
    root.set("Self", "d")
    root.set("StoryList", " ".join(all_story_ids) if all_story_ids else "")
    root.set("ActiveLayer", "layer1")
    root.set("ZeroPoint", "0 0")
    root.set("Name", f"edition_{doc.edition_id}.indd")

    etree.SubElement(root, _q("Graphic"), src="Resources/Graphic.xml")
    etree.SubElement(root, _q("Fonts"), src="Resources/Fonts.xml")
    etree.SubElement(root, _q("Styles"), src="Resources/Styles.xml")
    etree.SubElement(root, _q("Preferences"), src="Resources/Preferences.xml")
    etree.SubElement(root, _q("Tags"), src="XML/Tags.xml")

    layer = etree.SubElement(root, "Layer")
    layer.set("Self", "layer1")
    layer.set("Name", "DesignFlow Layout")
    layer.set("Visible", "true")
    layer.set("Locked", "false")
    layer.set("IgnoreWrap", "false")
    layer.set("ShowGuides", "true")
    layer.set("LockGuides", "false")
    layer.set("UI", "true")
    layer.set("Expendable", "true")
    layer.set("Printable", "true")

    section = etree.SubElement(root, "Section")
    section.set("Self", "section1")
    section.set("PageStart", f"pg{doc.pages[0].page_id}" if doc.pages else "pg1")
    section.set("Length", str(doc.no_paginas))
    section.set("PageNumberStart", "1")
    section.set("SectionPrefix", "")
    section.set("PageNumberStyle", "Arabic")
    section.set("ContinueNumbering", "false")
    section.set("IncludeSectionPrefix", "false")
    section.set("Marker", "")

    for name in spread_names:
        etree.SubElement(root, _q("Spread"), src=name)

    etree.SubElement(root, _q("BackingStory"), src="XML/BackingStory.xml")

    for page in doc.pages:
        for frame in page.frames:
            etree.SubElement(root, _q("Story"), src=f"Stories/Story_{_story_id(frame)}.xml")

    pi = etree.ProcessingInstruction("aid", 'style="50" type="document" readerVersion="6.0" featureSet="257" product="21.4(4)"')
    root.addprevious(pi)
    return etree.tostring(root.getroottree(), xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_master_spread_none(doc: LayoutDocument) -> bytes:
    """
    [None] master spread with 2 pages (left+right) for facing-pages documents,
    or 1 page for single-page documents. This satisfies InDesign's internal model
    without adding phantom pages to the document.
    """
    page_w_pt = doc.ancho_mm * MM_TO_PT
    page_h_pt = doc.alto_mm * MM_TO_PT

    root = _root("MasterSpread")
    page_count = 2 if doc.facing_pages else 1

    ms = etree.SubElement(root, "MasterSpread")
    ms.set("Self", "MasterSpread/$ID/[None]")
    ms.set("Name", "[None]")
    ms.set("NamePrefix", "")
    ms.set("BaseName", "[None]")
    ms.set("ShowMasterItems", "true")
    ms.set("PageCount", str(page_count))
    ms.set("OverriddenPageItemProps", "")
    ms.set("PrimaryTextFrame", "")
    ms.set("ItemTransform", "1 0 0 1 0 0")

    if doc.facing_pages:
        pages = [
            ("mspg_L", -page_w_pt),
            ("mspg_R", 0.0),
        ]
    else:
        pages = [("mspg_L", -page_w_pt / 2)]

    for self_id, x_offset in pages:
        pg = etree.SubElement(ms, "Page")
        pg.set("Self", self_id)
        pg.set("Name", "")
        pg.set("AppliedMaster", "")
        pg.set("OverrideList", "")
        pg.set("TabOrder", "")
        pg.set("MasterPageTransform", "1 0 0 1 0 0")
        pg.set("GridStartingPoint", "TopLeftCorner")
        pg.set("UseMasterGrid", "true")
        pg.set("ItemTransform", f"1 0 0 1 {x_offset:.4f} 0")
        pg.set("GeometricBounds", f"0 0 {page_h_pt:.4f} {page_w_pt:.4f}")
        pg.set("AppliedTrapPreset", "TrapPreset/$ID/kDefaultTrapStyleName")
        margin = etree.SubElement(pg, "MarginPreference")
        margin.set("ColumnCount", "1")
        margin.set("ColumnGutter", "0")
        margin.set("Top", _pt(doc.margen_superior_mm))
        margin.set("Bottom", _pt(doc.margen_inferior_mm))
        margin.set("Left", _pt(doc.margen_izquierdo_mm))
        margin.set("Right", _pt(doc.margen_derecho_mm))
        margin.set("ColumnDirection", "Horizontal")
        content_w = doc.ancho_mm - doc.margen_izquierdo_mm - doc.margen_derecho_mm
        margin.set("ColumnsPositions", f"0 {_pt(content_w)}")

    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_spread(spread_pages: list, doc: LayoutDocument, spread_id: str) -> bytes:
    """
    Spread coordinate system: origin at center of spread.
    - Single page:      page ItemTransform tx = -pageWidth/2
    - Left page (pair): page ItemTransform tx = -pageWidth
    - Right page (pair):page ItemTransform tx = 0

    TextFrame positioning (from real IDML):
      ItemTransform = "1 0 0 1 {tx} {ty}"  where tx,ty is top-left corner in spread coords
      PathGeometry PathPointArray = 4 corners relative to ItemTransform origin:
        top-left:     0        0
        bottom-left:  0        height
        bottom-right: width    height
        top-right:    width    0
    """
    page_w_pt = doc.ancho_mm * MM_TO_PT
    page_h_pt = doc.alto_mm * MM_TO_PT

    root = _root("Spread")

    spread_elem = etree.SubElement(root, "Spread")
    spread_elem.set("Self", spread_id)
    spread_elem.set("PageCount", str(len(spread_pages)))
    spread_elem.set("ShowMasterItems", "true")
    spread_elem.set("AllowPageShuffle", "true")
    spread_elem.set("FlattenerOverride", "Default")
    spread_elem.set("ItemTransform", "1 0 0 1 0 0")
    spread_elem.set("BindingLocation", "1" if len(spread_pages) == 2 else "0")
    spread_elem.set("SpreadHidden", "false")

    for i, sp in enumerate(spread_pages):
        if len(spread_pages) == 1:
            x_offset = -page_w_pt / 2
        elif i == 0:
            x_offset = -page_w_pt
        else:
            x_offset = 0.0

        page_elem = etree.SubElement(spread_elem, "Page")
        page_elem.set("Self", f"pg{sp.page_id}")
        page_elem.set("Name", str(sp.no_pagina))
        page_elem.set("AppliedMaster", "")
        page_elem.set("OverrideList", "")
        page_elem.set("TabOrder", "")
        page_elem.set("GridStartingPoint", "TopLeftCorner")
        page_elem.set("UseMasterGrid", "true")
        page_elem.set("ItemTransform", f"1 0 0 1 {x_offset:.4f} 0")
        page_elem.set("GeometricBounds", f"0 0 {page_h_pt:.4f} {page_w_pt:.4f}")
        page_elem.set("AppliedTrapPreset", "TrapPreset/$ID/kDefaultTrapStyleName")

        margin = etree.SubElement(page_elem, "MarginPreference")
        margin.set("ColumnCount", "1")
        margin.set("ColumnGutter", "0")
        margin.set("Top", _pt(doc.margen_superior_mm))
        margin.set("Bottom", _pt(doc.margen_inferior_mm))
        margin.set("Left", _pt(doc.margen_izquierdo_mm))
        margin.set("Right", _pt(doc.margen_derecho_mm))
        margin.set("ColumnDirection", "Horizontal")
        content_w = doc.ancho_mm - doc.margen_izquierdo_mm - doc.margen_derecho_mm
        margin.set("ColumnsPositions", f"0 {_pt(content_w)}")

        is_right_page = (len(spread_pages) == 2 and i == 1)
        for frame in sp.frames:
            _add_text_frame(spread_elem, frame, sp, x_offset, is_right_page, doc.ancho_mm)

    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def _add_text_frame(spread_elem: etree._Element, frame: Frame, page: SpreadPage,
                    page_x_offset: float, is_right_page: bool, page_w_mm: float) -> None:
    """
    Position via ItemTransform (tx=left in spread coords, ty=top in spread coords).
    PathGeometry defines the shape relative to ItemTransform origin: 0,0 to width,height.
    The backend adds pageWidth to leftMm/rightMm for right-side facing pages — normalize first.
    """
    left_mm = frame.left_mm - page_w_mm if is_right_page else frame.left_mm
    right_mm = frame.right_mm - page_w_mm if is_right_page else frame.right_mm

    top_pt = frame.top_mm * MM_TO_PT
    left_pt = left_mm * MM_TO_PT
    width_pt = (right_mm - left_mm) * MM_TO_PT
    height_pt = (frame.bottom_mm - frame.top_mm) * MM_TO_PT

    # Spread-absolute top-left corner
    tx = page_x_offset + left_pt
    ty = top_pt

    tf = etree.SubElement(spread_elem, "TextFrame")
    tf.set("Self", _frame_id(frame))
    tf.set("ParentStory", _story_id(frame))
    tf.set("PreviousTextFrame", "")
    tf.set("NextTextFrame", "")
    tf.set("ContentType", "TextType")
    tf.set("AppliedObjectStyle", "ObjectStyle/$ID/[Normal Text Frame]")
    tf.set("ItemLayer", "layer1")
    tf.set("Visible", "true")
    tf.set("Name", frame.descripcion)
    tf.set("ItemTransform", f"1 0 0 1 {tx:.4f} {ty:.4f}")
    tf.set("GeometricBounds", f"{ty:.4f} {tx:.4f} {ty + height_pt:.4f} {tx + width_pt:.4f}")

    props = etree.SubElement(tf, "Properties")
    pg = etree.SubElement(props, "PathGeometry")
    gpt = etree.SubElement(pg, "GeometryPathType")
    gpt.set("PathOpen", "false")
    ppa = etree.SubElement(gpt, "PathPointArray")
    for ax, ay in [(0, 0), (0, height_pt), (width_pt, height_pt), (width_pt, 0)]:
        pt_elem = etree.SubElement(ppa, "PathPointType")
        coord = f"{ax:.4f} {ay:.4f}"
        pt_elem.set("Anchor", coord)
        pt_elem.set("LeftDirection", coord)
        pt_elem.set("RightDirection", coord)

    tf_pref = etree.SubElement(tf, "TextFramePreference")
    tf_pref.set("TextColumnCount", "1")
    tf_pref.set("TextColumnGutter", "0")

    tw = etree.SubElement(tf, "TextWrapPreference")
    tw.set("Inverse", "false")
    tw.set("ApplyToMasterPageOnly", "false")
    tw.set("TextWrapSide", "BothSides")
    tw.set("TextWrapMode", "None")


def build_story(frame: Frame, page: SpreadPage) -> bytes:
    """One Story file per frame."""
    root = _root("Story")

    story = etree.SubElement(root, "Story")
    story.set("Self", _story_id(frame))
    story.set("UserText", "true")
    story.set("IsEndnoteStory", "false")
    story.set("AppliedTOCStyle", "n")
    story.set("TrackChanges", "false")
    story.set("StoryTitle", frame.descripcion)
    story.set("AppliedNamedGrid", "n")

    pref = etree.SubElement(story, "StoryPreference")
    pref.set("OpticalMarginAlignment", "false")
    pref.set("OpticalMarginSize", "12")
    pref.set("FrameType", "TextFrameType")
    pref.set("StoryOrientation", "Horizontal")
    pref.set("StoryDirection", "LeftToRightDirection")

    para = etree.SubElement(story, "ParagraphStyleRange")
    para.set("AppliedParagraphStyle", "ParagraphStyle/$ID/[No paragraph style]")
    char = etree.SubElement(para, "CharacterStyleRange")
    char.set("AppliedCharacterStyle", "CharacterStyle/$ID/[No character style]")
    etree.SubElement(char, "Content").text = frame.descripcion

    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_fonts() -> bytes:
    return etree.tostring(_root("Fonts"), xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_graphic() -> bytes:
    root = _root("Graphic")

    # Required system colors
    black = etree.SubElement(root, "Color")
    black.set("Self", "Color/Black")
    black.set("Model", "Process")
    black.set("Space", "CMYK")
    black.set("ColorValue", "0 0 0 100")
    black.set("ColorOverride", "Specialblack")
    black.set("Name", "Black")
    black.set("ColorEditable", "false")
    black.set("ColorRemovable", "false")
    black.set("Visible", "true")

    paper = etree.SubElement(root, "Color")
    paper.set("Self", "Color/Paper")
    paper.set("Model", "Process")
    paper.set("Space", "CMYK")
    paper.set("ColorValue", "0 0 0 0")
    paper.set("ColorOverride", "Specialpaper")
    paper.set("Name", "Paper")
    paper.set("ColorEditable", "false")
    paper.set("ColorRemovable", "false")
    paper.set("Visible", "true")

    reg = etree.SubElement(root, "Color")
    reg.set("Self", "Color/Registration")
    reg.set("Model", "Registration")
    reg.set("Space", "CMYK")
    reg.set("ColorValue", "100 100 100 100")
    reg.set("ColorOverride", "Specialregistration")
    reg.set("Name", "Registration")
    reg.set("ColorEditable", "false")
    reg.set("ColorRemovable", "false")
    reg.set("Visible", "false")

    none_sw = etree.SubElement(root, "Swatch")
    none_sw.set("Self", "Swatch/None")
    none_sw.set("Name", "None")
    none_sw.set("ColorEditable", "false")
    none_sw.set("ColorRemovable", "false")
    none_sw.set("Visible", "true")

    # Required built-in stroke styles
    for ss_name in ["Solid", "Dashed", "Wavy"]:
        ss = etree.SubElement(root, "StrokeStyle")
        ss.set("Self", f"StrokeStyle/$ID/{ss_name}")
        ss.set("Name", f"$ID/{ss_name}")

    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_styles() -> bytes:
    root = _root("Styles")

    rcsg = etree.SubElement(root, "RootCharacterStyleGroup")
    rcsg.set("Self", "u71")
    cs = etree.SubElement(rcsg, "CharacterStyle")
    cs.set("Self", "CharacterStyle/$ID/[No character style]")
    cs.set("Name", "$ID/[No character style]")
    cs.set("Imported", "false")

    rpsg = etree.SubElement(root, "RootParagraphStyleGroup")
    rpsg.set("Self", "u72")
    for self_id, name in [
        ("ParagraphStyle/$ID/[No paragraph style]", "$ID/[No paragraph style]"),
        ("ParagraphStyle/$ID/NormalParagraphStyle", "$ID/NormalParagraphStyle"),
    ]:
        ps = etree.SubElement(rpsg, "ParagraphStyle")
        ps.set("Self", self_id)
        ps.set("Name", name)

    rsg = etree.SubElement(root, "RootObjectStyleGroup")
    rsg.set("Self", "u73")
    for self_id, name in [
        ("ObjectStyle/$ID/[None]", "$ID/[None]"),
        ("ObjectStyle/$ID/[Normal Text Frame]", "$ID/[Normal Text Frame]"),
        ("ObjectStyle/$ID/[Normal Graphics Frame]", "$ID/[Normal Graphics Frame]"),
    ]:
        obj = etree.SubElement(rsg, "ObjectStyle")
        obj.set("Self", self_id)
        obj.set("Name", name)

    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_preferences(doc: LayoutDocument) -> bytes:
    root = _root("Preferences")

    dp = etree.SubElement(root, "DocumentPreference")
    dp.set("PageWidth", _pt(doc.ancho_mm))
    dp.set("PageHeight", _pt(doc.alto_mm))
    dp.set("FacingPages", str(doc.facing_pages).lower())
    dp.set("PagesPerDocument", str(doc.no_paginas))
    dp.set("DocumentBleedTopOffset", "0")
    dp.set("DocumentBleedBottomOffset", "0")
    dp.set("DocumentBleedInsideOrLeftOffset", "0")
    dp.set("DocumentBleedOutsideOrRightOffset", "0")

    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_tags() -> bytes:
    return etree.tostring(_root("Tags"), xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)


def build_backing_story() -> bytes:
    root = _root("BackingStory")
    xs = etree.SubElement(root, "XmlStory")
    xs.set("Self", "backing")
    xs.set("UserText", "true")
    xs.set("IsEndnoteStory", "false")
    xs.set("AppliedTOCStyle", "n")
    xs.set("TrackChanges", "false")
    xs.set("StoryTitle", "$ID/")
    xs.set("AppliedNamedGrid", "n")
    para = etree.SubElement(xs, "ParagraphStyleRange")
    para.set("AppliedParagraphStyle", "ParagraphStyle/$ID/NormalParagraphStyle")
    char = etree.SubElement(para, "CharacterStyleRange")
    char.set("AppliedCharacterStyle", "CharacterStyle/$ID/[No character style]")
    xe = etree.SubElement(para, "XMLElement")
    xe.set("Self", "backing_root")
    xe.set("MarkupTag", "XMLTag/Root")
    etree.SubElement(xe, "CharacterStyleRange").set(
        "AppliedCharacterStyle", "CharacterStyle/$ID/[No character style]"
    )
    return etree.tostring(root, xml_declaration=True, encoding="UTF-8", standalone=True, pretty_print=True)
