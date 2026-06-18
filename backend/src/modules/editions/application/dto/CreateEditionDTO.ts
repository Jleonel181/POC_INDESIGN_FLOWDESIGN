export interface CreateEditionDTO {
    no_paginas: number;
    ancho_mm: number;
    alto_mm: number;
    cuadros_ancho: number;
    cuadros_alto: number;
    facing_pages?: boolean;
    margen_superior_mm?: number;
    margen_inferior_mm?: number;
    margen_izquierdo_mm?: number;
    margen_derecho_mm?: number;
}
