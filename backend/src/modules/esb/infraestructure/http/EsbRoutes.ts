import { Router } from "express";
import { EsbController } from "./EsbController";

export function createEsbRoutes(esbController: EsbController): Router {
    const router = Router();
    router.get("/ads", esbController.getAdsByDate);
    return router;
}
