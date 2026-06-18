import { Pauta } from "../../../domain/entities/Pauta";
import { PautaEntity } from "../entities/PautaEntity";

export class PautaMapper {
    static toDomain(entity: PautaEntity): Pauta {
        return new Pauta(
            entity.id,
            entity.descripcion_pauta,
            entity.cuadros_alto,
            entity.cuadros_ancho,
            entity.ubicacion_cuadros_x,
            entity.ubicacion_cuadros_y,
            entity.pagina_id
        );
    }

    static toEntity(domain: Pauta): PautaEntity {
        const entity = new PautaEntity();
        entity.id = domain.id;
        entity.descripcion_pauta = domain.descripcion_pauta;
        entity.cuadros_alto = domain.cuadros_alto;
        entity.cuadros_ancho = domain.cuadros_ancho;
        entity.ubicacion_cuadros_x = domain.ubicacion_cuadros_x;
        entity.ubicacion_cuadros_y = domain.ubicacion_cuadros_y;
        entity.pagina_id = domain.paginaId;
        return entity;
    }
}
