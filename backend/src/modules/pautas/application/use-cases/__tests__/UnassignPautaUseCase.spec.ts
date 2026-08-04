import { UnassignPautaUseCase } from "../UnassignPautaUseCase";
import { InMemoryPautaRepository } from "../../../infraestructure/persistence/InMemoryPautaRepository";
import { Pauta } from "../../../domain/entities/Pauta";

describe("UnassignPautaUseCase", () => {
  let useCase: UnassignPautaUseCase;
  let repo: InMemoryPautaRepository;

  beforeEach(async () => {
    repo = new InMemoryPautaRepository();
    useCase = new UnassignPautaUseCase(repo);

    // Pauta asignada
    await repo.save(new Pauta(0, "Nota", 3, 3, 0, 0, 10));
    // Pauta sin asignar
    await repo.save(new Pauta(0, "Libre", 2, 2, null, null, null));
  });

  it("desvincula una pauta asignada", async () => {
    const result = await useCase.execute(1);

    expect(result.paginaId).toBeNull();
    expect(result.ubicacion_cuadros_x).toBeNull();
    expect(result.ubicacion_cuadros_y).toBeNull();
    expect(result.descripcion_pauta).toBe("Nota");
  });

  it("falla si la pauta no existe", async () => {
    await expect(useCase.execute(999)).rejects.toThrow(/no existe/);
  });

  it("falla si la pauta ya está sin asignar", async () => {
    await expect(useCase.execute(2)).rejects.toThrow(/no está asignada/);
  });

  it("falla si pautaId es 0", async () => {
    await expect(useCase.execute(0)).rejects.toThrow(/pautaId/);
  });
});
