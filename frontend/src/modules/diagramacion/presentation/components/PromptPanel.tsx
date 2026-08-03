"use client";

import { useState } from "react";
import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001/api";

interface PromptPanelProps {
  edition: Edition;
  pages: Page[];
}

export function PromptPanel({ edition }: PromptPanelProps) {
  const [includePageNumbers, setIncludePageNumbers] = useState(false);
  const [downloading, setDownloading] = useState(false);

  const handleDownloadIdml = async () => {
    setDownloading(true);
    try {
      const folio = includePageNumbers ? "true" : "false";
      const url = `${API_BASE}/layout/${edition.id}/idml?folio=${folio}`;
      const response = await fetch(url);

      if (!response.ok) {
        const err = await response.json().catch(() => ({ error: response.statusText }));
        alert(`Error al generar IDML: ${err.detail || err.error}`);
        return;
      }

      const blob = await response.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = `edicion-${edition.id}.idml`;
      a.click();
      URL.revokeObjectURL(a.href);
    } catch (err) {
      alert(`Error de conexión: ${err instanceof Error ? err.message : err}`);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div className="mt-4 border border-gray-200 rounded-lg overflow-hidden">
      <div className="flex items-center gap-3 px-4 py-2.5 bg-gray-50">
        <span className="text-sm font-medium text-gray-700">Exportar IDML</span>

        <button
          onClick={() => setIncludePageNumbers((v) => !v)}
          className={`text-xs px-3 py-1 rounded border transition-colors ${
            includePageNumbers
              ? "bg-blue-600 text-white border-blue-600"
              : "bg-white text-gray-600 border-gray-300 hover:border-blue-400"
          }`}
        >
          {includePageNumbers ? "# Con folio" : "# Sin folio"}
        </button>

        <button
          onClick={handleDownloadIdml}
          disabled={downloading}
          className="ml-auto text-sm px-4 py-1.5 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 transition-colors"
        >
          {downloading ? "Generando..." : "Descargar IDML"}
        </button>
      </div>
    </div>
  );
}
