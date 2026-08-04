import { AssignPautaUseCase } from "../AssignPautaUseCase";
import { InMemoryPautaRepository } from "../../../infraestructure/persistence/InMemoryPautaRepository";
import { Pauta } from "../../../domain/entities/Pauta";

describe("AssignPautaUseCase", () => {
  let useCase: AssignPautaUseCase;
  let repo: InMemoryPautaRepository;

  beforeEach(async () => {
    repo = new InMemoryPautaRepository();
    useCase = new AssignPautaUseCase(repo);

    // Crear pautas en la biblioteca
    await repo.save(new Pauta(0, "Nota", 3, 3, null, null, null));
    await repo.save(new Pauta(0, "Publicidad", 2, 5, null, null, null));
  });

  it("asigna una pauta a una página con posición", async () => {
    const result = await useCase.execute({
      pautaId: 1,
      pagina_id: 10,
      ubicacion_cuadros_x: 0,
      ubicacion_cuadros_y: 0,
    });

    expect(result.paginaId).toBe(10);
    expect(result.ubicacion_cuadros_x).toBe(0);
    expect(result.ubicacion_cuadros_y).toBe(0);
    expect(result.descripcion_pauta).toBe("Nota");
  });

  it("falla si la pauta no existe", async () => {
    await expect(useCase.execute({
      pautaId: 999,
      pagina_id: 10,
      ubicacion_cuadros_x: 0,
      ubicacion_cuadros_y: 0,
    })).rejects.toThrow(/no existe/);
  });

  it("falla si ubicacion_cuadros_x es negativo", async () => {
    await expect(useCase.execute({
      pautaId: 1,
      pagina_id: 10,
      ubicacion_cuadros_x: -1,
      ubicacion_cuadros_y: 0,
    })).rejects.toThrow(/negativo/);
  });

  it("falla si pagina_id es 0", async () => {
    await expect(useCase.execute({
      pautaId: 1,
      pagina_id: 0,
      ubicacion_cuadros_x: 0,
      ubicacion_cuadros_y: 0,
    })).rejects.toThrow(/pagina_id/);
  });

  it("falla si se solapa con otra pauta en la misma página", async () => {
    // Asignar primera pauta
    await useCase.execute({
      pautaId: 1,
      pagina_id: 10,
      ubicacion_cuadros_x: 0,
      ubicacion_cuadros_y: 0,
    });

    // Intentar asignar segunda pauta solapada (Publicidad: 5x2 en pos 2,0 se solapa con Nota: 3x3 en pos 0,0)
    await expect(useCase.execute({
      pautaId: 2,
      pagina_id: 10,
      ubicacion_cuadros_x: 2,
      ubicacion_cuadros_y: 0,
    })).rejects.toThrow(/solapa/);
  });

  it("no falla si las pautas no se solapan", async () => {
    await useCase.execute({
      pautaId: 1,
      pagina_id: 10,
      ubicacion_cuadros_x: 0,
      ubicacion_cuadros_y: 0,
    });

    // Publicidad: 5x2 en pos 0,3 — debajo de Nota: 3x3 en pos 0,0
    const result = await useCase.execute({
      pautaId: 2,
      pagina_id: 10,
      ubicacion_cuadros_x: 0,
      ubicacion_cuadros_y: 3,
    });

    expect(result.paginaId).toBe(10);
  });
});
