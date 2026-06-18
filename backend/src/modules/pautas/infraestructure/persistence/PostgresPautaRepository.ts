import { Repository } from "typeorm";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaRepository } from "../../domain/repositories/PautaRepository";
import { PautaEntity } from "./entities/PautaEntity";
import { PautaMapper } from "./mappers/PautaMapper";

export class PostgresPautaRepository implements PautaRepository {
    constructor(private readonly repository: Repository<PautaEntity>) {}

    async findByPageId(pageId: number): Promise<Pauta[]> {
        try {
            const entities = await this.repository.find({
                where: { pagina_id: pageId },
                order: { id: 'ASC' }
            });

            return entities.map(entity => PautaMapper.toDomain(entity));
        } catch (error) {
            console.error(`Error finding pautas by pagina_id ${pageId}:`, error);
            throw new Error(`Failed to find pautas for page ${pageId}`);
        }
    }
}
