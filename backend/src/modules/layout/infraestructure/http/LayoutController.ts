import { Request, Response, NextFunction } from "express";
import { GenerateEditionLayoutUseCase } from "../../application/use-cases/GenerateEditionLayoutUseCase";
import { GetAllEditionsLayoutUseCase } from "../../application/use-cases/GetAllEditionsLayoutUseCase";
import { GenerateIdmlUseCase } from "../../application/use-cases/GenerateIdmlUseCase";
import { generateDummyPdf } from "../pdf/DummyPdfGenerator";
import { generateOverviewPdf } from "../pdf/OverviewPdfGenerator";

// Configuración de la plantilla de folio (MasterSpread).
// En Docker la imagen copia la plantilla a /app/templates/Pag_Ind.idml.
// En desarrollo local se sobreescribe con FOLIO_TEMPLATE_PATH apuntando al archivo real.
const FOLIO_TEMPLATE_PATH = process.env.FOLIO_TEMPLATE_PATH
    || "/app/templates/Pag_Ind.idml";
const FOLIO_MASTER_SPREAD_NAME = process.env.FOLIO_MASTER_SPREAD_NAME
    || "02-Noticias Apertura";

/** Formatea la fecha del día en español para el encabezado del folio. Ej: "Lunes, 25 de agosto de 2026" */
function formatFolioDate(): string {
    const now = new Date();
    const days = ["Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"];
    const months = ["enero", "febrero", "marzo", "abril", "mayo", "junio",
        "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"];
    return `${days[now.getDay()]}, ${now.getDate()} de ${months[now.getMonth()]} de ${now.getFullYear()}`;
}


export class LayoutController {

    constructor(
        private readonly generateEditionLayoutUseCase: GenerateEditionLayoutUseCase,
        private readonly getAllEditionsLayoutUseCase: GetAllEditionsLayoutUseCase,
        private readonly generateIdmlUseCase: GenerateIdmlUseCase
    ) {}

    getAllEditions = async (
        _req: Request,
        res: Response,
        next: NextFunction
    ): Promise<void> => {
        try {
            const layouts = await this.getAllEditionsLayoutUseCase.execute();
            res.json(layouts);
        } catch (error) {
            next(error);
        }
    }

    generateEditionId = async (
        req: Request,
        res: Response,
        next: NextFunction
    ): Promise<void> => {
        try {
            const editionId = Number(req.params.editionId);
            const layout = await this.generateEditionLayoutUseCase.execute({ editionId });
            res.json(layout);
        } catch (error) {
            next(error);
        }
    }

    generateIdml = async (
        req: Request,
        res: Response,
        next: NextFunction
    ): Promise<void> => {
        try {
            const editionId = Number(req.params.editionId);
            const folio = req.query.folio === "true";

            const masterSpreadSource = folio
                ? { templatePath: FOLIO_TEMPLATE_PATH, masterSpreadName: FOLIO_MASTER_SPREAD_NAME, folioDate: formatFolioDate() }
                : undefined;

            const idmlBuffer = await this.generateIdmlUseCase.execute({
                editionId,
                folio,
                masterSpreadSource
            });

            res.set({
                "Content-Type": "application/octet-stream",
                "Content-Disposition": `attachment; filename="edicion-${editionId}.idml"`,
                "Content-Length": idmlBuffer.length.toString()
            });
            res.send(idmlBuffer);
        } catch (error) {
            next(error);
        }
    }

    generateDummyPdf = async (
        req: Request,
        res: Response,
        next: NextFunction
    ): Promise<void> => {
        try {
            const editionId = Number(req.params.editionId);
            const layout = await this.generateEditionLayoutUseCase.execute({ editionId });
            const pdfBuffer = await generateDummyPdf(layout);

            res.set({
                "Content-Type": "application/pdf",
                "Content-Disposition": `attachment; filename="dummy-edicion-${editionId}.pdf"`,
                "Content-Length": pdfBuffer.length.toString()
            });
            res.send(pdfBuffer);
        } catch (error) {
            next(error);
        }
    }

    generateOverviewPdf = async (
        req: Request,
        res: Response,
        next: NextFunction
    ): Promise<void> => {
        try {
            const editionId = Number(req.params.editionId);
            const layout = await this.generateEditionLayoutUseCase.execute({ editionId });
            const pdfBuffer = await generateOverviewPdf(layout);

            res.set({
                "Content-Type": "application/pdf",
                "Content-Disposition": `attachment; filename="overview-edicion-${editionId}.pdf"`,
                "Content-Length": pdfBuffer.length.toString()
            });
            res.send(pdfBuffer);
        } catch (error) {
            next(error);
        }
    }
}
