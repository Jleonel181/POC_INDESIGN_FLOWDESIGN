"use client";

import { use } from "react";
import { useDiagramador } from "@/modules/diagramacion/presentation/hooks/useDiagramador";
import { DroppableCanvas } from "@/modules/diagramacion/presentation/components/DroppableCanvas";
import { LayoutJsonPreview } from "@/modules/diagramacion/presentation/components/LayoutJsonPreview";
import { PromptPanel } from "@/modules/diagramacion/presentation/components/PromptPanel";
import { PautasSidebar } from "@/modules/diagramacion/presentation/components/PautasSidebar";
import { Edition } from "@/modules/diagramacion/domain/entities/Edition";
import { Page } from "@/modules/diagramacion/domain/entities/Page";

function FacingPagesView({ edition, pages, onPautaAssigned }: { edition: Edition; pages: Page[]; onPautaAssigned: () => void }) {
  const spreads: Page[][] = [];

  for (let i = 0; i < pages.length; i++) {
    if (i === 0) {
      spreads.push([pages[i]]);
    } else if (i === pages.length - 1 && pages.length % 2 === 0) {
      spreads.push([pages[i]]);
    } else if (i % 2 === 1 && i + 1 < pages.length) {
      spreads.push([pages[i], pages[i + 1]]);
      i++;
    } else {
      spreads.push([pages[i]]);
    }
  }

  const cover = spreads[0];
  const lastSpread = spreads[spreads.length - 1];
  const isLastSingle = spreads.length > 1 && lastSpread.length === 1;
  const middle = isLastSingle ? spreads.slice(1, -1) : spreads.slice(1);
  const backCover = isLastSingle ? lastSpread : null;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-5 gap-6">
        <div className="space-y-1 min-w-0">
          <span className="text-xs text-gray-500">Página {cover[0].noPagina}</span>
          <DroppableCanvas edition={edition} page={cover[0]} onPautaAssigned={onPautaAssigned} />
        </div>
      </div>

      {middle.length > 0 && (
        <div className="grid grid-cols-3 gap-6">
          {middle.map((spread, idx) => (
            <div key={idx} className="space-y-1 min-w-0">
              <span className="text-xs text-gray-500">
                {spread.length === 1
                  ? `Página ${spread[0].noPagina}`
                  : `Páginas ${spread[0].noPagina}–${spread[1].noPagina}`}
              </span>
              <div className="flex gap-0.5 overflow-x-auto">
                <div className="shrink-0" style={{ width: spread.length === 2 ? "50%" : "100%" }}>
                  <DroppableCanvas edition={edition} page={spread[0]} isFacing={spread.length === 2} onPautaAssigned={onPautaAssigned} />
                </div>
                {spread[1] && (
                  <div className="shrink-0" style={{ width: "50%" }}>
                    <DroppableCanvas edition={edition} page={spread[1]} isFacing={true} onPautaAssigned={onPautaAssigned} />
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {backCover && (
        <div className="grid grid-cols-5 gap-6">
          <div className="space-y-1 min-w-0">
            <span className="text-xs text-gray-500">Página {backCover[0].noPagina}</span>
            <DroppableCanvas edition={edition} page={backCover[0]} onPautaAssigned={onPautaAssigned} />
          </div>
        </div>
      )}
    </div>
  );
}

interface PageProps {
  params: Promise<{ editionId: string }>;
}

export default function DiagramadorPage({ params }: PageProps) {
  const { editionId } = use(params);
  const { edition, pages, loading, error, rawDTO, reload } = useDiagramador(Number(editionId));

  if (loading) {
    return <div className="flex items-center justify-center h-screen text-gray-500">Cargando diagramación...</div>;
  }

  if (error) {
    return <div className="flex items-center justify-center h-screen text-red-500">Error: {error}</div>;
  }

  if (!edition || pages.length === 0) {
    return <div className="flex items-center justify-center h-screen text-gray-500">No se encontró la edición</div>;
  }

  return (
    <div className="flex h-screen">
      {/* Sidebar con biblioteca de pautas */}
      <PautasSidebar />

      {/* Área principal */}
      <div className="flex-1 p-6 overflow-y-auto space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold text-gray-800">DesignFlow POC</h1>
            <p className="text-sm text-gray-500">
              {edition.anchoMm}×{edition.altoMm}mm | Grilla {edition.gridColumns}×{edition.gridRows} | {edition.noPaginas} páginas
            </p>
          </div>
          <div className="flex gap-2">
            <a
              href={`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001/api"}/layout/${editionId}/overview`}
              target="_blank"
              rel="noopener noreferrer"
              className="px-3 py-1.5 text-sm bg-gray-800 text-white rounded hover:bg-gray-900 transition-colors"
            >
              Overview PDF
            </a>
            <a
              href={`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001/api"}/layout/${editionId}/pdf`}
              target="_blank"
              rel="noopener noreferrer"
              className="px-3 py-1.5 text-sm border border-gray-300 text-gray-700 rounded hover:bg-gray-100 transition-colors"
            >
              Dummy PDF (1:1)
            </a>
          </div>
        </div>

        {edition.facingPages ? (
          <FacingPagesView edition={edition} pages={pages} onPautaAssigned={reload} />
        ) : (
          <div className="grid grid-cols-5 gap-6">
            {pages.map((page) => (
              <div key={page.id} className="space-y-1">
                <span className="text-xs text-gray-500">Página {page.noPagina}</span>
                <DroppableCanvas edition={edition} page={page} onPautaAssigned={reload} />
              </div>
            ))}
          </div>
        )}

        <PromptPanel edition={edition} pages={pages} />
        <LayoutJsonPreview dto={rawDTO} />
      </div>
    </div>
  );
}
