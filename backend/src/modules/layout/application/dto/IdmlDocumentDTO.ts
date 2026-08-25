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
  /** Fecha a inyectar en el encabezado (reemplaza {{fecha}} en las stories del master). */
  folioDate?: string;
}

export interface IdmlGuideDTO {
  orientation: "vertical" | "horizontal";
  locationMm: number;
}

export interface IdmlPageDTO {
  frames: IdmlFrameDTO[];
}

export type IdmlFrameDTO = IdmlTextFrameDTO | IdmlImageFrameDTO;

export interface IdmlTextFrameDTO {
  type: "text";
  name: string;
  bounds: IdmlBoundsDTO;
  content: string;
  options: {
    verticalJustification?: string;
    contentIsRaw?: boolean;
  };
}

export interface IdmlImageFrameDTO {
  type: "image";
  name: string;
  bounds: IdmlBoundsDTO;
  /** Imagen codificada en base64 estándar (RFC 4648). */
  imageBase64: string;
}

export interface IdmlBoundsDTO {
  topMm: number;
  leftMm: number;
  bottomMm: number;
  rightMm: number;
}
