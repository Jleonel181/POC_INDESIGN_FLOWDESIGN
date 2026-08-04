import { GridLayoutCalculator } from "../GridLayoutCalculator";

describe("GridLayoutCalculator", () => {
  const calc = new GridLayoutCalculator();

  const baseInput = {
    pageWidthMm: 265,
    pageHeightMm: 370,
    gridColumns: 5,
    gridRows: 8,
    gridX: 0,
    gridY: 0,
    gridWidth: 1,
    gridHeight: 1,
    margenSuperiorMm: 10,
    margenInferiorMm: 10,
    margenIzquierdoMm: 10,
    margenDerechoMm: 10,
  };

  it("calcula bounds de una celda en (0,0) con márgenes", () => {
    const result = calc.calculate(baseInput);
    expect(result.xMm).toBe(10); // margen izquierdo
    expect(result.yMm).toBe(10); // margen superior
    expect(result.widthMm).toBe(49); // (265 - 20) / 5 = 49
    expect(result.heightMm).toBe(43.75); // (370 - 20) / 8 = 43.75
  });

  it("calcula bounds de una celda en (2,3) de tamaño 2x2", () => {
    const result = calc.calculate({
      ...baseInput,
      gridX: 2,
      gridY: 3,
      gridWidth: 2,
      gridHeight: 2,
    });
    expect(result.xMm).toBe(10 + 2 * 49); // 108
    expect(result.yMm).toBe(10 + 3 * 43.75); // 141.25
    expect(result.widthMm).toBe(98); // 2 * 49
    expect(result.heightMm).toBe(87.5); // 2 * 43.75
  });

  it("sin márgenes, la celda empieza en (0,0)", () => {
    const result = calc.calculate({
      ...baseInput,
      margenSuperiorMm: 0,
      margenInferiorMm: 0,
      margenIzquierdoMm: 0,
      margenDerechoMm: 0,
    });
    expect(result.xMm).toBe(0);
    expect(result.yMm).toBe(0);
    expect(result.widthMm).toBe(53); // 265 / 5
    expect(result.heightMm).toBe(46.25); // 370 / 8
  });

  it("facing pages: página derecha (impar > 1) suma pageWidth al X", () => {
    const result = calc.calculate({
      ...baseInput,
      facingPages: true,
      pageNumber: 3,
    });
    expect(result.xMm).toBe(10 + 265); // margen + pageWidth
  });

  it("facing pages: página izquierda (par) no suma offset", () => {
    const result = calc.calculate({
      ...baseInput,
      facingPages: true,
      pageNumber: 2,
    });
    expect(result.xMm).toBe(10); // solo margen
  });

  it("facing pages: página 1 (portada) no suma offset", () => {
    const result = calc.calculate({
      ...baseInput,
      facingPages: true,
      pageNumber: 1,
    });
    expect(result.xMm).toBe(10);
  });
});
