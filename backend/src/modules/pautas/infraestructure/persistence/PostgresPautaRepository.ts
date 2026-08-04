import { Repository, IsNull } from "typeorm";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { PautaEntity } from "./entities/PautaEntity";
import { PautaMapper } from "./mappers/PautaMapper";

export class PostgresPautaRepository implements PautaRepository {
    constructor(private readonly repository: Repository<PautaEntity>) {}

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
}
