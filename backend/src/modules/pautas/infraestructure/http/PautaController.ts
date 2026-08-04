import { Request, Response, NextFunction } from "express";
import { CreatePautaUseCase } from "../../application/use-cases/CreatePautaUseCase";
import { AssignPautaUseCase } from "../../application/use-cases/AssignPautaUseCase";
import { PautaRepository } from "../../domain/repositories/PautaRepository";

export class PautaController {
    constructor(
        private readonly createPautaUseCase: CreatePautaUseCase,
        private readonly assignPautaUseCase: AssignPautaUseCase,
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

    assign = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautaId = Number(req.params.id);
            const pauta = await this.assignPautaUseCase.execute({
                pautaId,
                pagina_id: req.body.pagina_id,
                ubicacion_cuadros_x: req.body.ubicacion_cuadros_x,
                ubicacion_cuadros_y: req.body.ubicacion_cuadros_y,
            });
            res.json(pauta);
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
