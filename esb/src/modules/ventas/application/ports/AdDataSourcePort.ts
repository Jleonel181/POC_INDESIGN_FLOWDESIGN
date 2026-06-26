import { AdImport } from "../../domain/AdImport";

export interface AdDataSourcePort {
    fetchByDate(date: string): Promise<AdImport[]>;
}
