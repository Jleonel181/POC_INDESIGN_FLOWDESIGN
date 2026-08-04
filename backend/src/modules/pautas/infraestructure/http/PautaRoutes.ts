import { Router } from "express";
import { PautaController } from "./PautaController";

export function createPautaRoutes(pautaController: PautaController) {
    const router = Router();

    router.get("/", pautaController.listAll);
    router.get("/unassigned", pautaController.listUnassigned);
    router.post("/", pautaController.create);
    router.put("/:id/assign", pautaController.assign);
    router.put("/:id/unassign", pautaController.unassign);

    return router;
}
