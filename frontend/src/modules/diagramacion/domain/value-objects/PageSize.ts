export class PageSize {
  constructor(
    public readonly widthMm: number,
    public readonly heightMm: number
  ) {
    if (widthMm <= 0) throw new Error('El ancho de página debe ser mayor a cero');
    if (heightMm <= 0) throw new Error('El alto de página debe ser mayor a cero');
  }
}
