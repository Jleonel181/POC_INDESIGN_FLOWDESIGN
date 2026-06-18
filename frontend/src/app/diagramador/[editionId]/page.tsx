"use client";

import { use } from "react";
import { useDiagramador } from "@/modules/diagramacion/presentation/hooks/useDiagramador";
import { DiagramadorCanvas } from "@/modules/diagramacion/presentation/components/DiagramadorCanvas";
import { LayoutJsonPreview } from "@/modules/diagramacion/presentation/components/LayoutJsonPreview";
import { Edition } from "@/modules/diagramacion/domain/entities/Edition";
import { Page } from "@/modules/diagramacion/domain/entities/Page";

function FacingPagesView({ edition, pages }: { edition: Edition; pages: Page[] }) {
  const spreads: Page[][] = [];

  for (let i = 0; i < pages.length; i++) {
    if (i === 0) {
      // Portada: sola
      spreads.push([pages[i]]);
    } else if (i === pages.length - 1 && pages.length % 2 === 0) {
      // Contraportada: sola (si el total de páginas es par, la última queda sola)
      spreads.push([pages[i]]);
    } else if (i % 2 === 1 && i + 1 < pages.length) {
      // Páginas enfrentadas
      spreads.push([pages[i], pages[i + 1]]);
      i++; // skip next
    } else {
      // Página suelta restante
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
      {/* Portada sola */}
      <div className="grid grid-cols-5 gap-6">
        <div className="space-y-1 min-w-0">
          <span className="text-xs text-gray-500">Página {cover[0].noPagina}</span>
          <DiagramadorCanvas edition={edition} page={cover[0]} />
        </div>
      </div>

      {/* Spreads en grid */}
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
                  <DiagramadorCanvas edition={edition} page={spread[0]} isFacing={spread.length === 2} />
                </div>
                {spread[1] && (
                  <div className="shrink-0" style={{ width: "50%" }}>
                    <DiagramadorCanvas edition={edition} page={spread[1]} isFacing={true} />
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Contraportada sola */}
      {backCover && (
        <div className="grid grid-cols-5 gap-6">
          <div className="space-y-1 min-w-0">
            <span className="text-xs text-gray-500">Página {backCover[0].noPagina}</span>
            <DiagramadorCanvas edition={edition} page={backCover[0]} />
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
  const { edition, pages, loading, error, rawDTO } = useDiagramador(Number(editionId));

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
    <div className="p-6 mx-auto space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-gray-800">
             DesignFlow POC
          </h1>
          <p className="text-sm text-gray-500">
            {edition.anchoMm}×{edition.altoMm}mm | Grilla {edition.gridColumns}×{edition.gridRows} | {edition.noPaginas} páginas
          </p>
        </div>
        {/* <button onClick={saveLayout} className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700">
          Guardar Layout
        </button> */}
      </div>

      {edition.facingPages ? (
        <FacingPagesView edition={edition} pages={pages} />
      ) : (
        <div className="grid grid-cols-5 gap-6">
          {pages.map((page) => (
            <div key={page.id} className="space-y-1">
              <span className="text-xs text-gray-500">Página {page.noPagina}</span>
              <DiagramadorCanvas edition={edition} page={page} />
            </div>
          ))}
        </div>
      )}

      <LayoutJsonPreview dto={rawDTO} />
    </div>
  );
}
