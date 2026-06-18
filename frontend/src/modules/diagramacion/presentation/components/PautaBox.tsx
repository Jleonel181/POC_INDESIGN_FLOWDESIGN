"use client";

import { Pauta } from "../../domain/entities/Pauta";

interface PautaBoxProps {
  pauta: Pauta;
  cellWidth: number;
  cellHeight: number;
  onSelect?: (pauta: Pauta) => void;
  isSelected?: boolean;
}

export function PautaBox({ pauta, cellWidth, cellHeight, onSelect, isSelected }: PautaBoxProps) {
  const pos  = pauta.getPosition();
  const size = pauta.getSize();

  const style: React.CSSProperties = {
    position:        "absolute",
    boxSizing:       "border-box",
    left:            pos.x  * cellWidth,
    top:             pos.y  * cellHeight,
    width:           size.width  * cellWidth,
    height:          size.height * cellHeight,
    border:          `1px solid ${isSelected ? "#097CC8" : "#6b7280"}`,
    backgroundColor: isSelected ? "rgba(9,124,200,0.15)" : "rgba(95,212,216,0.35)",
    borderRadius:    3,
    cursor:          "pointer",
    display:         "flex",
    alignItems:      "center",
    justifyContent:  "center",
    overflow:        "hidden",
    transition:      "border-color 0.15s, background-color 0.15s",
  };

  return (
    <div style={style} onClick={() => onSelect?.(pauta)} title={pauta.descripcion}>
      <span
        style={{
          fontSize:     10,
          lineHeight:   1.2,
          padding:      "0 3px",
          overflow:     "hidden",
          textOverflow: "ellipsis",
          whiteSpace:   "nowrap",
          maxWidth:     "100%",
          color:        "#1a1a2e",
        }}
      >
        {pauta.descripcion}
      </span>
    </div>
  );
}
