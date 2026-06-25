//Actualmente este adapter no se utiliza, ya que la fuente de datos es Mysql, pero se deja por si en un futuro se quiere cambiar a un archivo plano

import * as fs from "fs";
import * as path from "path";
import { AdImport } from "../../domain/AdImport";
import { AdDataSourcePort } from "../../application/ports/AdDataSourcePort";
import { AdImportTransformer } from "../transformers/AdImportTransformer";

export class FileAdDataSourceAdapter implements AdDataSourcePort {
    constructor(private readonly filePath: string) {}

    async fetchByDate(date: string): Promise<AdImport[]> {
        const absolutePath = path.resolve(this.filePath);

        if (!fs.existsSync(absolutePath)) {
            throw new Error(`ESB data source file not found: ${absolutePath}`);
        }

        const raw = JSON.parse(fs.readFileSync(absolutePath, "utf-8"));
        const records: unknown[] = Object.values(raw)[0] as unknown[];

        return records
            .filter((r: any) => r.CoverDate?.startsWith(date))
            .map((r: any) => AdImportTransformer.transform(r));
    }
}
