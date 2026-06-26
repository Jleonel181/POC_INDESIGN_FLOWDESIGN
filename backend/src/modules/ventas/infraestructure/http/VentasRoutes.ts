import { Router } from "express";
import { VentasController } from "./VentasController";

export function createVentasRoutes(ventasController: VentasController): Router {
    const router = Router();
    router.get("/ads", ventasController.getAdsByDate);
    return router;
}
