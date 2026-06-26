import { Request, Response, NextFunction } from "express";
import { GetImportedAdsUseCase } from "../../application/use-cases/GetImportedAdsUseCase";

export class VentasController {
    constructor(private readonly getImportedAds: GetImportedAdsUseCase) {}

    getAdsByDate = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const { date } = req.query;

            if (!date || typeof date !== "string") {
                res.status(400).json({ error: "BAD_REQUEST", message: "Query param 'date' is required (YYYY-MM-DD)" });
                return;
            }

            const ads = await this.getImportedAds.execute(date);
            res.json({ date, total: ads.length, ads });
        } catch (error) {
            next(error);
        }
    };
}
