"use client";

import { useState, useRef } from "react";
import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { DiagramadorCanvas } from "./DiagramadorCanvas";
import { apiClient } from "@/shared/infraestructure/http/apiClient";

interface DroppableCanvasProps {
  edition: Edition;
  page: Page;
  isFacing?: boolean;
  onPautaAssigned: () => void;
}

/** Extrae el tamaño de la pauta desde los types del dataTransfer (legible durante dragOver). */
function getPautaSizeFromTypes(types: readonly string[]): { w: number; h: number } | null {
  for (const t of types) {
    const match = t.match(/^pauta-size\/(\d+)x(\d+)$/);
    if (match) return { w: Number(match[1]), h: Number(match[2]) };
  }
  return null;
}

export function DroppableCanvas({ edition, page, isFacing = false, onPautaAssigned }: DroppableCanvasProps) {
  const [dragOver, setDragOver] = useState(false);
  const [ghostPos, setGhostPos] = useState<{ col: number; row: number; w: number; h: number } | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const getGridPosition = (e: React.DragEvent): { col: number; row: number } | null => {
    const container = containerRef.current;
    if (!container) return null;

    const rect = container.getBoundingClientRect();
    const scale = rect.width / edition.anchoMm;

    const marginLeft = edition.margenIzquierdoMm * scale;
    const marginTop = edition.margenSuperiorMm * scale;
    const marginRight = edition.margenDerechoMm * scale;
    const marginBottom = edition.margenInferiorMm * scale;

    const contentWidth = rect.width - marginLeft - marginRight;
    const contentHeight = rect.height - marginTop - marginBottom;

    const cellWidth = contentWidth / edition.gridColumns;
    const cellHeight = contentHeight / edition.gridRows;

    const x = e.clientX - rect.left - marginLeft;
    const y = e.clientY - rect.top - marginTop;

    if (x < 0 || y < 0 || x > contentWidth || y > contentHeight) return null;

    const col = Math.floor(x / cellWidth);
    const row = Math.floor(y / cellHeight);

    return { col, row };
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "copy";
    setDragOver(true);

    const pos = getGridPosition(e);
    if (pos) {
      const size = getPautaSizeFromTypes(e.dataTransfer.types);
      const w = size?.w ?? 1;
      const h = size?.h ?? 1;
      setGhostPos({ col: pos.col, row: pos.row, w, h });
    } else {
      setGhostPos(null);
    }
  };

  const handleDragLeave = () => {
    setDragOver(false);
    setGhostPos(null);
  };

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    setGhostPos(null);

    const data = e.dataTransfer.getData("application/json");
    if (!data) return;

    const pauta = JSON.parse(data) as { id: number; cuadros_ancho: number; cuadros_alto: number };
    const pos = getGridPosition(e);
    if (!pos) return;

    if (pos.col + pauta.cuadros_ancho > edition.gridColumns || pos.row + pauta.cuadros_alto > edition.gridRows) {
      alert("La pauta no cabe en esta posición.");
      return;
    }

    try {
      await apiClient.put(`/pautas/${pauta.id}/assign`, {
        pagina_id: page.id,
        ubicacion_cuadros_x: pos.col,
        ubicacion_cuadros_y: pos.row,
      });
      onPautaAssigned();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Error al asignar pauta");
    }
  };

  // Cálculos para el ghost
  const contentRatioW = (edition.anchoMm - edition.margenIzquierdoMm - edition.margenDerechoMm) / edition.anchoMm;
  const contentRatioH = (edition.altoMm - edition.margenSuperiorMm - edition.margenInferiorMm) / edition.altoMm;
  const marginLeftPct = (edition.margenIzquierdoMm / edition.anchoMm) * 100;
  const marginTopPct = (edition.margenSuperiorMm / edition.altoMm) * 100;

  const exceedsGrid = ghostPos
    ? ghostPos.col + ghostPos.w > edition.gridColumns || ghostPos.row + ghostPos.h > edition.gridRows
    : false;

  return (
    <div
      ref={containerRef}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className={`relative transition-all ${dragOver ? "ring-2 ring-blue-400 ring-offset-1" : ""}`}
    >
      <DiagramadorCanvas edition={edition} page={page} isFacing={isFacing} />

      {dragOver && ghostPos && (
        <div
          className={`absolute pointer-events-none border-2 border-dashed rounded-sm ${
            exceedsGrid ? "bg-red-200/40 border-red-500" : "bg-green-200/40 border-green-500"
          }`}
          style={{
            left: `${marginLeftPct + (ghostPos.col / edition.gridColumns) * contentRatioW * 100}%`,
            top: `${marginTopPct + (ghostPos.row / edition.gridRows) * contentRatioH * 100}%`,
            width: `${(ghostPos.w / edition.gridColumns) * contentRatioW * 100}%`,
            height: `${(ghostPos.h / edition.gridRows) * contentRatioH * 100}%`,
          }}
        />
      )}
    </div>
  );
}
