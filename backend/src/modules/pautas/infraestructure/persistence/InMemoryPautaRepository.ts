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

    async save(pauta: Pauta): Promise<Pauta> {
        const id = this.pautas.length + 1;
        const saved = new Pauta(id, pauta.descripcion_pauta, pauta.cuadros_alto, pauta.cuadros_ancho, pauta.ubicacion_cuadros_x, pauta.ubicacion_cuadros_y, pauta.paginaId);
        (this.pautas as Pauta[]).push(saved);
        return saved;
    }
}