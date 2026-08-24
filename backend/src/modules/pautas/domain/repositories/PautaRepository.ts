import { Pauta } from "../entities/Pauta";

export interface PautaRepository {
    findById(id: number): Promise<Pauta | null>;
    findByPageId(pageId: number): Promise<Pauta[]>;
    findAll(): Promise<Pauta[]>;
    findByDate(date: string): Promise<Pauta[]>;
    findByDescription(desc: string): Promise<Pauta[]>;
    findUnassigned(): Promise<Pauta[]>;
    save(pauta: Pauta): Promise<Pauta>;
    updateImage(pautaId: number, imageBase64: string): Promise<void>;
    updateImageByDescription(desc: string, imageBase64: string): Promise<void>;
    removeImage(pautaId: number): Promise<void>;
}
