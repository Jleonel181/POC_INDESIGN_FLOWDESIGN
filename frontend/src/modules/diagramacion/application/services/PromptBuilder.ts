import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { Pauta } from "../../domain/entities/Pauta";

export type PromptMode = "single" | "spread" | "full";

export interface PromptContext {
  edition: Edition;
  pages: Page[];
  selectedPageIndex: number;
  mode: PromptMode;
  includePageNumbers?: boolean;
}

function formatPauta(p: Pauta): string {
  const pos  = p.getPosition();
  const size = p.getSize();
  const b    = p.indesignBounds;
  const mmInfo = b
    ? ` | bounds: top=${b.topMm}mm left=${b.leftMm}mm bottom=${b.bottomMm}mm right=${b.rightMm}mm`
    : "";
  return `  - "${p.descripcion}": grid(col=${pos.x}, row=${pos.y}, w=${size.width}, h=${size.height})${mmInfo}`;
}

function pageBlock(page: Page, edition: Edition): string {
  const pautaLines = page.pautas.length
    ? page.pautas.map(formatPauta).join("\n")
    : "  (sin pautas)";
  return `Página ${page.noPagina}:\n${pautaLines}`;
}

function resolveSpread(pages: Page[], idx: number): Page[] {
  if (idx === 0) return [pages[0]];
  const isLeft = idx % 2 === 1;
  if (isLeft && pages[idx + 1]) return [pages[idx], pages[idx + 1]];
  if (!isLeft && pages[idx - 1]) return [pages[idx - 1], pages[idx]];
  return [pages[idx]];
}

/**
 * Builds the IDML-based pagination instruction section.
 * Uses the folio pattern from pag_simple.idml (single pages) or Pag_Ind.idml (facing pages).
 */
function buildPaginationSection(edition: Edition, targetPages: Page[]): string {
  const templateFile = edition.facingPages ? "Pag_Ind.idml" : "pag_simple.idml";

  const folioExample = edition.facingPages
    ? buildFacingPagesFolioExample(targetPages)
    : buildSimpleFolioExample(targetPages);

  return `

## Numeración de páginas (folio)
Plantilla IDML de referencia: \`${templateFile}\`

Agrega un marco de texto de folio en el pie de cada página con las siguientes características:
- Fuente: Popular Bold, 10pt, leading 13.55pt, kerning óptico.
- Alineación vertical: inferior (bottom-align).
- El número de página se genera con el marcador IDML \`<?ACE 18?>\` (auto page number).
${folioExample}`;
}

function buildFacingPagesFolioExample(targetPages: Page[]): string {
  const lines = targetPages.map((p) => {
    const isLeft = p.noPagina % 2 === 0;
    if (isLeft) {
      return `- Página ${p.noPagina} (izquierda): "[número]. Sección | Fecha" — alineado a la izquierda, número en color C=100.`;
    }
    return `- Página ${p.noPagina} (derecha): "Fecha | Sección. [número]" — alineado a la derecha, número en color C=100.`;
  });

  return `
Formato facing pages (doble):
${lines.join("\n")}
- Estructura IDML: CharacterStyleRange con FillColor="Color/C=100 M=0 Y=0 K=0" para el número/sección.`;
}

function buildSimpleFolioExample(targetPages: Page[]): string {
  const pageNums = targetPages.map((p) => p.noPagina).join(", ");
  return `
Formato página simple:
- Páginas: ${pageNums}
- Contenido: "[número]. Fecha" — alineado a la izquierda, número en color C=100.
- Estructura IDML: CharacterStyleRange con FillColor="Color/C=100 M=0 Y=0 K=0" para el número.`;
}

export function buildPrompt(ctx: PromptContext): string {
  const { edition, pages, selectedPageIndex, mode, includePageNumbers } = ctx;

  const editionInfo = [
    `Edición: ${edition.anchoMm}×${edition.altoMm}mm`,
    `Grilla: ${edition.gridColumns} columnas × ${edition.gridRows} filas`,
    `Márgenes: sup=${edition.margenSuperiorMm}mm inf=${edition.margenInferiorMm}mm izq=${edition.margenIzquierdoMm}mm der=${edition.margenDerechoMm}mm`,
    `Facing pages: ${edition.facingPages ? "sí" : "no"}`,
  ].join(" | ");

  let targetPages: Page[];
  if (mode === "full") {
    targetPages = pages;
  } else if (mode === "spread" && edition.facingPages) {
    targetPages = resolveSpread(pages, selectedPageIndex);
  } else {
    targetPages = [pages[selectedPageIndex]];
  }

  const pagesSection = targetPages.map((p) => pageBlock(p, edition)).join("\n\n");

  const paginationSection = includePageNumbers
    ? buildPaginationSection(edition, targetPages)
    : "";

  return `Eres un asistente de diagramación editorial. Usa el MCP de Affinity Publisher para crear el layout descrito.

## Contexto de la edición
${editionInfo}

## Layout a diagramar
${pagesSection}

## Instrucciones
- Crea un documento en Affinity Publisher con las dimensiones y márgenes indicados.
- Por cada pauta, crea un marco de texto en la posición y tamaño de grilla especificados.
- Usa los bounds en mm si están disponibles para posicionamiento exacto.
- Nombra cada marco con la descripción de la pauta.
- Respeta el sistema de coordenadas: origen en esquina superior-izquierda del área de contenido.${paginationSection}`;
}
