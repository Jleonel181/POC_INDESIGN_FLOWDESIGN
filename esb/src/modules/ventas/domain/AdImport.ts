export class AdImport {
    constructor(
        public readonly fid: string,
        public readonly customer: string,
        public readonly product: string,
        public readonly slogan: string,
        public readonly coverDate: string,
        public readonly formatFid: string,
        public readonly cuadrosAncho: number,
        public readonly cuadrosAlto: number,
        public readonly widthMm: number,
        public readonly heightMm: number,
        public readonly colorFid: string,
        public readonly placementComment: string,
        public readonly statusFid: string,
        public readonly printSystemFid: string,
        public readonly regionalName: string,
        public readonly productCategory: string,
        public readonly observations: string | null
    ) {}
}
