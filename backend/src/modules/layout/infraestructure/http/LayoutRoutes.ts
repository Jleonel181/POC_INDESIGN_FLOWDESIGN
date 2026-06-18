import { Router } from "express";
import { LayoutController } from "./LayoutController";

export function createLayoutRoutes(layoutController: LayoutController) {
	const router = Router();

	router.get("/", layoutController.getAllEditions);
	router.get("/:editionId", layoutController.generateEditionId);

	return router;
}
