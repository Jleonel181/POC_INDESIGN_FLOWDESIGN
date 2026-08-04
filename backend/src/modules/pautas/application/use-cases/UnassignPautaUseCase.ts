import { UseCase } from "../../../../shared/application/UseCase";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { DomainError } from "../../../../shared/domain/DomainError";

/**
 * Desvincula una pauta de su página asignada, devolviéndola a la biblioteca.
 */
export class UnassignPautaUseCase implements UseCase<number, Pauta> {
    constructor(private readonly pautaRepository: PautaRepository) {}

    async execute(pautaId: number): Promise<Pauta> {
        if (!pautaId || pautaId <= 0) {
            throw new DomainError("pautaId es obligatorio y debe ser mayor a 0.");
        }

        const existing = await this.pautaRepository.findById(pautaId);
        if (!existing) {
            throw new DomainError(`La pauta con id ${pautaId} no existe.`);
        }

        if (!existing.isAssigned) {
            throw new DomainError(`La pauta "${existing.descripcion_pauta}" no está asignada a ninguna página.`);
        }

        const unassigned = new Pauta(
            existing.id,
            existing.descripcion_pauta,
            existing.cuadros_alto,
            existing.cuadros_ancho,
            null,
            null,
            null
        );

        return this.pautaRepository.save(unassigned);
    }
}
