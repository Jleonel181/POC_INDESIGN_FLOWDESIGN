import { Request, Response, NextFunction } from "express";
import { GenerateEditionLayoutUseCase } from "../../application/use-cases/GenerateEditionLayoutUseCase";
import { GetAllEditionsLayoutUseCase } from "../../application/use-cases/GetAllEditionsLayoutUseCase";
import { GenerateIdmlUseCase } from "../../application/use-cases/GenerateIdmlUseCase";


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

            const idmlBuffer = await this.generateIdmlUseCase.execute({ editionId, folio });

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
}
