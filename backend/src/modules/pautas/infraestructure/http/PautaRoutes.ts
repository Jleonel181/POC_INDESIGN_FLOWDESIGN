import { Router } from "express";
import { PautaController } from "./PautaController";

export function createPautaRoutes(pautaController: PautaController) {
    const router = Router();

    router.post("/", pautaController.create);

    return router;
}
