import { GenerateIdmlUseCase } from "../GenerateIdmlUseCase";
import { GenerateEditionLayoutUseCase } from "../GenerateEditionLayoutUseCase";
import { IdmlGenerator } from "../../../domain/ports/IdmlGenerator";
import { IdmlDocumentDTO } from "../../dto/IdmlDocumentDTO";

// Mock del generador — solo captura lo que recibe
class MockIdmlGenerator implements IdmlGenerator {
  lastDocument: IdmlDocumentDTO | null = null;

  async generate(document: IdmlDocumentDTO): Promise<Buffer> {
    this.lastDocument = document;
    return Buffer.from("fake-idml");
  }
}

// Mock del use case de layout
const mockLayout = {
  metadata: { version: "1.0.0", unit: "mm" as const, coordinateSystem: "grid" as const, origin: "top-left" as const },
  edition: {
    id: 1, no_paginas: 2, ancho_mm: 265, alto_mm: 370,
    cuadros_ancho: 5, cuadros_alto: 8, facing_pages: true,
    margen_superior_mm: 10, margen_inferior_mm: 10,
    margen_izquierdo_mm: 10, margen_derecho_mm: 10,
  },
  pages: [
    { id: 1, no_pagina: 1, pautas: [
      { id: 1, descripcion_pauta: "Cabezal", cuadros_alto: 1, cuadros_ancho: 5, ubicacion_cuadros_x: 0, ubicacion_cuadros_y: 0, indesignBounds: { topMm: 10, leftMm: 10, bottomMm: 53.75, rightMm: 255 } },
    ]},
    { id: 2, no_pagina: 2, pautas: [
      { id: 2, descripcion_pauta: "Nota", cuadros_alto: 3, cuadros_ancho: 3, ubicacion_cuadros_x: 0, ubicacion_cuadros_y: 0, indesignBounds: { topMm: 10, leftMm: 10, bottomMm: 141.25, rightMm: 157 } },
    ]},
    { id: 3, no_pagina: 3, pautas: [
      { id: 3, descripcion_pauta: "Deporte", cuadros_alto: 2, cuadros_ancho: 2, ubicacion_cuadros_x: 0, ubicacion_cuadros_y: 0, indesignBounds: { topMm: 10, leftMm: 275, bottomMm: 97.5, rightMm: 373 } },
    ]},
  ],
};

const mockGenerateLayout = {
  execute: jest.fn().mockResolvedValue(mockLayout),
} as unknown as GenerateEditionLayoutUseCase;

describe("GenerateIdmlUseCase", () => {
  let useCase: GenerateIdmlUseCase;
  let mockGenerator: MockIdmlGenerator;

  beforeEach(() => {
    mockGenerator = new MockIdmlGenerator();
    useCase = new GenerateIdmlUseCase(mockGenerateLayout, mockGenerator);
  });

  it("produce un IdmlDocumentDTO con las dimensiones correctas", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    const doc = mockGenerator.lastDocument!;
    expect(doc.document.widthMm).toBe(265);
    expect(doc.document.heightMm).toBe(370);
    expect(doc.document.facingPages).toBe(true);
    expect(doc.document.columns).toBe(5);
  });

  it("genera una página por cada page del layout", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    const doc = mockGenerator.lastDocument!;
    expect(doc.pages).toHaveLength(3);
  });

  it("traduce pautas a frames con bounds", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    const frame = mockGenerator.lastDocument!.pages[0].frames[0];
    expect(frame.type).toBe("text");
    expect(frame.name).toBe("Cabezal");
    expect(frame.bounds.topMm).toBe(10);
    expect(frame.bounds.leftMm).toBe(10);
  });

  it("resta el offset de facing pages en páginas derechas (impares > 1)", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    // Página 3 es derecha, sus bounds originales tienen leftMm=275 (offset de 265 sumado)
    const frame = mockGenerator.lastDocument!.pages[2].frames[0];
    expect(frame.bounds.leftMm).toBe(275 - 265); // 10
    expect(frame.bounds.rightMm).toBe(373 - 265); // 108
  });

  it("no resta offset en páginas izquierdas (pares)", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    const frame = mockGenerator.lastDocument!.pages[1].frames[0];
    expect(frame.bounds.leftMm).toBe(10); // sin cambio
  });

  it("agrega folio a cada página cuando folio=true", async () => {
    await useCase.execute({ editionId: 1, folio: true });

    const doc = mockGenerator.lastDocument!;
    for (const page of doc.pages) {
      const folioFrame = page.frames.find(f => f.name === "folio");
      expect(folioFrame).toBeDefined();
      expect(folioFrame!.content).toBe("<?ACE 18?>");
      expect(folioFrame!.options.contentIsRaw).toBe(true);
      expect(folioFrame!.options.verticalJustification).toBe("BottomAlign");
    }
  });

  it("no agrega folio cuando folio=false", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    const doc = mockGenerator.lastDocument!;
    for (const page of doc.pages) {
      const folioFrame = page.frames.find(f => f.name === "folio");
      expect(folioFrame).toBeUndefined();
    }
  });

  it("genera guías de grilla", async () => {
    await useCase.execute({ editionId: 1, folio: false });

    const guides = mockGenerator.lastDocument!.document.guides;
    // 5 columnas → 4 guías verticales, 8 filas → 7 guías horizontales = 11 total
    expect(guides).toHaveLength(11);
    expect(guides.filter(g => g.orientation === "vertical")).toHaveLength(4);
    expect(guides.filter(g => g.orientation === "horizontal")).toHaveLength(7);
  });
});
