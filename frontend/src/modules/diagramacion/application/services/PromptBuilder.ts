import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { Pauta } from "../../domain/entities/Pauta";

export type PromptMode = "single" | "spread" | "full";

export interface PromptContext {
  edition: Edition;
  pages: Page[];
  selectedPageIndex: number;
  mode: PromptMode;
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

export function buildPrompt(ctx: PromptContext): string {
  const { edition, pages, selectedPageIndex, mode } = ctx;

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
- Respeta el sistema de coordenadas: origen en esquina superior-izquierda del área de contenido.`;
}
