import { Router } from "express";
import { VentasController } from "./VentasController";

export function createVentasRoutes(ventasController: VentasController): Router {
    const router = Router();
    router.get("/ads", ventasController.getAdsByDate);
    router.post("/import", ventasController.importByDate);
    return router;
}
