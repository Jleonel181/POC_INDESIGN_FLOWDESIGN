import { Edition } from "../../domain/entities/Edition";
import { EditionRepository } from "../../domain/repositories/EditionRepository";


export class InMemoryEditionRepository implements EditionRepository {
    private readonly editions: Edition[] = [
        new Edition(
            1,
            2,
            280,
            350,
            6,
            8
        )
    ];

    async findById(id: number): Promise<Edition | null> {
        const edition = this.editions.find(e => e.id === id);
        return edition || null;
    }

    async findAll(): Promise<Edition[]> {
        return this.editions;
    }

    async save(edition: Edition): Promise<Edition> {
        const newEdition = new Edition(
            this.editions.length + 1,
            edition.no_paginas,
            edition.ancho_mm,
            edition.alto_mm,
            edition.cuadros_ancho,
            edition.cuadros_alto,
            edition.facing_pages,
            edition.margen_superior_mm,
            edition.margen_inferior_mm,
            edition.margen_izquierdo_mm,
            edition.margen_derecho_mm
        );
        this.editions.push(newEdition);
        return newEdition;
    }
}
