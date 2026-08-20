import PDFDocument from "pdfkit";
import { LayoutContract } from "../../application/contracts/LayoutContract";

const MM_TO_PT = 2.83465;

// ponytail: tabloide landscape para más espacio horizontal con spreads
const SHEET_WIDTH = 17 * 72;   // 1224pt
const SHEET_HEIGHT = 11 * 72;  // 792pt
const MARGIN = 36;
const GAP = 14;
const HEADER_HEIGHT = 24;

interface Spread {
    pages: LayoutContract["pages"][number][];
}

/**
 * Agrupa las páginas en spreads (como facing pages en un periódico):
 * - Página 1 sola (portada)
 * - Pares de páginas interiores
 * - Última página sola si queda impar (contraportada)
 */
function buildSpreads(pages: LayoutContract["pages"], facingPages: boolean): Spread[] {
    if (!facingPages) {
        return pages.map(p => ({ pages: [p] }));
    }

    const spreads: Spread[] = [];
    let i = 0;

    // Portada sola
    if (pages.length > 0) {
        spreads.push({ pages: [pages[0]] });
        i = 1;
    }

    // Interiores en pares
    while (i < pages.length - 1) {
        spreads.push({ pages: [pages[i], pages[i + 1]] });
        i += 2;
    }

    // Contraportada sola si quedó
    if (i < pages.length) {
        spreads.push({ pages: [pages[i]] });
    }

    return spreads;
}

function drawPageThumb(
    doc: InstanceType<typeof PDFDocument>,
    page: LayoutContract["pages"][number],
    edition: LayoutContract["edition"],
    originX: number,
    originY: number,
    thumbW: number,
    thumbH: number,
    scale: number
) {
    // Clip to thumbnail bounds — only for pautas that might overflow
    doc.save();

    // Page background
    doc.rect(originX, originY, thumbW, thumbH).fillColor("#FFFFFF").fill();

    // Page border
    doc.rect(originX, originY, thumbW, thumbH)
        .lineWidth(0.5).strokeColor("#999999").undash().stroke();

    // Page number
    doc.fontSize(5).fillColor("#666666")
        .text(`Pág. ${page.no_pagina}`, originX + 2, originY + 1, { width: thumbW - 4, lineBreak: false });

    // Margins
    const mTop = edition.margen_superior_mm * MM_TO_PT * scale;
    const mBottom = edition.margen_inferior_mm * MM_TO_PT * scale;
    const mLeft = edition.margen_izquierdo_mm * MM_TO_PT * scale;
    const mRight = edition.margen_derecho_mm * MM_TO_PT * scale;

    const contentX = originX + mLeft;
    const contentY = originY + mTop;
    const contentW = thumbW - mLeft - mRight;
    const contentH = thumbH - mTop - mBottom;

    doc.save()
        .rect(contentX, contentY, contentW, contentH)
        .lineWidth(0.3).strokeColor("#FF00FF").dash(2, { space: 1 }).stroke()
        .restore();

    // Grid
    const colWidth = contentW / edition.cuadros_ancho;
    const rowHeight = contentH / edition.cuadros_alto;

    doc.save().strokeColor("#8C8C8C").lineWidth(0.15).dash(1, { space: 1 });
    for (let c = 1; c < edition.cuadros_ancho; c++) {
        const x = contentX + c * colWidth;
        doc.moveTo(x, contentY).lineTo(x, contentY + contentH).stroke();
    }
    for (let r = 1; r < edition.cuadros_alto; r++) {
        const y = contentY + r * rowHeight;
        doc.moveTo(contentX, y).lineTo(contentX + contentW, y).stroke();
    }
    doc.restore();

    // Pautas — usar coordenadas de grilla (locales a la página, no de spread)
    doc.save();
    doc.rect(originX, originY, thumbW, thumbH).clip();
    for (const pauta of page.pautas) {
        const x = contentX + pauta.ubicacion_cuadros_x * colWidth;
        const y = contentY + pauta.ubicacion_cuadros_y * rowHeight;
        const w = pauta.cuadros_ancho * colWidth;
        const h = pauta.cuadros_alto * rowHeight;

        doc.save().rect(x, y, w, h).fillColor("#DBEAFE").fill().restore();
        doc.save().rect(x, y, w, h).lineWidth(0.5).strokeColor("#2563EB").undash().stroke().restore();

        if (w > 18 && h > 10) {
            const label = pauta.descripcion_pauta.length > 18
                ? pauta.descripcion_pauta.slice(0, 15) + "..."
                : pauta.descripcion_pauta;
            doc.save().fontSize(4).fillColor("#1E40AF")
                .text(label, x + 1, y + 1, { width: w - 2, height: h - 2, ellipsis: true })
                .restore();
        }
    }

    // End pauta clip
    doc.restore();

    // End function save
    doc.restore();
}

