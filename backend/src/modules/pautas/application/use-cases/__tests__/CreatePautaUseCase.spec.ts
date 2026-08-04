import { CreatePautaUseCase } from "../CreatePautaUseCase";
import { InMemoryPautaRepository } from "../../../infraestructure/persistence/InMemoryPautaRepository";

describe("CreatePautaUseCase", () => {
  let useCase: CreatePautaUseCase;
  let repo: InMemoryPautaRepository;

  beforeEach(() => {
    repo = new InMemoryPautaRepository();
    useCase = new CreatePautaUseCase(repo);
  });

  it("crea una pauta sin asignar (biblioteca)", async () => {
    const result = await useCase.execute({
      descripcion_pauta: "Nota principal",
      cuadros_alto: 4,
      cuadros_ancho: 3,
    });

    expect(result.id).toBeGreaterThan(0);
    expect(result.descripcion_pauta).toBe("Nota principal");
    expect(result.cuadros_alto).toBe(4);
    expect(result.cuadros_ancho).toBe(3);
    expect(result.paginaId).toBeNull();
    expect(result.ubicacion_cuadros_x).toBeNull();
    expect(result.ubicacion_cuadros_y).toBeNull();
  });

  it("falla si descripcion_pauta está vacío", async () => {
    await expect(useCase.execute({
      descripcion_pauta: "",
      cuadros_alto: 2,
      cuadros_ancho: 2,
    })).rejects.toThrow(/descripcion_pauta/);
  });

  it("falla si cuadros_alto es 0", async () => {
    await expect(useCase.execute({
      descripcion_pauta: "Test",
      cuadros_alto: 0,
      cuadros_ancho: 2,
    })).rejects.toThrow(/cuadros_alto/);
  });

  it("falla si cuadros_ancho es negativo", async () => {
    await expect(useCase.execute({
      descripcion_pauta: "Test",
      cuadros_alto: 2,
      cuadros_ancho: -1,
    })).rejects.toThrow(/cuadros_ancho/);
  });
});
