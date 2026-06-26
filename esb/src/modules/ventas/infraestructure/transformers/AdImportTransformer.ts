import { AdImport } from "../../domain/AdImport";

const INCHES_TO_MM = 25.4;
const FIXED_GRID_ANCHO = 6;
const FIXED_GRID_ALTO = 8;

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

export class AdImportTransformer {
    static transform(raw: RawAdRecord): AdImport {
        return new AdImport(
            raw.FID,
            raw.Customer,
            raw.Product,
            raw.Slogan,
            raw.CoverDate,
            raw.FormatFID,
            FIXED_GRID_ANCHO,
            FIXED_GRID_ALTO,
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
