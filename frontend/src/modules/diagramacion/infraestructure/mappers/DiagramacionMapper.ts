import { DiagramacionDTO } from "../../application/dto/DiagramacionDTO";
import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { Pauta } from "../../domain/entities/Pauta";
import { GridPosition } from "../../domain/value-objects/GridPosition";
import { GridSize } from "../../domain/value-objects/GridSize";

export class DiagramacionMapper {
  static toDomain(dto: DiagramacionDTO): { edition: Edition; pages: Page[] } {
    const edition = new Edition(
      dto.edition.id,
      dto.edition.no_paginas,
      dto.edition.ancho_mm,
      dto.edition.alto_mm,
      dto.edition.cuadros_ancho,
      dto.edition.cuadros_alto,
      dto.edition.facing_pages,
      dto.edition.margen_superior_mm,
      dto.edition.margen_inferior_mm,
      dto.edition.margen_izquierdo_mm,
      dto.edition.margen_derecho_mm
    );

    const pages = dto.pages.map(
      (pageDto) =>
        new Page(
          pageDto.id,
          pageDto.no_pagina,
          pageDto.pautas.map(
            (pautaDto) =>
              new Pauta(
                pautaDto.id,
                pautaDto.descripcion_pauta,
                new GridPosition(pautaDto.ubicacion_cuadros_x, pautaDto.ubicacion_cuadros_y),
                new GridSize(pautaDto.cuadros_ancho, pautaDto.cuadros_alto),
                pautaDto.indesignBounds
              )
          )
        )
    );

    return { edition, pages };
  }
}
