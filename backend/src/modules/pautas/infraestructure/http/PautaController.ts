import { Request, Response, NextFunction } from "express";
import sharp from "sharp";
import { CreatePautaUseCase } from "../../application/use-cases/CreatePautaUseCase";
import { AssignPautaUseCase } from "../../application/use-cases/AssignPautaUseCase";
import { UnassignPautaUseCase } from "../../application/use-cases/UnassignPautaUseCase";
import { PautaRepository } from "../../domain/repositories/PautaRepository";

export class PautaController {
    constructor(
        private readonly createPautaUseCase: CreatePautaUseCase,
        private readonly assignPautaUseCase: AssignPautaUseCase,
        private readonly unassignPautaUseCase: UnassignPautaUseCase,
        private readonly pautaRepository: PautaRepository
    ) {}

    create = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pauta = await this.createPautaUseCase.execute(req.body);
            res.status(201).json(pauta);
        } catch (error) {
            next(error);
        }
    }

    assign = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautaId = Number(req.params.id);
            const pauta = await this.assignPautaUseCase.execute({
                pautaId,
                pagina_id: req.body.pagina_id,
                ubicacion_cuadros_x: req.body.ubicacion_cuadros_x,
                ubicacion_cuadros_y: req.body.ubicacion_cuadros_y,
            });
            res.json(pauta);
        } catch (error) {
            next(error);
        }
    }

    /**
     * GET /pautas?date=YYYY-MM-DD (opcional)
     * Sin date: devuelve todas. Con date: filtra por created_at de ese día.
     */
    listAll = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const date = req.query.date as string | undefined;
            let pautas;
            if (date && /^\d{4}-\d{2}-\d{2}$/.test(date)) {
                pautas = await this.pautaRepository.findByDate(date);
            } else {
                pautas = await this.pautaRepository.findAll();
            }
            res.json(pautas);
        } catch (error) {
            next(error);
        }
    }

    listUnassigned = async (_req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautas = await this.pautaRepository.findUnassigned();
            res.json(pautas);
        } catch (error) {
            next(error);
        }
    }

    unassign = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautaId = Number(req.params.id);
            const pauta = await this.unassignPautaUseCase.execute(pautaId);
            res.json(pauta);
        } catch (error) {
            next(error);
        }
    }

    /**
     * PUT /pautas/:id/image
     * Acepta multipart/form-data con un campo "file" (imagen en cualquier formato).
     * Convierte a JPG y guarda como base64 en la DB local.
     * También acepta JSON con { imageBase64 } para compatibilidad con el frontend.
     */
    setImage = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautaId = Number(req.params.id);

            const existing = await this.pautaRepository.findById(pautaId);
            if (!existing) {
                res.status(404).json({ error: `Pauta ${pautaId} no encontrada` });
                return;
            }

            let jpgBase64: string;

            if (req.file) {
                // Multipart file upload — convertir a JPG con sharp
                const jpgBuffer = await sharp(req.file.buffer)
                    .jpeg({ quality: 85 })
                    .toBuffer();
                jpgBase64 = jpgBuffer.toString("base64");
            } else if (req.body.imageBase64) {
                // JSON body con base64 — convertir a JPG
                const inputBuffer = Buffer.from(req.body.imageBase64, "base64");
                const jpgBuffer = await sharp(inputBuffer)
                    .jpeg({ quality: 85 })
                    .toBuffer();
                jpgBase64 = jpgBuffer.toString("base64");
            } else {
                res.status(400).json({ error: "Se requiere un archivo (campo 'file') o 'imageBase64' en el body" });
                return;
            }

            await this.pautaRepository.updateImage(pautaId, jpgBase64);
            res.json({ id: pautaId, content_type: "image", message: "Imagen establecida" });
        } catch (error) {
            next(error);
        }
    }

    /**
     * DELETE /pautas/:id/image
     * Remueve la imagen y vuelve content_type a "text".
     */
    removeImage = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautaId = Number(req.params.id);
            await this.pautaRepository.removeImage(pautaId);
            res.json({ id: pautaId, content_type: "text", message: "Imagen removida" });
        } catch (error) {
            next(error);
        }
    }

    /**
     * GET /pautas/:id/image
     * Devuelve la imagen como JPEG binary (para preview en gestión de artes).
     */
    getImage = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
        try {
            const pautaId = Number(req.params.id);
            const pauta = await this.pautaRepository.findById(pautaId);

            if (!pauta || pauta.content_type !== "image" || !pauta.image_base64) {
                res.status(404).json({ error: "Esta pauta no tiene imagen asignada" });
                return;
            }

            const buffer = Buffer.from(pauta.image_base64, "base64");
            res.set({ "Content-Type": "image/jpeg", "Content-Length": buffer.length.toString() });
            res.send(buffer);
        } catch (error) {
            next(error);
        }
    }
}
