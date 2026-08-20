import PDFDocument from "pdfkit";
import { LayoutContract } from "../../application/contracts/LayoutContract";

// ponytail: 1mm = 2.83465pt (PDF points). PDFKit uses points natively.
const MM_TO_PT = 2.83465;

/**
 * Genera un PDF tipo "dummy/boceto" del layout de una edición.
 * Cada página del diario se dibuja en una página del PDF con:
 * - La grilla de referencia (líneas punteadas)
 * - Los márgenes
 * - Los rectángulos de cada pauta posicionada con su nombre
 */
export function generateDummyPdf(layout: LayoutContract): Promise<Buffer> {
    const { edition, pages } = layout;

    const pageWidthPt = edition.ancho_mm * MM_TO_PT;
    const pageHeightPt = edition.alto_mm * MM_TO_PT;

    const doc = new PDFDocument({
        size: [pageWidthPt, pageHeightPt],
        margins: { top: 0, bottom: 0, left: 0, right: 0 },
        autoFirstPage: false,
    });

    const chunks: Buffer[] = [];
    doc.on("data", (chunk: Buffer) => chunks.push(chunk));

    const finished = new Promise<Buffer>((resolve, reject) => {
        doc.on("end", () => resolve(Buffer.concat(chunks)));
        doc.on("error", reject);
    });

    for (let i = 0; i < pages.length; i++) {
        const page = pages[i];
        doc.addPage({ size: [pageWidthPt, pageHeightPt], margins: { top: 0, bottom: 0, left: 0, right: 0 } });

        // Header
        doc.fontSize(8).fillColor("#666666")
            .text(`Página ${page.no_pagina}`, 4, 4, { width: pageWidthPt - 8 });

        // Draw margins area
        const mTop = edition.margen_superior_mm * MM_TO_PT;
        const mBottom = edition.margen_inferior_mm * MM_TO_PT;
        const mLeft = edition.margen_izquierdo_mm * MM_TO_PT;
        const mRight = edition.margen_derecho_mm * MM_TO_PT;

        const contentX = mLeft;
        const contentY = mTop;
        const contentW = pageWidthPt - mLeft - mRight;
        const contentH = pageHeightPt - mTop - mBottom;

        // Margin lines (magenta, thin)
        doc.save()
            .rect(contentX, contentY, contentW, contentH)
            .lineWidth(0.5)
            .strokeColor("#FF00FF")
            .dash(3, { space: 2 })
            .stroke()
            .restore();

        // Grid lines (light gray, dotted)
        const colWidth = contentW / edition.cuadros_ancho;
        const rowHeight = contentH / edition.cuadros_alto;

        doc.save().strokeColor("#CCCCCC").lineWidth(0.25).dash(1, { space: 2 });
        for (let col = 1; col < edition.cuadros_ancho; col++) {
            const x = contentX + col * colWidth;
            doc.moveTo(x, contentY).lineTo(x, contentY + contentH).stroke();
        }
        for (let row = 1; row < edition.cuadros_alto; row++) {
            const y = contentY + row * rowHeight;
            doc.moveTo(contentX, y).lineTo(contentX + contentW, y).stroke();
        }
        doc.restore();

        // Draw pautas — usar coordenadas de grilla (locales a la página)
        for (const pauta of page.pautas) {
            const x = contentX + pauta.ubicacion_cuadros_x * colWidth;
            const y = contentY + pauta.ubicacion_cuadros_y * rowHeight;
            const w = pauta.cuadros_ancho * colWidth;
            const h = pauta.cuadros_alto * rowHeight;

            // Filled rectangle
            doc.save()
                .rect(x, y, w, h)
                .fillColor("#DBEAFE")
                .fill()
                .restore();

            // Border
            doc.save()
                .rect(x, y, w, h)
                .lineWidth(1)
                .strokeColor("#2563EB")
                .undash()
                .stroke()
                .restore();

            // Label
            const label = pauta.descripcion_pauta.length > 25
                ? pauta.descripcion_pauta.slice(0, 22) + "..."
                : pauta.descripcion_pauta;
            const sizeLabel = `${pauta.cuadros_ancho}×${pauta.cuadros_alto}`;

            doc.save().fillColor("#1E40AF").fontSize(6);
            doc.text(label, x + 2, y + 2, { width: w - 4, height: h - 4, ellipsis: true });
            doc.text(sizeLabel, x + 2, y + h - 10, { width: w - 4 });
            doc.restore();
        }
    }

    doc.end();
    return finished;
}
