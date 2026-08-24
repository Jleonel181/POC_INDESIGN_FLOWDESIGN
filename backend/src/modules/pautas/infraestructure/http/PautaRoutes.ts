import { Router } from "express";
import multer from "multer";
import { PautaController } from "./PautaController";

const upload = multer({
    storage: multer.memoryStorage(),
    limits: { fileSize: 20 * 1024 * 1024 }, // 20 MB máximo
    fileFilter: (_req, file, cb) => {
        if (file.mimetype.startsWith("image/")) {
            cb(null, true);
        } else {
            cb(new Error("Solo se aceptan archivos de imagen"));
        }
    },
});

export function createPautaRoutes(pautaController: PautaController) {
    const router = Router();

    router.get("/", pautaController.listAll);
    router.get("/unassigned", pautaController.listUnassigned);
    router.post("/", pautaController.create);
    router.put("/:id/assign", pautaController.assign);
    router.put("/:id/unassign", pautaController.unassign);
    router.put("/:id/image", upload.single("file"), pautaController.setImage);
    router.get("/:id/image", pautaController.getImage);
    router.delete("/:id/image", pautaController.removeImage);

    return router;
}
