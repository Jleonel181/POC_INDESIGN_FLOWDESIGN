import { UseCase } from "../../../../shared/application/UseCase";
import { Edition } from "../../domain/entities/Edition";
import { EditionRepository } from "../../domain/repositories/EditionRepository";
import { PageRepository } from "../../../pages/domain/repositories/PageRepository";
import { Page } from "../../../pages/domain/entities/Page";
import { CreateEditionDTO } from "../dto/CreateEditionDTO";

interface Output {
    edition: Edition;
    pages: Page[];
}

export class CreateEditionUseCase implements UseCase<CreateEditionDTO, Output> {
    constructor(
        private readonly editionRepository: EditionRepository,
        private readonly pageRepository: PageRepository
    ) {}

    async execute(input: CreateEditionDTO): Promise<Output> {
        const edition = new Edition(
            0,
            input.no_paginas,
            input.ancho_mm,
            input.alto_mm,
            input.cuadros_ancho,
            input.cuadros_alto,
            input.facing_pages ?? false,
            input.margen_superior_mm ?? 0,
            input.margen_inferior_mm ?? 0,
            input.margen_izquierdo_mm ?? 0,
            input.margen_derecho_mm ?? 0
        );

        const savedEdition = await this.editionRepository.save(edition);

        const pages = Array.from({ length: input.no_paginas }, (_, i) =>
            new Page(0, i + 1, savedEdition.id)
        );

        const savedPages = await this.pageRepository.saveMany(pages);

        return { edition: savedEdition, pages: savedPages };
    }
}
