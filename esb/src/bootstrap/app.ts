import express from "express";
import cors from "cors";
import { EnvironmentConfig } from "../config/environment.config";
import { MysqlAdDataSourceAdapter } from "../modules/ventas/infraestructure/adapters/MysqlAdDataSourceAdapter";
import { GetImportedAdsUseCase } from "../modules/ventas/application/use-cases/GetImportedAdsUseCase";
import { VentasController } from "../modules/ventas/infraestructure/http/VentasController";
import { createVentasRoutes } from "../modules/ventas/infraestructure/http/VentasRoutes";

export function createApp() {
    const app = express();
    const config = EnvironmentConfig.getInstance().getConfig();

    app.use(cors({ origin: config.corsAllowedOrigins, credentials: true }));
    app.use(express.json());

    app.get("/health", (_req, res) => {
        res.json({ status: "ok", service: "flowdesign-esb" });
    });

    // Ventas module
    const ventasDataSource = new MysqlAdDataSourceAdapter(config.ventasDatabase);
    const getImportedAdsUseCase = new GetImportedAdsUseCase(ventasDataSource);
    const ventasController = new VentasController(getImportedAdsUseCase);

    app.use("/api/ventas", createVentasRoutes(ventasController));

    app.use((_req, res) => res.status(404).json({ error: "NOT_FOUND" }));

    app.use((error: Error, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
        console.error(error);
        res.status(500).json({ error: "INTERNAL_SERVER_ERROR", message: error.message });
    });

    return app;
}
