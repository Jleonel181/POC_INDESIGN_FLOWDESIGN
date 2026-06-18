export class Edition {
    constructor(
        public id: number,
        public no_paginas: number,
        public ancho_mm: number,
        public alto_mm: number,
        public cuadros_ancho: number,
        public cuadros_alto: number,
        public facing_pages: boolean = false,
        public margen_superior_mm: number = 0,
        public margen_inferior_mm: number = 0,
        public margen_izquierdo_mm: number = 0,
        public margen_derecho_mm: number = 0
    ) {}
}