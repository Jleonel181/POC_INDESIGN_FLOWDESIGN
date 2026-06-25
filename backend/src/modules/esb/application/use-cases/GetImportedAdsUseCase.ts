import { AdImport } from "../../domain/AdImport";
import { AdDataSourcePort } from "../ports/AdDataSourcePort";

export class GetImportedAdsUseCase {
    constructor(private readonly dataSource: AdDataSourcePort) {}

    async execute(date: string): Promise<AdImport[]> {
        if (!date || !/^\d{4}-\d{2}-\d{2}$/.test(date)) {
            throw new Error("Invalid date format. Expected YYYY-MM-DD");
        }
        return this.dataSource.fetchByDate(date);
    }
}
