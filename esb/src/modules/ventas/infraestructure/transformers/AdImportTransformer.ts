import { AdImport } from "../../domain/AdImport";

const INCHES_TO_MM = 25.4;

interface RawAdRecord {
    FID: string;
    Customer: string;
    Product: string;
    Slogan: string;
    CoverDate: string;
    FormatFID: string;
    Width: number;
    Height: number;
    ColorFID: string;
    PlacementComment: string;
    StatusFID: string;
    PrintSystemFID: string;
    regionalName: string;
    ProductCategoryFID: string;
    observations: string | null;
}

/**
 * Parsea FormatFID con formato "anchoXalto-grid" (ej: "3x4-1" o "2.5x3.5-1")
 * y extrae cuadros ancho/alto redondeados.
 * Fallback a 6x8 si el formato no es parseable.
 */
function parseGridFromFormat(formatFid: string): { ancho: number; alto: number } {
    // ponytail: regex acepta enteros y decimales; redondea al entero más cercano
    const match = formatFid?.match(/^([\d.]+)\s*x\s*([\d.]+)/i);
    if (match) {
        return {
            ancho: Math.round(parseFloat(match[1])),
            alto: Math.round(parseFloat(match[2]))
        };
    }
    return { ancho: 6, alto: 8 };
}

export class AdImportTransformer {
    static transform(raw: RawAdRecord): AdImport {
        const grid = parseGridFromFormat(raw.FormatFID);

        return new AdImport(
            raw.FID,
            raw.Customer,
            raw.Product,
            raw.Slogan,
            raw.CoverDate,
            raw.FormatFID,
            grid.ancho,
            grid.alto,
            parseFloat((raw.Width * INCHES_TO_MM).toFixed(2)),
            parseFloat((raw.Height * INCHES_TO_MM).toFixed(2)),
            raw.ColorFID,
            raw.PlacementComment,
            raw.StatusFID,
            raw.PrintSystemFID,
            raw.regionalName,
            raw.ProductCategoryFID,
            raw.observations ?? null
        );
    }
}
