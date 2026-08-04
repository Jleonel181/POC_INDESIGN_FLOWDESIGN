import { Request, Response, NextFunction } from "express";
import { CreatePautaUseCase } from "../../application/use-cases/CreatePautaUseCase";
import { PautaRepository } from "../../domain/repositories/PautaRepository";

export class PautaController {
    constructor(
        private readonly createPautaUseCase: CreatePautaUseCase,
        private readonly pautaRepository: PautaRepository
    ) {}

    create = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pauta = await this.createPautaUseCase.execute(req.body);
            res.status(201).json(pauta);
        } catch (error) {
            next(error);
        }
    }

    listAll = async (_req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautas = await this.pautaRepository.findAll();
            res.json(pautas);
        } catch (error) {
            next(error);
        }
    }

    listUnassigned = async (_req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautas = await this.pautaRepository.findUnassigned();
            res.json(pautas);
        } catch (error) {
            next(error);
        }
    }
}
