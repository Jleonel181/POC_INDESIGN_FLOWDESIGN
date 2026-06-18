"use client";

import { useState } from "react";
import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { Pauta } from "../../domain/entities/Pauta";
import { PautaBox } from "./PautaBox";
import { useElementWidth } from "@/shared/presentation/hooks/useElementWidth";

interface DiagramadorCanvasProps {
  edition: Edition;
  page: Page;
  isFacing?: boolean;
  onPautaSelect?: (pauta: Pauta) => void;
}

export function DiagramadorCanvas({ edition, page, isFacing = false, onPautaSelect }: DiagramadorCanvasProps) {
  const [selectedPautaId, setSelectedPautaId] = useState<number | null>(null);
  const [wrapperRef, canvasWidth] = useElementWidth(isFacing ? 160 : 220);

  const scale = canvasWidth > 0 ? canvasWidth / edition.anchoMm : 0;
  const canvasHeight = edition.altoMm * scale;

  const marginTop    = edition.margenSuperiorMm  * scale;
  const marginBottom = edition.margenInferiorMm  * scale;
  const marginLeft   = edition.margenIzquierdoMm * scale;
  const marginRight  = edition.margenDerechoMm   * scale;

  const contentWidth  = canvasWidth  - marginLeft - marginRight;
  const contentHeight = canvasHeight - marginTop  - marginBottom;

  const cellWidth  = contentWidth  / edition.gridColumns;
  const cellHeight = contentHeight / edition.gridRows;

  const handleSelect = (pauta: Pauta) => {
    setSelectedPautaId(pauta.id);
    onPautaSelect?.(pauta);
  };

  return (
    <div ref={wrapperRef} className="w-full">
      <div
        className="relative shrink-0 border border-gray-300 bg-white shadow-sm overflow-hidden"
        style={{ width: canvasWidth, height: canvasHeight, boxSizing: "border-box" }}
      >
        <svg
          className="absolute inset-0 pointer-events-none"
          width={canvasWidth}
          height={canvasHeight}
          style={{ display: "block" }}
        >
          <rect
            x={marginLeft}
            y={marginTop}
            width={contentWidth}
            height={contentHeight}
            fill="none"
            stroke="#097CC8"
            strokeWidth={0.75}
            strokeDasharray="4 2"
          />
          {Array.from({ length: edition.gridColumns + 1 }, (_, i) => (
            <line
              key={`v-${i}`}
              x1={marginLeft + i * cellWidth}
              y1={marginTop}
              x2={marginLeft + i * cellWidth}
              y2={marginTop + contentHeight}
              stroke="#097CC8"
              strokeWidth={0.5}
            />
          ))}
          {Array.from({ length: edition.gridRows + 1 }, (_, i) => (
            <line
              key={`h-${i}`}
              x1={marginLeft}
              y1={marginTop + i * cellHeight}
              x2={marginLeft + contentWidth}
              y2={marginTop + i * cellHeight}
              stroke="#097CC8"
              strokeWidth={0.5}
            />
          ))}
        </svg>

        <div
          className="absolute overflow-hidden"
          style={{ top: marginTop, left: marginLeft, width: contentWidth, height: contentHeight }}
        >
          {page.pautas.map((pauta) => (
            <PautaBox
              key={pauta.id}
              pauta={pauta}
              cellWidth={cellWidth}
              cellHeight={cellHeight}
              onSelect={handleSelect}
              isSelected={pauta.id === selectedPautaId}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
