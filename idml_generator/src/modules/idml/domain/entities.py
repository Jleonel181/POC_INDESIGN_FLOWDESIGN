from dataclasses import dataclass, field
from typing import List


@dataclass(frozen=True)
class Frame:
    pauta_id: int
    descripcion: str
    top_mm: float
    left_mm: float
    bottom_mm: float
    right_mm: float


@dataclass(frozen=True)
class SpreadPage:
    page_id: int
    no_pagina: int
    frames: List[Frame] = field(default_factory=list)


@dataclass(frozen=True)
class LayoutDocument:
    edition_id: int
    ancho_mm: float
    alto_mm: float
    no_paginas: int
    facing_pages: bool
    margen_superior_mm: float
    margen_inferior_mm: float
    margen_izquierdo_mm: float
    margen_derecho_mm: float
    pages: List[SpreadPage] = field(default_factory=list)
