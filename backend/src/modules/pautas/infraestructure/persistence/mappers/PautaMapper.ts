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
            entity.pagina_id,
            (entity.content_type as "text" | "image") || "text",
            entity.image_base64,
            entity.cover_date
        );
    }

    static toEntity(domain: Pauta): PautaEntity {
        const entity = new PautaEntity();
        if (domain.id > 0) entity.id = domain.id;
        entity.descripcion_pauta = domain.descripcion_pauta;
        entity.cuadros_alto = domain.cuadros_alto;
        entity.cuadros_ancho = domain.cuadros_ancho;
        entity.ubicacion_cuadros_x = domain.ubicacion_cuadros_x;
        entity.ubicacion_cuadros_y = domain.ubicacion_cuadros_y;
        entity.pagina_id = domain.paginaId;
        entity.content_type = domain.content_type;
        entity.image_base64 = domain.image_base64;
        entity.cover_date = domain.cover_date;
        return entity;
    }
}
