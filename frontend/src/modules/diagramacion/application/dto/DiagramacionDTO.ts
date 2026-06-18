export interface DiagramacionDTO {
  metadata: {
    version: string;
    unit: "mm";
    coordinateSystem: "grid";
    origin: "top-left";
  };
  edition: {
    id: number;
    no_paginas: number;
    ancho_mm: number;
    alto_mm: number;
    cuadros_ancho: number;
    cuadros_alto: number;
    facing_pages: boolean;
    margen_superior_mm: number;
    margen_inferior_mm: number;
    margen_izquierdo_mm: number;
    margen_derecho_mm: number;
  };
  pages: Array<{
    id: number;
    no_pagina: number;
    pautas: Array<{
      id: number;
      descripcion_pauta: string;
      cuadros_alto: number;
      cuadros_ancho: number;
      ubicacion_cuadros_x: number;
      ubicacion_cuadros_y: number;
      indesignBounds: {
        topMm: number;
        leftMm: number;
        bottomMm: number;
        rightMm: number;
      };
    }>;
  }>;
}
