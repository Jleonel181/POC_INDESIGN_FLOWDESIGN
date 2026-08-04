import { UseCase } from "../../../../shared/application/UseCase";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { CreatePautaDTO } from "../dto/CreatePautaDTO";
import { DomainError } from "../../../../shared/domain/DomainError";

/**
 * Crea una pauta en la biblioteca (sin asignar a ninguna página).
 * Solo requiere nombre y dimensiones en cuadros.
 */
export class CreatePautaUseCase implements UseCase<CreatePautaDTO, Pauta> {
    constructor(private readonly pautaRepository: PautaRepository) {}

    async execute(input: CreatePautaDTO): Promise<Pauta> {
        if (!input.descripcion_pauta || input.descripcion_pauta.trim() === "") {
            throw new DomainError("descripcion_pauta es obligatorio.");
        }
        if (input.cuadros_alto <= 0) {
            throw new DomainError("cuadros_alto debe ser mayor a 0.");
        }
        if (input.cuadros_ancho <= 0) {
            throw new DomainError("cuadros_ancho debe ser mayor a 0.");
        }

        const pauta = new Pauta(
            0,
            input.descripcion_pauta.trim(),
            input.cuadros_alto,
            input.cuadros_ancho,
            null, // sin posición
            null, // sin posición
            null  // sin página
        );

        return this.pautaRepository.save(pauta);
    }
}
