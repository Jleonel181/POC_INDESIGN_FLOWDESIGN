import { Repository } from "typeorm";
import { Edition } from "../../domain/entities/Edition";
import { EditionRepository } from "../../domain/repositories/EditionRepository";
import { EditionEntity } from "./entities/EditionEntity";
import { EditionMapper } from "./mappers/EditionMapper";

export class PostgresEditionRepository implements EditionRepository {
    constructor(private readonly repository: Repository<EditionEntity>) {}

    async findById(id: number): Promise<Edition | null> {
        try {
            const entity = await this.repository.findOne({
                where: { id }
            });

            if (!entity) {
                return null;
            }

            return EditionMapper.toDomain(entity);
        } catch (error) {
            console.error(`Error finding edition by id ${id}:`, error);
            throw new Error(`Failed to find edition with id ${id}`);
        }
    }

    async findAll(): Promise<Edition[]> {
        const entities = await this.repository.find();
        return entities.map(EditionMapper.toDomain);
    }

    async save(edition: Edition): Promise<Edition> {
        const entity = EditionMapper.toEntity(edition);
        const saved = await this.repository.save(entity);
        return EditionMapper.toDomain(saved);
    }
}
