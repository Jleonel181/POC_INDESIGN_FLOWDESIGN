import { UseCase } from "../../../../shared/application/UseCase";
import { EditionRepository } from "../../../editions/domain/repositories/EditionRepository";
import { PageRepository } from "../../../pages/domain/repositories/PageRepository";
import { PautaRepository } from "../../../pautas/domain/repositories/PautaRepository";
import { GridLayoutCalculator } from "../../domain/services/GridLayoutCalculator";
import { LayoutValidator } from "../../domain/services/LayoutValidator";
import { EditionLayoutResponseDTO } from "../dto/EditionLayoutResponseDTO";
import { Pauta } from "../../../pautas/domain/entities/Pauta";
import { assertLayoutContract } from "../contracts/LayoutContractValidator";


interface Input {
    editionId: number;
}

export class GenerateEditionLayoutUseCase implements UseCase<Input, EditionLayoutResponseDTO> {
    constructor(
        private readonly editionRepository: EditionRepository,
        private readonly pageRepository: PageRepository,
        private readonly pautaRepository: PautaRepository,
        private readonly calculator: GridLayoutCalculator,
        private readonly validator: LayoutValidator
    ) {}

    async execute(input: Input): Promise<EditionLayoutResponseDTO> {
        const edition = await this.editionRepository.findById(input.editionId);

        if (!edition) {
            throw new Error(`Edition with id ${input.editionId} not found`);
        }

        const pages = await this.pageRepository.findByEdicionId(edition.id);
        const pautasByPage = await Promise.all(
            pages.map((page) => this.pautaRepository.findByPageId(page.id))
        );

        const responsePages = pages.map((page, index) => {
            const pagePautas = pautasByPage[index] ?? [];

            pagePautas.forEach((pauta) => {
                this.validator.validatePautaInsideGrid(edition, pauta);
            });

            this.validator.validateNoOverlap(pagePautas);

            return {
                id: page.id,
                no_pagina: page.no_pagina,
                pautas: pagePautas.map((pauta: Pauta) => {
                    const bounds = this.calculator.calculate({
                        pageWidthMm: edition.ancho_mm,
                        pageHeightMm: edition.alto_mm,
                        gridColumns: edition.cuadros_ancho,
                        gridRows: edition.cuadros_alto,
                        gridX: pauta.ubicacion_cuadros_x,
                        gridY: pauta.ubicacion_cuadros_y,
                        gridWidth: pauta.cuadros_ancho,
                        gridHeight: pauta.cuadros_alto,
                        facingPages: edition.facing_pages,
                        pageNumber: page.no_pagina,
                        margenSuperiorMm: edition.margen_superior_mm,
                        margenInferiorMm: edition.margen_inferior_mm,
                        margenIzquierdoMm: edition.margen_izquierdo_mm,
                        margenDerechoMm: edition.margen_derecho_mm
                    });

                    return {
                        id: pauta.id,
                        descripcion_pauta: pauta.descripcion_pauta,
                        cuadros_alto: pauta.cuadros_alto,
                        cuadros_ancho: pauta.cuadros_ancho,
                        ubicacion_cuadros_x: pauta.ubicacion_cuadros_x,
                        ubicacion_cuadros_y: pauta.ubicacion_cuadros_y,
                        indesignBounds: {
                            topMm: bounds.yMm,
                            leftMm: bounds.xMm,
                            bottomMm: Number((bounds.yMm + bounds.heightMm).toFixed(2)),
                            rightMm: Number((bounds.xMm + bounds.widthMm).toFixed(2))
                        }
                    };
                })
            };
        });

        const response: EditionLayoutResponseDTO = {
            metadata: {
                version: "1.0.0",
                unit: "mm",
                coordinateSystem: "grid",
                origin: "top-left"
            },
            edition: {
                id: edition.id,
                no_paginas: edition.no_paginas,
                ancho_mm: edition.ancho_mm,
                alto_mm: edition.alto_mm,
                cuadros_ancho: edition.cuadros_ancho,
                cuadros_alto: edition.cuadros_alto,
                facing_pages: edition.facing_pages,
                margen_superior_mm: edition.margen_superior_mm,
                margen_inferior_mm: edition.margen_inferior_mm,
                margen_izquierdo_mm: edition.margen_izquierdo_mm,
                margen_derecho_mm: edition.margen_derecho_mm
            },
            pages: responsePages
        };

        assertLayoutContract(response);

        return response;
    }
}