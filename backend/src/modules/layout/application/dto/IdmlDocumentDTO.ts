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
