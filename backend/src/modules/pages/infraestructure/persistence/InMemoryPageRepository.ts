import { Page } from "../../domain/entities/Page";
import { PageRepository } from "../../domain/repositories/PageRepository";


export class InMemoryPageRepository implements PageRepository {
  private pages: Page[] = [
    new Page(1,1,1),
    new Page(2,1,1)
  ];

  async findByEdicionId(edicionId: number): Promise<Page[]> {
    return this.pages.filter((page) => page.edicionId === edicionId);
  }

  async saveMany(pages: Page[]): Promise<Page[]> {
    const saved = pages.map((p, i) => new Page(this.pages.length + i + 1, p.no_pagina, p.edicionId));
    this.pages.push(...saved);
    return saved;
  }
}