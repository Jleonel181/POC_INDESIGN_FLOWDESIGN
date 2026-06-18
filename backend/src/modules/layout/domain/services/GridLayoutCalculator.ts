export interface GridLayoutInput {
  pageWidthMm: number;
  pageHeightMm: number;
  gridColumns: number;
  gridRows: number;
  gridX: number;
  gridY: number;
  gridWidth: number;
  gridHeight: number;
  facingPages?: boolean;
  pageNumber?: number;
  margenSuperiorMm?: number;
  margenInferiorMm?: number;
  margenIzquierdoMm?: number;
  margenDerechoMm?: number;
}

export interface GridLayoutOutput {
  xMm: number;
  yMm: number;
  widthMm: number;
  heightMm: number;
}

export class GridLayoutCalculator {
  calculate(input: GridLayoutInput): GridLayoutOutput {
    // Márgenes (por defecto 0)
    const margenSuperior = input.margenSuperiorMm || 0;
    const margenIzquierdo = input.margenIzquierdoMm || 0;
    const margenDerecho = input.margenDerechoMm || 0;
    const margenInferior = input.margenInferiorMm || 0;

    // Área de contenido (dentro de los márgenes)
    const contentWidthMm = input.pageWidthMm - margenIzquierdo - margenDerecho;
    const contentHeightMm = input.pageHeightMm - margenSuperior - margenInferior;

    // Tamaño de cada cuadro dentro del área de contenido
    const cellWidthMm = contentWidthMm / input.gridColumns;
    const cellHeightMm = contentHeightMm / input.gridRows;

    // Coordenadas relativas al área de contenido
    let xMm = this.round(input.gridX * cellWidthMm);
    let yMm = this.round(input.gridY * cellHeightMm);
    const widthMm = this.round(input.gridWidth * cellWidthMm);
    const heightMm = this.round(input.gridHeight * cellHeightMm);

    // Sumar offset de márgenes para obtener coordenadas absolutas
    xMm += margenIzquierdo;
    yMm += margenSuperior;

    // Ajuste para facing pages:
    // En InDesign con facing pages (spreads), las coordenadas son relativas al spread completo
    // - Página 1 (portada): sola, coordenadas 0 a pageWidth
    // - Páginas pares (2, 4, 6...): izquierda del spread, coordenadas 0 a pageWidth
    // - Páginas impares (3, 5, 7...): derecha del spread, coordenadas pageWidth a pageWidth*2
    if (input.facingPages && input.pageNumber) {
      if (input.pageNumber > 1) {
        const isRightPage = input.pageNumber % 2 !== 0; // Impar = derecha
        
        if (isRightPage) {
          // Página derecha: sumar el ancho de la página izquierda
          xMm = xMm + input.pageWidthMm;
        }
        // Página izquierda (par): coordenadas normales (0 en adelante)
      }
    }

    return {
      xMm,
      yMm,
      widthMm,
      heightMm
    };
  }

  private round(value: number): number {
    return Number(value.toFixed(2));
  }
}