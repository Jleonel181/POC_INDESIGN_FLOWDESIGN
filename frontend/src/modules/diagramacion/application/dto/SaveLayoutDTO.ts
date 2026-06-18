export interface SaveLayoutItemDTO {
  pautaId: number;
  pageId: number;
  ubicacion_cuadros_x: number;
  ubicacion_cuadros_y: number;
  cuadros_ancho: number;
  cuadros_alto: number;
}

export interface SaveLayoutDTO {
  items: SaveLayoutItemDTO[];
}
