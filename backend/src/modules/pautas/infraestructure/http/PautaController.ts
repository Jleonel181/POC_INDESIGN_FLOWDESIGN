import { Request, Response, NextFunction } from "express";
import { CreatePautaUseCase } from "../../application/use-cases/CreatePautaUseCase";

export class PautaController {
    constructor(
        private readonly createPautaUseCase: CreatePautaUseCase
    ) {}

    create = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pauta = await this.createPautaUseCase.execute(req.body);
            res.status(201).json(pauta);
        } catch (error) {
            next(error);
        }
    }
}
