"use client";

import { useState } from "react";
import { apiClient } from "@/shared/infraestructure/http/apiClient";

interface PautaItem {
  id: number;
  descripcion_pauta: string;
  cuadros_alto: number;
  cuadros_ancho: number;
}

export function PautasSidebar() {
  const [pautas, setPautas] = useState<PautaItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [importDate, setImportDate] = useState("");
  const [importing, setImporting] = useState(false);

  const handleImport = async () => {
    if (!importDate) return;
    setImporting(true);
    try {
      const result = await apiClient.post<{ imported: number; pautas: PautaItem[] }>("/ventas/import", { date: importDate });
      setPautas(result.pautas);
    } catch (err) {
      console.error("Error importando pautas:", err);
    } finally {
      setImporting(false);
    }
  };

  const handleDragStart = (e: React.DragEvent, pauta: PautaItem) => {
    e.dataTransfer.setData("application/json", JSON.stringify(pauta));
    // Codificamos el tamaño en un type custom para que sea legible durante dragOver.
    e.dataTransfer.setData(`pauta-size/${pauta.cuadros_ancho}x${pauta.cuadros_alto}`, "");
    e.dataTransfer.effectAllowed = "copy";
  };

  return (
    <div className="w-56 border-r border-gray-200 bg-gray-50 p-3 overflow-y-auto">
      <h3 className="text-xs font-semibold text-gray-500 uppercase mb-2">Biblioteca</h3>

      {/* Importar pautas desde ventas */}
      <div className="mb-3 space-y-1">
        <input
          type="date"
          value={importDate}
          onChange={(e) => setImportDate(e.target.value)}
          className="w-full text-xs border border-gray-300 rounded px-2 py-1"
        />
        <button
          onClick={handleImport}
          disabled={importing || !importDate}
          className="w-full text-xs bg-blue-600 text-white rounded px-2 py-1 disabled:opacity-50 hover:bg-blue-700 transition-colors"
        >
          {importing ? "Importando..." : "Importar de Ventas"}
        </button>
      </div>

      {loading && <p className="text-xs text-gray-400">Cargando...</p>}

      {!loading && pautas.length === 0 && (
        <p className="text-xs text-gray-400">Selecciona una fecha e importa</p>
      )}

      <div className="space-y-1.5">
        {pautas.map((p) => (
          <div
            key={p.id}
            draggable
            onDragStart={(e) => handleDragStart(e, p)}
            className="flex items-center gap-2 p-2 bg-white border border-gray-200 rounded cursor-grab active:cursor-grabbing hover:border-blue-400 transition-colors"
          >
            <div className="w-8 h-8 bg-blue-50 border border-blue-200 rounded flex items-center justify-center text-[9px] text-blue-600 font-mono shrink-0">
              {p.cuadros_ancho}×{p.cuadros_alto}
            </div>
            <span className="text-xs text-gray-700 truncate">{p.descripcion_pauta}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
