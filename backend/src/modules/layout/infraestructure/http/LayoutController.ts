import { Request,Response,NextFunction } from "express";
import { GenerateEditionLayoutUseCase } from "../../application/use-cases/GenerateEditionLayoutUseCase";
import { GetAllEditionsLayoutUseCase } from "../../application/use-cases/GetAllEditionsLayoutUseCase";


export class LayoutController {

    constructor(
        private readonly generateEditionLayoutUseCase: GenerateEditionLayoutUseCase,
        private readonly getAllEditionsLayoutUseCase: GetAllEditionsLayoutUseCase
    ){}

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
    ): Promise <void> => {
        try{
            const editionId = Number(req.params.editionId);

            const layout = await this.generateEditionLayoutUseCase.execute({ editionId });

            res.json(layout);

        }catch(error){
            next(error);
        }
    }

}