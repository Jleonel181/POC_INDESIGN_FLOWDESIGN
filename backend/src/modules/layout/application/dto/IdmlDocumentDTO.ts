/**
 * DTO que describe un documento IDML genérico.
 * Este es el contrato de comunicación con idmlgen.
 * No contiene conceptos del dominio editorial (pautas, cuadros, ediciones).
 */

export interface IdmlDocumentDTO {
  document: {
    widthMm: number;
    heightMm: number;
    margins: {
      top: number;
      bottom: number;
      left: number;
      right: number;
    };
    facingPages: boolean;
    columns: number;
    guides: IdmlGuideDTO[];
  };
  pages: IdmlPageDTO[];
  masterSpreadSource?: IdmlMasterSpreadSourceDTO;
}

/**
 * Indica de dónde extraer un MasterSpread existente para inyectarlo como cabecera.
 * idmlgen abre el IDML plantilla, busca el master por nombre, y lo copia al documento
 * generado. Todas las páginas lo referencian automáticamente.
 */
export interface IdmlMasterSpreadSourceDTO {
  /** Ruta absoluta al archivo .idml que contiene el master spread. */
  templatePath: string;
  /** Atributo Name del MasterSpread a extraer (ej: "02-Noticias Apertura"). */
  masterSpreadName: string;
}

export interface IdmlGuideDTO {
  orientation: "vertical" | "horizontal";
  locationMm: number;
}

export interface IdmlPageDTO {
  frames: IdmlFrameDTO[];
}

export interface IdmlFrameDTO {
  type: "text";
  name: string;
  bounds: {
    topMm: number;
    leftMm: number;
    bottomMm: number;
    rightMm: number;
  };
  content: string;
  options: {
    verticalJustification?: string;
    contentIsRaw?: boolean;
  };
}
