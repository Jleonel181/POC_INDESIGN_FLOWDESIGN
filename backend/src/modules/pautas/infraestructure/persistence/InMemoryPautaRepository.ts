import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";

export class InMemoryPautaRepository implements PautaRepository {
    private pautas: Pauta[] = [];

    async findByPageId(pageId: number): Promise<Pauta[]> {
        return this.pautas.filter(p => p.paginaId === pageId);
    }

    async findAll(): Promise<Pauta[]> {
        return [...this.pautas];
    }

    async findUnassigned(): Promise<Pauta[]> {
        return this.pautas.filter(p => p.paginaId === null);
    }

    async save(pauta: Pauta): Promise<Pauta> {
        const id = this.pautas.length + 1;
        const saved = new Pauta(id, pauta.descripcion_pauta, pauta.cuadros_alto, pauta.cuadros_ancho, pauta.ubicacion_cuadros_x, pauta.ubicacion_cuadros_y, pauta.paginaId);
        this.pautas.push(saved);
        return saved;
    }
}
