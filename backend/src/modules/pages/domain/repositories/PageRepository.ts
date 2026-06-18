import { Page } from '../entities/Page';

export interface PageRepository {
    findByEdicionId(edicionId: number): Promise<Page[]>;
    saveMany(pages: Page[]): Promise<Page[]>;
}