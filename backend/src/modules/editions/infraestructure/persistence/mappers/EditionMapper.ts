import { Edition } from "../../../domain/entities/Edition";
import { EditionEntity } from "../entities/EditionEntity";

export class EditionMapper {
    static toDomain(entity: EditionEntity): Edition {
        return new Edition(
            entity.id,
            entity.no_paginas,
            Number(entity.ancho_mm),
            Number(entity.alto_mm),
            entity.cuadros_ancho,
            entity.cuadros_alto,
            entity.facing_pages,
            Number(entity.margen_superior_mm),
            Number(entity.margen_inferior_mm),
            Number(entity.margen_izquierdo_mm),
            Number(entity.margen_derecho_mm)
        );
    }

    static toEntity(domain: Edition): EditionEntity {
        const entity = new EditionEntity();
        entity.id = domain.id;
        entity.no_paginas = domain.no_paginas;
        entity.ancho_mm = domain.ancho_mm;
        entity.alto_mm = domain.alto_mm;
        entity.cuadros_ancho = domain.cuadros_ancho;
        entity.cuadros_alto = domain.cuadros_alto;
        entity.facing_pages = domain.facing_pages;
        entity.margen_superior_mm = domain.margen_superior_mm;
        entity.margen_inferior_mm = domain.margen_inferior_mm;
        entity.margen_izquierdo_mm = domain.margen_izquierdo_mm;
        entity.margen_derecho_mm = domain.margen_derecho_mm;
        return entity;
    }
}
