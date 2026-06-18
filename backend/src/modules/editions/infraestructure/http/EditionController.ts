import { Request, Response, NextFunction } from "express";
import { CreateEditionUseCase } from "../../application/use-cases/CreateEditionUseCase";

export class EditionController {
    constructor(
        private readonly createEditionUseCase: CreateEditionUseCase
    ) {}

    create = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const result = await this.createEditionUseCase.execute(req.body);
            res.status(201).json(result);
        } catch (error) {
            next(error);
        }
    }
}
