import { UseCase } from "../../../../shared/application/UseCase";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { AssignPautaDTO } from "../dto/AssignPautaDTO";
import { DomainError } from "../../../../shared/domain/DomainError";

/**
 * Asigna una pauta de la biblioteca a una página en una posición específica.
 */
export class AssignPautaUseCase implements UseCase<AssignPautaDTO, Pauta> {
    constructor(private readonly pautaRepository: PautaRepository) {}

    async execute(input: AssignPautaDTO): Promise<Pauta> {
        if (input.ubicacion_cuadros_x < 0) {
            throw new DomainError("ubicacion_cuadros_x no puede ser negativo.");
        }
        if (input.ubicacion_cuadros_y < 0) {
            throw new DomainError("ubicacion_cuadros_y no puede ser negativo.");
        }
        if (!input.pagina_id || input.pagina_id <= 0) {
            throw new DomainError("pagina_id es obligatorio.");
        }

        const existing = await this.pautaRepository.findById(input.pautaId);
        if (!existing) {
            throw new DomainError(`La pauta con id ${input.pautaId} no existe.`);
        }

        // Verificar que no se solape con otras pautas en la misma página.
        const pagePautas = await this.pautaRepository.findByPageId(input.pagina_id);
        for (const p of pagePautas) {
            if (p.id === input.pautaId) continue;
            const px = p.ubicacion_cuadros_x ?? 0;
            const py = p.ubicacion_cuadros_y ?? 0;

            const overlapX = input.ubicacion_cuadros_x < px + p.cuadros_ancho
                && input.ubicacion_cuadros_x + existing.cuadros_ancho > px;
            const overlapY = input.ubicacion_cuadros_y < py + p.cuadros_alto
                && input.ubicacion_cuadros_y + existing.cuadros_alto > py;

            if (overlapX && overlapY) {
                throw new DomainError(`Se solapa con "${p.descripcion_pauta}".`);
            }
        }

        const assigned = new Pauta(
            existing.id,
            existing.descripcion_pauta,
            existing.cuadros_alto,
            existing.cuadros_ancho,
            input.ubicacion_cuadros_x,
            input.ubicacion_cuadros_y,
            input.pagina_id
        );

        return this.pautaRepository.save(assigned);
    }
}
