import { Router } from "express";
import { PautaController } from "./PautaController";

export function createPautaRoutes(pautaController: PautaController) {
    const router = Router();

    router.get("/", pautaController.listAll);
    router.get("/unassigned", pautaController.listUnassigned);
    router.post("/", pautaController.create);

    return router;
}
