import { Router } from "express";
import { EditionController } from "./EditionController";

export function createEditionRoutes(editionController: EditionController) {
    const router = Router();

    router.post("/", editionController.create);

    return router;
}
