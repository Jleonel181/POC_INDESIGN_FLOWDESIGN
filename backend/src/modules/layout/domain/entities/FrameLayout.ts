export class FrameLayout {

    constructor(
        public readonly pautaId: number,
        public readonly descripcion: string,
        public readonly cuadros_alto: number,
        public readonly cuadros_ancho: number,
        public readonly ubicacion_cuadros_x: number,
        public readonly ubicacion_cuadros_y: number
    ){}
}