import { DomainError } from "../../../../shared/domain/DomainError";
import { Edition } from "../../../editions/domain/entities/Edition";
import { Pauta } from "../../../pautas/domain/entities/Pauta";


export class LayoutValidator {
    validatePautaInsideGrid(edition: Edition, pauta: Pauta): void{
        if(pauta.ubicacion_cuadros_x < 0 || pauta.ubicacion_cuadros_y < 0) {
            throw new DomainError("La ubicación de la pauta no puede ser negativa.");
        }

        if(pauta.cuadros_alto <= 0 || pauta.cuadros_ancho <= 0) {
            throw new DomainError("La pauta debe tener un tamaño positivo.");
        }

        if(pauta.cuadros_ancho > edition.cuadros_ancho || pauta.cuadros_alto > edition.cuadros_alto) {
            throw new DomainError("La pauta no puede ser más grande que la edición.");
        }

        if(pauta.ubicacion_cuadros_x + pauta.cuadros_ancho > edition.cuadros_ancho || pauta.ubicacion_cuadros_y + pauta.cuadros_alto > edition.cuadros_alto) {
            throw new DomainError("La pauta no puede exceder los límites de la edición.");
        }
        
    }

    validateNoOverlap(pautas: Pauta[]): void{
        for(let i = 0; i < pautas.length; i++) {
            for(let j = i + 1; j < pautas.length; j++) {
                const pautaA = pautas[i];
                const pautaB = pautas[j];

                const overlaps = 
                    pautaA.ubicacion_cuadros_x < pautaB.ubicacion_cuadros_x + pautaB.cuadros_ancho &&
                    pautaA.ubicacion_cuadros_x + pautaA.cuadros_ancho > pautaB.ubicacion_cuadros_x &&
                    pautaA.ubicacion_cuadros_y < pautaB.ubicacion_cuadros_y + pautaB.cuadros_alto &&
                    pautaA.ubicacion_cuadros_y + pautaA.cuadros_alto > pautaB.ubicacion_cuadros_y;

                if(overlaps) {
                    throw new DomainError(`Las pautas con ID ${pautaA.id} y ${pautaB.id} se superponen.`);
                }
            }
        }
    }
}