/**
 * Genera un PDF panorámico tipo "overview" de toda la edición.
 * Facing pages aparecen juntas como spreads.
 * Cada pauta se clipea dentro de su thumbnail.
 */
export function generateOverviewPdf(layout: LayoutContract): Promise<Buffer> {
    const { edition, pages } = layout;

    const spreads = buildSpreads(pages, edition.facing_pages);

    const doc = new PDFDocument({
        size: [SHEET_WIDTH, SHEET_HEIGHT],
        margins: { top: 0, bottom: 0, left: 0, right: 0 },
        autoFirstPage: false,
    });

    const chunks: Buffer[] = [];
    doc.on("data", (chunk: Buffer) => chunks.push(chunk));

    const finished = new Promise<Buffer>((resolve, reject) => {
        doc.on("end", () => resolve(Buffer.concat(chunks)));
        doc.on("error", reject);
    });

    const availW = SHEET_WIDTH - 2 * MARGIN;
    const availH = SHEET_HEIGHT - 2 * MARGIN - HEADER_HEIGHT;

    const pageAspect = edition.ancho_mm / edition.alto_mm;

    // Un spread doble ocupa el ancho de 2 páginas + un pequeño gap interno
    const SPREAD_INNER_GAP = 2;

    // Calcular tamaño de thumb para que quepan ~3 spreads por fila
    let cols = 3;
    // El ancho máximo de un spread (doble) = 2*thumbW + SPREAD_INNER_GAP
    // Total por fila = cols * (2*thumbW + SPREAD_INNER_GAP) + (cols-1)*GAP <= availW
    // Simplificamos: asumimos el peor caso (todos dobles)
    let thumbW = (availW - (cols - 1) * GAP - cols * SPREAD_INNER_GAP) / (cols * 2);
    let thumbH = thumbW / pageAspect;

    if (thumbH > availH * 0.45) {
        // Si los thumbs son muy altos, reducir
        thumbH = availH * 0.45;
        thumbW = thumbH * pageAspect;
    }

    const spreadW_single = thumbW;
    const spreadW_double = thumbW * 2 + SPREAD_INNER_GAP;

    // Layout spreads into rows
    interface RowItem { spread: Spread; x: number; width: number }
    const rows: RowItem[][] = [];
    let currentRow: RowItem[] = [];
    let currentX = 0;

    for (const spread of spreads) {
        const sw = spread.pages.length === 2 ? spreadW_double : spreadW_single;
        const neededX = currentX > 0 ? currentX + GAP + sw : sw;

        if (neededX > availW && currentRow.length > 0) {
            rows.push(currentRow);
            currentRow = [];
            currentX = 0;
        }

        const x = currentX > 0 ? currentX + GAP : 0;
        currentRow.push({ spread, x, width: sw });
        currentX = x + sw;
    }
    if (currentRow.length > 0) rows.push(currentRow);

    // Calcular cuántas filas caben por hoja
    const rowHeight = thumbH + GAP;
    const rowsPerSheet = Math.max(1, Math.floor((availH + GAP) / rowHeight));

    for (let sheetRow = 0; sheetRow < rows.length; sheetRow += rowsPerSheet) {
        doc.addPage({ size: [SHEET_WIDTH, SHEET_HEIGHT], margins: { top: 0, bottom: 0, left: 0, right: 0 } });

        // Sheet header
        doc.fontSize(9).fillColor("#333333")
            .text(`Planificación – Edición #${edition.id} (${edition.ancho_mm}×${edition.alto_mm}mm)`,
                MARGIN, MARGIN, { width: availW });

        const sheetRows = rows.slice(sheetRow, sheetRow + rowsPerSheet);

        for (let ri = 0; ri < sheetRows.length; ri++) {
            const row = sheetRows[ri];
            const rowY = MARGIN + HEADER_HEIGHT + ri * rowHeight;

            for (const item of row) {
                const { spread, x } = item;
                const baseX = MARGIN + x;

                for (let pi = 0; pi < spread.pages.length; pi++) {
                    const pageOriginX = baseX + pi * (thumbW + SPREAD_INNER_GAP);
                    const scale = thumbW / (edition.ancho_mm * MM_TO_PT);

                    drawPageThumb(doc, spread.pages[pi], edition, pageOriginX, rowY, thumbW, thumbH, scale);
                }
            }
        }
    }

    doc.end();
    return finished;
}
