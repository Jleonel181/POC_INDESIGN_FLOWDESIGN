import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";

export class InMemoryPautaRepository implements PautaRepository {
    private pautas: Pauta[] = [];

    async findById(id: number): Promise<Pauta | null> {
        return this.pautas.find(p => p.id === id) ?? null;
    }

    async findByPageId(pageId: number): Promise<Pauta[]> {
        return this.pautas.filter(p => p.paginaId === pageId);
    }

    async findAll(): Promise<Pauta[]> {
        return [...this.pautas];
    }

    async findByDate(date: string): Promise<Pauta[]> {
        return this.pautas.filter(p => p.cover_date === date);
    }

    async findUnassigned(): Promise<Pauta[]> {
        return this.pautas.filter(p => p.paginaId === null);
    }

    async save(pauta: Pauta): Promise<Pauta> {
        const id = this.pautas.length + 1;
        const saved = new Pauta(id, pauta.descripcion_pauta, pauta.cuadros_alto, pauta.cuadros_ancho, pauta.ubicacion_cuadros_x, pauta.ubicacion_cuadros_y, pauta.paginaId, pauta.content_type, pauta.image_base64, pauta.cover_date);
        this.pautas.push(saved);
        return saved;
    }

    async updateImage(pautaId: number, imageBase64: string): Promise<void> {
        const idx = this.pautas.findIndex(p => p.id === pautaId);
        if (idx >= 0) {
            const old = this.pautas[idx];
            this.pautas[idx] = new Pauta(old.id, old.descripcion_pauta, old.cuadros_alto, old.cuadros_ancho, old.ubicacion_cuadros_x, old.ubicacion_cuadros_y, old.paginaId, "image", imageBase64, old.cover_date);
        }
    }

    async updateImageByDescription(desc: string, imageBase64: string): Promise<void> {
        for (let i = 0; i < this.pautas.length; i++) {
            if (this.pautas[i].descripcion_pauta === desc) {
                const old = this.pautas[i];
                this.pautas[i] = new Pauta(old.id, old.descripcion_pauta, old.cuadros_alto, old.cuadros_ancho, old.ubicacion_cuadros_x, old.ubicacion_cuadros_y, old.paginaId, "image", imageBase64, old.cover_date);
            }
        }
    }

    async findByDescription(desc: string): Promise<Pauta[]> {
        return this.pautas.filter(p => p.descripcion_pauta === desc);
    }

    async removeImage(pautaId: number): Promise<void> {
        const idx = this.pautas.findIndex(p => p.id === pautaId);
        if (idx >= 0) {
            const old = this.pautas[idx];
            this.pautas[idx] = new Pauta(old.id, old.descripcion_pauta, old.cuadros_alto, old.cuadros_ancho, old.ubicacion_cuadros_x, old.ubicacion_cuadros_y, old.paginaId, "text", null, old.cover_date);
        }
    }
}
