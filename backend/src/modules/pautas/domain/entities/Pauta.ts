export class Pauta {
    constructor(
        public readonly id: number,
        public readonly descripcion_pauta: string,
        public readonly cuadros_alto: number,
        public readonly cuadros_ancho: number,
        public readonly ubicacion_cuadros_x: number,
        public readonly ubicacion_cuadros_y: number,
        public readonly paginaId: number
    ){}
}