export class Pauta {
    constructor(
        public readonly id: number,
        public readonly descripcion_pauta: string,
        public readonly cuadros_alto: number,
        public readonly cuadros_ancho: number,
        public readonly ubicacion_cuadros_x: number | null,
        public readonly ubicacion_cuadros_y: number | null,
        public readonly paginaId: number | null,
        public readonly content_type: "text" | "image" = "text",
        public readonly image_base64: string | null = null,
        public readonly cover_date: string | null = null
    ) {}

    /** Indica si esta pauta ya fue asignada a una página. */
    get isAssigned(): boolean {
        return this.paginaId !== null && this.ubicacion_cuadros_x !== null && this.ubicacion_cuadros_y !== null;
    }
}
