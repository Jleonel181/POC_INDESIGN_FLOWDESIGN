import { CreateEditionUseCase } from "../CreateEditionUseCase";
import { Edition } from "../../../domain/entities/Edition";
import { Page } from "../../../../pages/domain/entities/Page";
import { EditionRepository } from "../../../domain/repositories/EditionRepository";
import { PageRepository } from "../../../../pages/domain/repositories/PageRepository";

// Mocks mínimos
class MockEditionRepository implements EditionRepository {
  private editions: Edition[] = [];

  async findById(id: number) { return this.editions.find(e => e.id === id) ?? null; }
  async findAll() { return this.editions; }
  async save(edition: Edition) {
    const saved = new Edition(this.editions.length + 1, edition.no_paginas, edition.ancho_mm, edition.alto_mm, edition.cuadros_ancho, edition.cuadros_alto, edition.facing_pages, edition.margen_superior_mm, edition.margen_inferior_mm, edition.margen_izquierdo_mm, edition.margen_derecho_mm);
    this.editions.push(saved);
    return saved;
  }
}

class MockPageRepository implements PageRepository {
  private pages: Page[] = [];

  async findByEdicionId(edicionId: number) { return this.pages.filter(p => p.edicionId === edicionId); }
  async saveMany(pages: Page[]) {
    const saved = pages.map((p, i) => new Page(this.pages.length + i + 1, p.no_pagina, p.edicionId));
    this.pages.push(...saved);
    return saved;
  }
}

describe("CreateEditionUseCase", () => {
  let useCase: CreateEditionUseCase;

  beforeEach(() => {
    useCase = new CreateEditionUseCase(new MockEditionRepository(), new MockPageRepository());
  });

  it("crea edición y genera las páginas automáticamente", async () => {
    const result = await useCase.execute({
      no_paginas: 4,
      ancho_mm: 265,
      alto_mm: 370,
      cuadros_ancho: 5,
      cuadros_alto: 8,
      facing_pages: false,
      margen_superior_mm: 10,
      margen_inferior_mm: 10,
      margen_izquierdo_mm: 10,
      margen_derecho_mm: 10,
    });

    expect(result.edition.id).toBe(1);
    expect(result.edition.no_paginas).toBe(4);
    expect(result.edition.ancho_mm).toBe(265);
    expect(result.pages).toHaveLength(4);
    expect(result.pages[0].no_pagina).toBe(1);
    expect(result.pages[3].no_pagina).toBe(4);
  });

  it("crea edición con facing pages", async () => {
    const result = await useCase.execute({
      no_paginas: 8,
      ancho_mm: 210,
      alto_mm: 297,
      cuadros_ancho: 4,
      cuadros_alto: 6,
      facing_pages: true,
      margen_superior_mm: 15,
      margen_inferior_mm: 15,
      margen_izquierdo_mm: 20,
      margen_derecho_mm: 15,
    });

    expect(result.edition.facing_pages).toBe(true);
    expect(result.pages).toHaveLength(8);
  });

  it("usa márgenes por defecto 0 si no se especifican", async () => {
    const result = await useCase.execute({
      no_paginas: 1,
      ancho_mm: 100,
      alto_mm: 100,
      cuadros_ancho: 2,
      cuadros_alto: 2,
    });

    expect(result.edition.margen_superior_mm).toBe(0);
    expect(result.edition.facing_pages).toBe(false);
  });
});
