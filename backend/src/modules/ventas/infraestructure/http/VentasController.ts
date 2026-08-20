import { Request, Response, NextFunction } from "express";
import { HttpEsbAdapter } from "./HttpEsbAdapter";
import { ImportPautasFromVentasUseCase } from "../../application/use-cases/ImportPautasFromVentasUseCase";

export class VentasController {
    constructor(
        private readonly esbAdapter: HttpEsbAdapter,
        private readonly importPautasUseCase: ImportPautasFromVentasUseCase
    ) {}

    getAdsByDate = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const { date } = req.query;

            if (!date || typeof date !== "string") {
                res.status(400).json({ error: "BAD_REQUEST", message: "Query param 'date' is required (YYYY-MM-DD)" });
                return;
            }

            const ads = await this.esbAdapter.getAdsByDate(date);
            res.json({ date, total: ads.length, ads });
        } catch (error) {
            next(error);
        }
    };

    importByDate = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const { date } = req.body;

            if (!date || typeof date !== "string") {
                res.status(400).json({ error: "BAD_REQUEST", message: "Body field 'date' is required (YYYY-MM-DD)" });
                return;
            }

            const result = await this.importPautasUseCase.execute({ date });
            res.status(201).json(result);
        } catch (error) {
            next(error);
        }
    };
}
