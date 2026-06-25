import express from "express";
import cors from "cors";
import { DataSource } from "typeorm";
import { createDependencyContainer } from "./dependency-container";
import { createLayoutRoutes } from "../modules/layout/infraestructure/http/LayoutRoutes";
import { createEditionRoutes } from "../modules/editions/infraestructure/http/EditionRoutes";
import { createEsbRoutes } from "../modules/esb/infraestructure/http/EsbRoutes";
import { notFoundMiddleware } from "../shared/infraestructure/http/not-found.middleware";
import { errorHandlerMiddleware } from "../shared/infraestructure/http/error-handler.middleware";
import { EnvironmentConfig } from "../config/environment.config";

export function createApp(dataSource: DataSource) {
  const app = express();
  const { cors: corsConfig } = EnvironmentConfig.getInstance().getAppConfig();

  app.use(cors({
    origin: corsConfig.allowedOrigins,
    methods: corsConfig.allowedMethods,
    allowedHeaders: corsConfig.allowedHeaders,
    credentials: corsConfig.credentials,
  }));
  app.use(express.json());

  const container = createDependencyContainer(dataSource);

  app.get("/health", (_req, res) => {
    res.json({
      status: "ok",
      service: "designflow-backend",
      database: dataSource.isInitialized ? "connected" : "disconnected"
    });
  });

  app.use("/api/layout", createLayoutRoutes(container.layoutController));
  app.use("/api/editions", createEditionRoutes(container.editionController));
  app.use("/api/esb", createEsbRoutes(container.esbController));

  app.use(notFoundMiddleware);
  app.use(errorHandlerMiddleware);

  return app;
}