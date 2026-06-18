import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";


export class InMemoryPautaRepository implements PautaRepository {
    private readonly pautas: Pauta[] =[
        new Pauta(1, "Pauta 1", 2, 5, 0, 0, 1),
        new Pauta(2, "Pauta 2", 1, 3, 0, 2, 1),
        new Pauta(3, "Pauta 3", 1, 3, 0, 7, 1),
    ];

    async findByPageId(pageId: number): Promise<Pauta[]> {
        return this.pautas.filter(pauta => pauta.paginaId === pageId);
    }
}