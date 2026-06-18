export class GridSize {
  constructor(
    public readonly width: number,
    public readonly height: number
  ) {
    if (width <= 0) {
      throw new Error('El ancho debe ser mayor a cero');
    }

    if (height <= 0) {
      throw new Error('El alto debe ser mayor a cero');
    }
  }
}