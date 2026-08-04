import { Pauta } from "../entities/Pauta";

export interface PautaRepository {
    findById(id: number): Promise<Pauta | null>;
    findByPageId(pageId: number): Promise<Pauta[]>;
    findAll(): Promise<Pauta[]>;
    findUnassigned(): Promise<Pauta[]>;
    save(pauta: Pauta): Promise<Pauta>;
}
