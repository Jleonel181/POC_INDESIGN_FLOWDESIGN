import { DomainError } from "../../../../shared/domain/DomainError";
import { Edition } from "../../../editions/domain/entities/Edition";
import { Pauta } from "../../../pautas/domain/entities/Pauta";

export class LayoutValidator {
    validatePautaInsideGrid(edition: Edition, pauta: Pauta): void {
        const x = pauta.ubicacion_cuadros_x ?? 0;
        const y = pauta.ubicacion_cuadros_y ?? 0;

        if (x < 0 || y < 0) {
            throw new DomainError("La ubicación de la pauta no puede ser negativa.");
        }

        if (pauta.cuadros_alto <= 0 || pauta.cuadros_ancho <= 0) {
            throw new DomainError("La pauta debe tener un tamaño positivo.");
        }

        if (pauta.cuadros_ancho > edition.cuadros_ancho || pauta.cuadros_alto > edition.cuadros_alto) {
            throw new DomainError("La pauta no puede ser más grande que la edición.");
        }

        if (x + pauta.cuadros_ancho > edition.cuadros_ancho || y + pauta.cuadros_alto > edition.cuadros_alto) {
            throw new DomainError("La pauta no puede exceder los límites de la edición.");
        }
    }

    validateNoOverlap(pautas: Pauta[]): void {
        for (let i = 0; i < pautas.length; i++) {
            for (let j = i + 1; j < pautas.length; j++) {
                const a = pautas[i];
                const b = pautas[j];

                const ax = a.ubicacion_cuadros_x ?? 0;
                const ay = a.ubicacion_cuadros_y ?? 0;
                const bx = b.ubicacion_cuadros_x ?? 0;
                const by = b.ubicacion_cuadros_y ?? 0;

                const overlaps =
                    ax < bx + b.cuadros_ancho &&
                    ax + a.cuadros_ancho > bx &&
                    ay < by + b.cuadros_alto &&
                    ay + a.cuadros_alto > by;

                if (overlaps) {
                    throw new DomainError(`Las pautas "${a.descripcion_pauta}" y "${b.descripcion_pauta}" se superponen.`);
                }
            }
        }
    }
}
