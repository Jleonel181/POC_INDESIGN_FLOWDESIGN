"use client";

import Link from "next/link";
import { DiagramacionDTO } from "@/modules/diagramacion/application/dto/DiagramacionDTO";

interface EditionCardProps {
  edition: DiagramacionDTO["edition"];
}

export function EditionCard({ edition }: EditionCardProps) {
  return (
    <Link
      href={`/diagramador/${edition.id}`}
      className="group flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-5 shadow-sm transition hover:border-[#00b1ff] hover:shadow-md"
    >
      <div className="flex items-center justify-between">
        <span
          className="flex h-9 w-9 items-center justify-center rounded-lg text-white text-sm font-bold"
          style={{ background: "#00b1ff" }}
        >
          {edition.id}
        </span>
        <span className="material-symbols-rounded text-gray-300 group-hover:text-[#00b1ff] transition" style={{ fontSize: "1.2rem" }}>
          arrow_forward
        </span>
      </div>

      <div>
        <p className="font-semibold text-gray-800">Edición #{edition.id}</p>
        <p className="text-xs text-gray-500 mt-0.5">
          {edition.ancho_mm}×{edition.alto_mm} mm · {edition.no_paginas} páginas
        </p>
      </div>

      <div className="flex flex-wrap gap-2 mt-auto">
        <span className="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs text-gray-600">
          Grilla {edition.cuadros_ancho}×{edition.cuadros_alto}
        </span>
        {edition.facing_pages && (
          <span className="rounded-full px-2.5 py-0.5 text-xs text-white" style={{ background: "#49beff" }}>
            Facing pages
          </span>
        )}
      </div>
    </Link>
  );
}
