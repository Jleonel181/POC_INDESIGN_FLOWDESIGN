import { UseCase } from "../../../../shared/application/UseCase";
import { GenerateEditionLayoutUseCase } from "./GenerateEditionLayoutUseCase";
import { IdmlGenerator } from "../../domain/ports/IdmlGenerator";
import { IdmlDocumentDTO, IdmlFrameDTO, IdmlGuideDTO, IdmlPageDTO } from "../dto/IdmlDocumentDTO";
import { LayoutContract } from "../contracts/LayoutContract";

interface Input {
    editionId: number;
    folio: boolean;
    /** Ruta absoluta al .idml plantilla y nombre del master spread a inyectar como cabecera. */
    masterSpreadSource?: {
        templatePath: string;
        masterSpreadName: string;
    };
}

/**
 * Traduce el Layout Contract del dominio editorial al formato genérico de idmlgen.
 * Toda la lógica de negocio vive aquí. idmlgen no sabe nada del dominio.
 */
export class GenerateIdmlUseCase implements UseCase<Input, Buffer> {
    constructor(
        private readonly generateLayoutUseCase: GenerateEditionLayoutUseCase,
        private readonly idmlGenerator: IdmlGenerator
    ) {}

    async execute(input: Input): Promise<Buffer> {
        const layout = await this.generateLayoutUseCase.execute({ editionId: input.editionId });
        const idmlDoc = this.translateToIdmlDocument(layout, input.masterSpreadSource);
        return this.idmlGenerator.generate(idmlDoc);
    }

    private translateToIdmlDocument(layout: LayoutContract, masterSpreadSource?: Input["masterSpreadSource"]): IdmlDocumentDTO {
        const { edition, pages } = layout;

        const idmlPages: IdmlPageDTO[] = pages.map((page) => {
            const frames: IdmlFrameDTO[] = page.pautas.map((pauta) => {
                let leftMm = pauta.indesignBounds.leftMm;
                let rightMm = pauta.indesignBounds.rightMm;

                if (edition.facing_pages && page.no_pagina > 1) {
                    const isRightPage = page.no_pagina % 2 !== 0;
                    if (isRightPage) {
                        leftMm -= edition.ancho_mm;
                        rightMm -= edition.ancho_mm;
                    }
                }

                const bounds = {
                    topMm: pauta.indesignBounds.topMm,
                    leftMm,
                    bottomMm: pauta.indesignBounds.bottomMm,
                    rightMm,
                };

                // Si la pauta es de tipo imagen y tiene datos base64, generar frame de imagen.
                if (pauta.content_type === "image" && pauta.image_base64) {
                    return {
                        type: "image" as const,
                        name: pauta.descripcion_pauta,
                        bounds,
                        imageBase64: pauta.image_base64,
                    };
                }

                return {
                    type: "text" as const,
                    name: pauta.descripcion_pauta,
                    bounds,
                    content: pauta.descripcion_pauta,
                    options: {},
                };
            });

            return { frames };
        });

        const doc: IdmlDocumentDTO = {
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
                // Cuando se usa folio (master spread con cabecera), las columnas de InDesign
                // son irrelevantes — la grilla editorial se define con guías. Usar 1 evita
                // confusión visual con las guías de columna moradas.
                columns: masterSpreadSource ? 1 : edition.cuadros_ancho,
                guides: this.buildGridGuides(edition),
            },
            pages: idmlPages,
        };

        if (masterSpreadSource) {
            doc.masterSpreadSource = masterSpreadSource;
        }

        return doc;
    }

    /**
     * Calcula las guías de la grilla editorial como líneas de InDesign.
     */
    private buildGridGuides(edition: LayoutContract["edition"]): IdmlGuideDTO[] {
        const guides: IdmlGuideDTO[] = [];

        const contentWidthMm = edition.ancho_mm - edition.margen_izquierdo_mm - edition.margen_derecho_mm;
        const contentHeightMm = edition.alto_mm - edition.margen_superior_mm - edition.margen_inferior_mm;

        const cellWidthMm = contentWidthMm / edition.cuadros_ancho;
        const cellHeightMm = contentHeightMm / edition.cuadros_alto;

        for (let i = 1; i < edition.cuadros_ancho; i++) {
            guides.push({
                orientation: "vertical",
                locationMm: edition.margen_izquierdo_mm + i * cellWidthMm,
            });
        }

        for (let i = 1; i < edition.cuadros_alto; i++) {
            guides.push({
                orientation: "horizontal",
                locationMm: edition.margen_superior_mm + i * cellHeightMm,
            });
        }

        return guides;
    }
}
