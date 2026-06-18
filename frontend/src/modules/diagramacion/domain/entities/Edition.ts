export class Edition {
  constructor(
    public readonly id: number,
    public readonly noPaginas: number,
    public readonly anchoMm: number,
    public readonly altoMm: number,
    public readonly cuadrosAncho: number,
    public readonly cuadrosAlto: number,
    public readonly facingPages: boolean = false,
    public readonly margenSuperiorMm: number = 0,
    public readonly margenInferiorMm: number = 0,
    public readonly margenIzquierdoMm: number = 0,
    public readonly margenDerechoMm: number = 0
  ) {}

  get gridColumns(): number {
    return this.cuadrosAncho;
  }

  get gridRows(): number {
    return this.cuadrosAlto;
  }
}
