import { Repository, IsNull } from "typeorm";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { PautaEntity } from "./entities/PautaEntity";
import { PautaMapper } from "./mappers/PautaMapper";

export class PostgresPautaRepository implements PautaRepository {
    constructor(private readonly repository: Repository<PautaEntity>) {}

    async findById(id: number): Promise<Pauta | null> {
        const entity = await this.repository.findOne({ where: { id } });
        return entity ? PautaMapper.toDomain(entity) : null;
    }

    async findByPageId(pageId: number): Promise<Pauta[]> {
        const entities = await this.repository.find({
            where: { pagina_id: pageId },
            order: { id: "ASC" }
        });
        return entities.map(PautaMapper.toDomain);
    }

    async findAll(): Promise<Pauta[]> {
        const entities = await this.repository.find({ order: { id: "ASC" } });
        return entities.map(PautaMapper.toDomain);
    }

    async findByDate(date: string): Promise<Pauta[]> {
        const entities = await this.repository
            .createQueryBuilder("p")
            .where("p.cover_date = :date", { date })
            .orderBy("p.id", "ASC")
            .getMany();
        return entities.map(PautaMapper.toDomain);
    }

    async findUnassigned(): Promise<Pauta[]> {
        const entities = await this.repository.find({
            where: { pagina_id: IsNull() },
            order: { id: "ASC" }
        });
        return entities.map(PautaMapper.toDomain);
    }

    async save(pauta: Pauta): Promise<Pauta> {
        const entity = PautaMapper.toEntity(pauta);
        const saved = await this.repository.save(entity);
        return PautaMapper.toDomain(saved);
    }

    async updateImage(pautaId: number, imageBase64: string): Promise<void> {
        await this.repository.update(pautaId, {
            content_type: "image",
            image_base64: imageBase64,
        });
    }

    async updateImageByDescription(desc: string, imageBase64: string): Promise<void> {
        await this.repository
            .createQueryBuilder()
            .update()
            .set({ content_type: "image", image_base64: imageBase64 })
            .where("descripcion_pauta = :desc", { desc })
            .execute();
    }

    async findByDescription(desc: string): Promise<Pauta[]> {
        const entities = await this.repository.find({
            where: { descripcion_pauta: desc },
            order: { id: "ASC" },
        });
        return entities.map(PautaMapper.toDomain);
    }

    async removeImage(pautaId: number): Promise<void> {
        await this.repository.update(pautaId, {
            content_type: "text",
            image_base64: null,
        });
    }
}
