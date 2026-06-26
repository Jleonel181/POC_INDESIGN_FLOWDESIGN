export interface AdImportDto {
    fid: string;
    customer: string;
    product: string;
    slogan: string;
    coverDate: string;
    formatFid: string;
    cuadrosAncho: number;
    cuadrosAlto: number;
    widthMm: number;
    heightMm: number;
    colorFid: string;
    placementComment: string;
    statusFid: string;
    printSystemFid: string;
    regionalName: string;
    productCategory: string;
    observations: string | null;
}

export class HttpEsbAdapter {
    constructor(private readonly esbBaseUrl: string) {}

    async getAdsByDate(date: string): Promise<AdImportDto[]> {
        const url = `${this.esbBaseUrl}/api/ventas/ads?date=${date}`;
        const response = await fetch(url);

        if (!response.ok) {
            throw new Error(`ESB responded with status ${response.status} for date ${date}`);
        }

        const body = await response.json() as { ads: AdImportDto[] };
        return body.ads;
    }
}
