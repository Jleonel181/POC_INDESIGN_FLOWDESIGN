import mysql from "mysql2/promise";
import { AdImport } from "../../domain/AdImport";
import { AdDataSourcePort } from "../../application/ports/AdDataSourcePort";
import { AdImportTransformer } from "../transformers/AdImportTransformer";
import { VentasDbConfig } from "../../../../config/environment.config";

const SQL = `
    SELECT ID                                        FID
         , regionalName2                             GeoFID
         , customer                                  Customer
         , product                                   Product
         , regionalType                              ProductCategoryFID
         , positionName                              PlacementComment
         , publication_date                          CoverDate
         , CONCAT(width, 'x', height, '-', grid)     FormatFID
         , inches_width                              Width
         , inches_high                               Height
         , printColorDescription                     ColorFID
         , version_name                              Slogan
         , meanPaymentName                           AccountFID
         , Username                                  Creator
         , positionName                              Note
         , subtotal                                  AdRate
         , state                                     StatusFID
         , orderNumber                               PrintSystemFID
         , regionalName
         , printColorDescription
         , position
         , regionalName2
         , regionalType
         , positionName
         , rate
         , observations
    FROM vw_pauta_dataplan
    WHERE publication_date = ?
`;

export class MysqlAdDataSourceAdapter implements AdDataSourcePort {
    constructor(private readonly config: VentasDbConfig) {}

    async fetchByDate(date: string): Promise<AdImport[]> {
        const connection = await mysql.createConnection({
            host: this.config.host,
            port: this.config.port,
            user: this.config.username,
            password: this.config.password,
            database: this.config.database,
            charset: "LATIN1",
            connectTimeout: 10000,
        });

        try {
            const [rows] = await connection.execute(SQL, [date]);
            return (rows as any[]).map(r => AdImportTransformer.transform(r));
        } finally {
            await connection.end();
        }
    }
}
