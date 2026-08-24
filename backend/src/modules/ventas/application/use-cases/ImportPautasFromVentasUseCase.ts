import { UseCase } from "../../../../shared/application/UseCase";
import { Pauta } from "../../../pautas/domain/entities/Pauta";
import { PautaRepository } from "../../../pautas/domain/repositories/PautaRepository";
import { HttpEsbAdapter, AdImportDto } from "../../infraestructure/http/HttpEsbAdapter";

export interface ImportPautasInput {
    date: string; // YYYY-MM-DD
}

export interface ImportPautasOutput {
    imported: number;
    pautas: Pauta[];
}

/**
 * Consulta las pautas de ventas (via ESB → vw_pauta_dataplan) para una fecha
 * y las persiste como pautas sin asignar en la base local.
 */
export class ImportPautasFromVentasUseCase implements UseCase<ImportPautasInput, ImportPautasOutput> {
    constructor(
        private readonly esbAdapter: HttpEsbAdapter,
        private readonly pautaRepository: PautaRepository
    ) {}

    async execute(input: ImportPautasInput): Promise<ImportPautasOutput> {
        const ads = await this.esbAdapter.getAdsByDate(input.date);

        const pautas: Pauta[] = [];
        for (const ad of ads) {
            const desc = this.buildDescription(ad);

            // No duplicar: si ya existe con la misma descripción y cover_date, saltar.
            const existing = await this.pautaRepository.findByDescription(desc);
            const alreadyImported = existing.some(p => p.cover_date === (ad.coverDate || input.date));
            if (alreadyImported) {
                pautas.push(existing.find(p => p.cover_date === (ad.coverDate || input.date))!);
                continue;
            }

            const pauta = new Pauta(
                0,
                desc,
                ad.cuadrosAlto,
                ad.cuadrosAncho,
                null,
                null,
                null,
                "text",
                null,
                ad.coverDate || input.date
            );
            const saved = await this.pautaRepository.save(pauta);
            pautas.push(saved);
        }

        return { imported: pautas.length, pautas };
    }

    private buildDescription(ad: AdImportDto): string {
        // ponytail: compact description from available fields; enough to identify the ad on the canvas
        const parts = [ad.customer, ad.product, ad.formatFid].filter(Boolean);
        return parts.join(" – ") || `Pauta ${ad.fid}`;
    }
}
