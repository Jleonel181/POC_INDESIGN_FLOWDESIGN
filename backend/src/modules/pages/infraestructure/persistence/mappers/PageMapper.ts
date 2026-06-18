import { Page } from "../../../domain/entities/Page";
import { PageEntity } from "../entities/PageEntity";

export class PageMapper {
    static toDomain(entity: PageEntity): Page {
        return new Page(
            entity.id,
            entity.no_pagina,
            entity.edicion_id
        );
    }

    static toEntity(domain: Page): PageEntity {
        const entity = new PageEntity();
        entity.id = domain.id;
        entity.no_pagina = domain.no_pagina;
        entity.edicion_id = domain.edicionId;
        return entity;
    }
}
