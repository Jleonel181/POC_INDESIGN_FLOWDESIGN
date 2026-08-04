import { UseCase } from "../../../../shared/application/UseCase";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { PageRepository } from "../../../pages/domain/repositories/PageRepository";
import { EditionRepository } from "../../../editions/domain/repositories/EditionRepository";
import { CreatePautaDTO } from "../dto/CreatePautaDTO";
import { DomainError } from "../../../../shared/domain/DomainError";

export class CreatePautaUseCase implements UseCase<CreatePautaDTO, Pauta> {
    constructor(
        private readonly pautaRepository: PautaRepository,
        private readonly pageRepository: PageRepository,
        private readonly editionRepository: EditionRepository
    ) {}

    async execute(input: CreatePautaDTO): Promise<Pauta> {
        this.validateInput(input);

        // Verificar que la página existe y obtener la edición.
        const pages = await this.pageRepository.findByEdicionId(0);
        // Necesitamos buscar la página por ID para verificar que existe.
        // Buscamos todas las páginas de todas las ediciones y filtramos.
        // Mejor: buscamos la edición que contiene esta página.
        const edition = await this.findEditionForPage(input.pagina_id);

        if (!edition) {
            throw new DomainError(`La página con id ${input.pagina_id} no existe.`);
        }

        // Validar que la pauta cabe en la grilla de la edición.
        this.validateFitsInGrid(input, edition.cuadros_ancho, edition.cuadros_alto);

        // Validar que no se solapa con pautas existentes en la misma página.
        const existingPautas = await this.pautaRepository.findByPageId(input.pagina_id);
        this.validateNoOverlap(input, existingPautas);

        const pauta = new Pauta(
            0,
            input.descripcion_pauta,
            input.cuadros_alto,
            input.cuadros_ancho,
            input.ubicacion_cuadros_x,
            input.ubicacion_cuadros_y,
            input.pagina_id
        );

        return this.pautaRepository.save(pauta);
    }

    private validateInput(input: CreatePautaDTO): void {
        if (!input.descripcion_pauta || input.descripcion_pauta.trim() === "") {
            throw new DomainError("descripcion_pauta es obligatorio.");
        }
        if (input.cuadros_alto <= 0) {
            throw new DomainError("cuadros_alto debe ser mayor a 0.");
        }
        if (input.cuadros_ancho <= 0) {
            throw new DomainError("cuadros_ancho debe ser mayor a 0.");
        }
        if (input.ubicacion_cuadros_x < 0) {
            throw new DomainError("ubicacion_cuadros_x no puede ser negativo.");
        }
        if (input.ubicacion_cuadros_y < 0) {
            throw new DomainError("ubicacion_cuadros_y no puede ser negativo.");
        }
        if (!input.pagina_id || input.pagina_id <= 0) {
            throw new DomainError("pagina_id es obligatorio y debe ser mayor a 0.");
        }
    }

    private validateFitsInGrid(input: CreatePautaDTO, gridColumns: number, gridRows: number): void {
        if (input.ubicacion_cuadros_x + input.cuadros_ancho > gridColumns) {
            throw new DomainError(
                `La pauta excede la grilla horizontalmente: posición ${input.ubicacion_cuadros_x} + ancho ${input.cuadros_ancho} > ${gridColumns} columnas.`
            );
        }
        if (input.ubicacion_cuadros_y + input.cuadros_alto > gridRows) {
            throw new DomainError(
                `La pauta excede la grilla verticalmente: posición ${input.ubicacion_cuadros_y} + alto ${input.cuadros_alto} > ${gridRows} filas.`
            );
        }
    }

    private validateNoOverlap(input: CreatePautaDTO, existing: Pauta[]): void {
        for (const pauta of existing) {
            const overlapX = input.ubicacion_cuadros_x < pauta.ubicacion_cuadros_x + pauta.cuadros_ancho
                && input.ubicacion_cuadros_x + input.cuadros_ancho > pauta.ubicacion_cuadros_x;
            const overlapY = input.ubicacion_cuadros_y < pauta.ubicacion_cuadros_y + pauta.cuadros_alto
                && input.ubicacion_cuadros_y + input.cuadros_alto > pauta.ubicacion_cuadros_y;

            if (overlapX && overlapY) {
                throw new DomainError(
                    `La pauta se solapa con "${pauta.descripcion_pauta}" (id: ${pauta.id}).`
                );
            }
        }
    }

    private async findEditionForPage(pageId: number): Promise<{ cuadros_ancho: number; cuadros_alto: number } | null> {
        // Buscar en todas las ediciones cuál tiene una página con este ID.
        const editions = await this.editionRepository.findAll();
        for (const edition of editions) {
            const pages = await this.pageRepository.findByEdicionId(edition.id);
            if (pages.some(p => p.id === pageId)) {
                return { cuadros_ancho: edition.cuadros_ancho, cuadros_alto: edition.cuadros_alto };
            }
        }
        return null;
    }
}
