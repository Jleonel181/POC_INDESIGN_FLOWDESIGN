import { UseCase } from "../../../../shared/application/UseCase";
import { GenerateEditionLayoutUseCase } from "./GenerateEditionLayoutUseCase";
import { IdmlGenerator } from "../../domain/ports/IdmlGenerator";
import { IdmlDocumentDTO, IdmlFrameDTO, IdmlGuideDTO, IdmlPageDTO } from "../dto/IdmlDocumentDTO";
import { LayoutContract } from "../contracts/LayoutContract";

interface Input {
    editionId: number;
    folio: boolean;
}

/**
 * Traduce el Layout Contract del dominio editorial al formato genérico de idmlgen.
 * Toda la lógica de negocio (qué es un folio, cómo se posiciona, qué texto lleva)
 * vive aquí. idmlgen no sabe nada de esto.
 */
export class GenerateIdmlUseCase implements UseCase<Input, Buffer> {
    constructor(
        private readonly generateLayoutUseCase: GenerateEditionLayoutUseCase,
        private readonly idmlGenerator: IdmlGenerator
    ) {}

    async execute(input: Input): Promise<Buffer> {
        const layout = await this.generateLayoutUseCase.execute({ editionId: input.editionId });
        const idmlDoc = this.translateToIdmlDocument(layout, input.folio);
        return this.idmlGenerator.generate(idmlDoc);
    }

    private translateToIdmlDocument(layout: LayoutContract, folio: boolean): IdmlDocumentDTO {
        const { edition, pages } = layout;

        const idmlPages: IdmlPageDTO[] = pages.map((page) => {
            const frames: IdmlFrameDTO[] = page.pautas.map((pauta) => ({
                type: "text" as const,
                name: pauta.descripcion_pauta,
                bounds: {
                    topMm: pauta.indesignBounds.topMm,
                    leftMm: pauta.indesignBounds.leftMm,
                    bottomMm: pauta.indesignBounds.bottomMm,
                    rightMm: pauta.indesignBounds.rightMm,
                },
                content: pauta.descripcion_pauta,
                options: {},
            }));

            // Folio: un marco más al pie de la página, con número automático.
            if (folio) {
                frames.push(this.buildFolioFrame(edition, page.no_pagina));
            }

            return { frames };
        });

        return {
            document: {
                widthMm: edition.ancho_mm,
                heightMm: edition.alto_mm,
                margins: {
                    top: edition.margen_superior_mm,
                    bottom: edition.margen_inferior_mm,
                    left: edition.margen_izquierdo_mm,
                    right: edition.margen_derecho_mm,
                },
                facingPages: edition.facing_pages,
                columns: edition.cuadros_ancho,
                guides: this.buildGridGuides(edition),
            },
            pages: idmlPages,
        };
    }

    /**
     * Construye el frame del folio para una página.
     * El folio se posiciona al pie del área de contenido, con el ancho completo.
     * Usa <?ACE 18?> que InDesign interpreta como número de página automático.
     */
    private buildFolioFrame(
        edition: LayoutContract["edition"],
        _pageNumber: number
    ): IdmlFrameDTO {
        const folioHeightMm = 5;
        const topMm = edition.alto_mm - edition.margen_inferior_mm - folioHeightMm;

        return {
            type: "text",
            name: "folio",
            bounds: {
                topMm,
                leftMm: edition.margen_izquierdo_mm,
                bottomMm: topMm + folioHeightMm,
                rightMm: edition.ancho_mm - edition.margen_derecho_mm,
            },
            content: "<?ACE 18?>",
            options: {
                verticalJustification: "BottomAlign",
                contentIsRaw: true,
            },
        };
    }

    /**
     * Calcula las guías de la grilla editorial como líneas de InDesign.
     * Genera guías verticales y horizontales que dividen el área de contenido
     * según cuadros_ancho y cuadros_alto de la edición.
     */
    private buildGridGuides(edition: LayoutContract["edition"]): IdmlGuideDTO[] {
        const guides: IdmlGuideDTO[] = [];

        const contentWidthMm = edition.ancho_mm - edition.margen_izquierdo_mm - edition.margen_derecho_mm;
        const contentHeightMm = edition.alto_mm - edition.margen_superior_mm - edition.margen_inferior_mm;

        const cellWidthMm = contentWidthMm / edition.cuadros_ancho;
        const cellHeightMm = contentHeightMm / edition.cuadros_alto;

        // Guías verticales (líneas interiores de la grilla, no los bordes)
        for (let i = 1; i < edition.cuadros_ancho; i++) {
            guides.push({
                orientation: "vertical",
                locationMm: edition.margen_izquierdo_mm + i * cellWidthMm,
            });
        }

        // Guías horizontales (líneas interiores de la grilla, no los bordes)
        for (let i = 1; i < edition.cuadros_alto; i++) {
            guides.push({
                orientation: "horizontal",
                locationMm: edition.margen_superior_mm + i * cellHeightMm,
            });
        }

        return guides;
    }
}
