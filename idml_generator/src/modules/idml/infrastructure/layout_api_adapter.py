import requests
from ..domain.entities import LayoutDocument, SpreadPage, Frame
from ..domain.layout_fetcher_port import LayoutFetcherPort


class LayoutApiAdapter(LayoutFetcherPort):
    def __init__(self, base_url: str):
        self._base_url = base_url.rstrip("/")

    def fetch(self, edition_id: int) -> LayoutDocument:
        url = f"{self._base_url}/api/layout/{edition_id}"
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        data = response.json()
        return self._map(data)

    def _map(self, data: dict) -> LayoutDocument:
        edition = data["edition"]
        pages = [
            SpreadPage(
                page_id=p["id"],
                no_pagina=p["no_pagina"],
                frames=[
                    Frame(
                        pauta_id=f["id"],
                        descripcion=f["descripcion_pauta"],
                        top_mm=f["indesignBounds"]["topMm"],
                        left_mm=f["indesignBounds"]["leftMm"],
                        bottom_mm=f["indesignBounds"]["bottomMm"],
                        right_mm=f["indesignBounds"]["rightMm"],
                    )
                    for f in p["pautas"]
                ],
            )
            for p in data["pages"]
        ]
        return LayoutDocument(
            edition_id=edition["id"],
            ancho_mm=edition["ancho_mm"],
            alto_mm=edition["alto_mm"],
            no_paginas=edition["no_paginas"],
            facing_pages=edition.get("facing_pages", False),
            margen_superior_mm=edition.get("margen_superior_mm", 0),
            margen_inferior_mm=edition.get("margen_inferior_mm", 0),
            margen_izquierdo_mm=edition.get("margen_izquierdo_mm", 0),
            margen_derecho_mm=edition.get("margen_derecho_mm", 0),
            pages=pages,
        )
