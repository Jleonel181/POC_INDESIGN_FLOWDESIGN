import { Repository } from "typeorm";
import { Page } from "../../domain/entities/Page";
import { PageRepository } from "../../domain/repositories/PageRepository";
import { PageEntity } from "./entities/PageEntity";
import { PageMapper } from "./mappers/PageMapper";

export class PostgresPageRepository implements PageRepository {
    constructor(private readonly repository: Repository<PageEntity>) {}

    async findByEdicionId(edicionId: number): Promise<Page[]> {
        try {
            const entities = await this.repository.find({
                where: { edicion_id: edicionId },
                order: { no_pagina: 'ASC' }
            });

            return entities.map(entity => PageMapper.toDomain(entity));
        } catch (error) {
            console.error(`Error finding pages by edicion_id ${edicionId}:`, error);
            throw new Error(`Failed to find pages for edition ${edicionId}`);
        }
    }

    async saveMany(pages: Page[]): Promise<Page[]> {
        const entities = pages.map(PageMapper.toEntity);
        const saved = await this.repository.save(entities);
        return saved.map(PageMapper.toDomain);
    }
}
