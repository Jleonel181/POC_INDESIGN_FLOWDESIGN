"use client";

import { useState } from "react";
import { Pauta } from "../../domain/entities/Pauta";
import { apiClient } from "@/shared/infraestructure/http/apiClient";

interface PautaBoxProps {
  pauta: Pauta;
  cellWidth: number;
  cellHeight: number;
  onSelect?: (pauta: Pauta) => void;
  onUnassigned?: () => void;
  isSelected?: boolean;
}

export function PautaBox({ pauta, cellWidth, cellHeight, onSelect, onUnassigned, isSelected }: PautaBoxProps) {
  const [hovered, setHovered] = useState(false);
  const [unassigning, setUnassigning] = useState(false);

  const pos = pauta.getPosition();
  const size = pauta.getSize();

  const handleUnassign = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setUnassigning(true);
    try {
      await apiClient.put(`/pautas/${pauta.id}/unassign`, {});
      onUnassigned?.();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Error al desasignar");
    } finally {
      setUnassigning(false);
    }
  };

  const style: React.CSSProperties = {
    position: "absolute",
    boxSizing: "border-box",
    left: pos.x * cellWidth,
    top: pos.y * cellHeight,
    width: size.width * cellWidth,
    height: size.height * cellHeight,
    border: `1px solid ${isSelected ? "#097CC8" : "#6b7280"}`,
    backgroundColor: isSelected ? "rgba(9,124,200,0.15)" : "rgba(95,212,216,0.35)",
    borderRadius: 3,
    cursor: "pointer",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    overflow: "visible",
    transition: "border-color 0.15s, background-color 0.15s",
  };

  return (
    <div
      style={style}
      onClick={() => onSelect?.(pauta)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      title={pauta.descripcion}
    >
      <span
        style={{
          fontSize: 10,
          lineHeight: 1.2,
          padding: "0 3px",
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
          maxWidth: "100%",
          color: "#1a1a2e",
        }}
      >
        {pauta.descripcion}
      </span>

      {hovered && (
        <button
          onClick={handleUnassign}
          disabled={unassigning}
          className="absolute 6 left-1/2 -translate-x-1/2 text-[9px] px-2 py-0.5 bg-red-600 text-white rounded shadow hover:bg-red-700 disabled:opacity-50 whitespace-nowrap z-10"
        >
          {unassigning ? "..." : "Desasignar pauta"}
        </button>
      )}
    </div>
  );
}
