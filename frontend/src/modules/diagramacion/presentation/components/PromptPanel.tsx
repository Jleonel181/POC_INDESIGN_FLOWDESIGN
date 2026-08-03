"use client";

import { useState, useMemo } from "react";
import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { buildPrompt, PromptMode } from "../../application/services/PromptBuilder";

interface PromptPanelProps {
  edition: Edition;
  pages: Page[];
}

export function PromptPanel({ edition, pages }: PromptPanelProps) {
  const [open, setOpen] = useState(false);
  const [selectedPageIndex, setSelectedPageIndex] = useState(0);
  const [mode, setMode] = useState<PromptMode>("single");
  const [copied, setCopied] = useState(false);
  const [includePageNumbers, setIncludePageNumbers] = useState(false);

  const prompt = useMemo(
    () => buildPrompt({ edition, pages, selectedPageIndex, mode, includePageNumbers }),
    [edition, pages, selectedPageIndex, mode, includePageNumbers]
  );

  const handleCopy = () => {
    navigator.clipboard.writeText(prompt);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="mt-4 border border-gray-200 rounded-lg overflow-hidden">
      <button
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center justify-between px-4 py-2.5 bg-gray-50 hover:bg-gray-100 text-sm font-medium text-gray-700"
      >
        <span>🤖 Generar Prompt para Affinity MCP</span>
        <span className="text-gray-400">{open ? "▲" : "▼"}</span>
      </button>

      {open && (
        <div className="p-4 space-y-3 bg-white">
          {/* Controls */}
          <div className="flex flex-wrap gap-3 items-end">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Página</label>
              <select
                value={selectedPageIndex}
                onChange={(e) => setSelectedPageIndex(Number(e.target.value))}
                className="text-sm border border-gray-300 rounded px-2 py-1"
                disabled={mode === "full"}
              >
                {pages.map((p, i) => (
                  <option key={p.id} value={i}>
                    Página {p.noPagina} ({p.pautas.length} pautas)
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Alcance</label>
              <div className="flex gap-1">
                {(["single", ...(edition.facingPages ? ["spread"] : []), "full"] as PromptMode[]).map((m) => (
                  <button
                    key={m}
                    onClick={() => setMode(m)}
                    className={`text-xs px-3 py-1 rounded border transition-colors ${
                      mode === m
                        ? "bg-blue-600 text-white border-blue-600"
                        : "bg-white text-gray-600 border-gray-300 hover:border-blue-400"
                    }`}
                  >
                    {{ single: "Página", spread: "Spread", full: "Edición completa" }[m]}
                  </button>
                ))}
              </div>
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Numeración</label>
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
            </div>

            <button
              onClick={handleCopy}
              className="ml-auto text-sm px-4 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
            >
              {copied ? "✓ Copiado" : "Copiar prompt"}
            </button>
          </div>

          {/* Prompt preview */}
          <pre className="p-3 bg-gray-900 text-green-300 text-xs rounded overflow-auto max-h-72 whitespace-pre-wrap">
            {prompt}
          </pre>
        </div>
      )}
    </div>
  );
}
