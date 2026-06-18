import { Pauta } from "../entities/Pauta";


export interface PautaRepository {
    findByPageId(pageId: number): Promise<Pauta[]>;
}