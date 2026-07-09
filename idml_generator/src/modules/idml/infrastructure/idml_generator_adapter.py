import zipfile
import os
from ..domain.entities import LayoutDocument
from ..domain.ports import IdmlGeneratorPort
from .idml_xml_builder import (
    build_mimetype, build_container, build_designmap, build_spread, build_story,
    build_fonts, build_graphic, build_styles, build_preferences,
    build_tags, build_backing_story, _story_id,
)


class IdmlGeneratorAdapter(IdmlGeneratorPort):
    """
    Builds an IDML package (ZIP) from a LayoutDocument.

    Spread grouping (facing pages):
      - Page 1: alone
      - Pages 2-3, 4-5, 6-7...: pairs
    Single-page documents: each page is its own spread.
    """

    def generate(self, document: LayoutDocument, output_path: str) -> str:
        spread_groups = self._group_pages_into_spreads(document)
        spread_names = [f"Spreads/Spread_{i + 1}.xml" for i in range(len(spread_groups))]

        os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)

        with zipfile.ZipFile(output_path, "w", zipfile.ZIP_STORED) as zf:
            # mimetype MUST be first and uncompressed
            zf.writestr("mimetype", build_mimetype())
            zf.writestr("META-INF/container.xml", build_container())
            zf.writestr("designmap.xml", build_designmap(document, spread_names))
            zf.writestr("Resources/Fonts.xml", build_fonts())
            zf.writestr("Resources/Graphic.xml", build_graphic())
            zf.writestr("Resources/Styles.xml", build_styles())
            zf.writestr("Resources/Preferences.xml", build_preferences(document))
            zf.writestr("XML/Tags.xml", build_tags())
            zf.writestr("XML/BackingStory.xml", build_backing_story())

            for i, (spread_id, pages) in enumerate(spread_groups):
                zf.writestr(spread_names[i], build_spread(pages, document, spread_id))

            # One Story file per frame (not per page)
            for page in document.pages:
                for frame in page.frames:
                    zf.writestr(
                        f"Stories/Story_{_story_id(frame)}.xml",
                        build_story(frame, page),
                    )

        return output_path

    def _group_pages_into_spreads(self, document: LayoutDocument) -> list:
        pages = document.pages
        groups = []

        if not document.facing_pages or len(pages) <= 1:
            for page in pages:
                groups.append((f"sp{page.page_id}", [page]))
            return groups

        # Page 1 alone, then pairs
        groups.append((f"sp{pages[0].page_id}", [pages[0]]))
        i = 1
        while i < len(pages):
            if i + 1 < len(pages):
                groups.append((f"sp{pages[i].page_id}", [pages[i], pages[i + 1]]))
                i += 2
            else:
                groups.append((f"sp{pages[i].page_id}", [pages[i]]))
                i += 1

        return groups
