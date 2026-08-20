import { Router } from "express";
import { LayoutController } from "./LayoutController";

export function createLayoutRoutes(layoutController: LayoutController) {
	const router = Router();

	router.get("/", layoutController.getAllEditions);
	router.get("/:editionId/idml", layoutController.generateIdml);
	router.get("/:editionId/pdf", layoutController.generateDummyPdf);
	router.get("/:editionId/overview", layoutController.generateOverviewPdf);
	router.get("/:editionId", layoutController.generateEditionId);

	return router;
}
