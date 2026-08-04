"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/shared/infraestructure/http/apiClient";

interface PautaItem {
  id: number;
  descripcion_pauta: string;
  cuadros_alto: number;
  cuadros_ancho: number;
  paginaId: number | null;
}

export default function PautasPage() {
  const [pautas, setPautas] = useState<PautaItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form
  const [descripcion, setDescripcion] = useState("");
  const [cuadrosAncho, setCuadrosAncho] = useState(2);
  const [cuadrosAlto, setCuadrosAlto] = useState(2);

  const loadPautas = async () => {
    try {
      const data = await apiClient.get<PautaItem[]>("/pautas/unassigned");
      setPautas(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error al cargar");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadPautas(); }, []);

  const handleSubmit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      await apiClient.post("/pautas", {
        descripcion_pauta: descripcion,
        cuadros_alto: cuadrosAlto,
        cuadros_ancho: cuadrosAncho,
      });
      setDescripcion("");
      setCuadrosAncho(2);
      setCuadrosAlto(2);
      await loadPautas();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setSubmitting(false);
    }
  };

  // Preview: muestra la pauta como un bloque en una grilla genérica de 6×8
  const previewCols = 6;
  const previewRows = 8;
  const cellSize = 28;
  const previewW = previewCols * cellSize;
  const previewH = previewRows * cellSize;

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-gray-800">Biblioteca de Pautas</h1>
        <p className="text-sm text-gray-500">Crea pautas que luego podrás asignar a las páginas de una edición</p>
      </div>

      <div className="flex gap-8">
        {/* Formulario */}
        <div className="flex-1 space-y-4">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-gray-500">Descripción</label>
            <input
              type="text"
              value={descripcion}
              onChange={(e) => setDescripcion(e.target.value)}
              placeholder="Ej: Nota principal, Publicidad media..."
              className="text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Ancho (columnas)</label>
              <input
                type="number"
                min={1}
                max={12}
                value={cuadrosAncho}
                onChange={(e) => setCuadrosAncho(Number(e.target.value))}
                className="text-sm border border-gray-300 rounded px-2 py-1.5"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Alto (filas)</label>
              <input
                type="number"
                min={1}
                max={12}
                value={cuadrosAlto}
                onChange={(e) => setCuadrosAlto(Number(e.target.value))}
                className="text-sm border border-gray-300 rounded px-2 py-1.5"
              />
            </div>
          </div>

          {error && <p className="text-xs text-red-500">{error}</p>}

          <button
            onClick={handleSubmit}
            disabled={submitting || !descripcion.trim() || cuadrosAncho <= 0 || cuadrosAlto <= 0}
            className="w-full text-sm px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {submitting ? "Guardando..." : "Agregar a biblioteca"}
          </button>
        </div>

        {/* Preview */}
        <div className="flex-shrink-0">
          <p className="text-xs text-gray-500 mb-2">Previsualización (proporcional)</p>
          <div
            className="relative border border-gray-200 bg-gray-50"
            style={{ width: previewW, height: previewH }}
          >
            <svg width={previewW} height={previewH}>
              {/* Grilla de referencia */}
              {Array.from({ length: previewCols + 1 }, (_, i) => (
                <line
                  key={`v-${i}`}
                  x1={i * cellSize} y1={0} x2={i * cellSize} y2={previewH}
                  stroke="#e5e7eb" strokeWidth={0.5}
                />
              ))}
              {Array.from({ length: previewRows + 1 }, (_, i) => (
                <line
                  key={`h-${i}`}
                  x1={0} y1={i * cellSize} x2={previewW} y2={i * cellSize}
                  stroke="#e5e7eb" strokeWidth={0.5}
                />
              ))}

              {/* Pauta nueva */}
              <rect
                x={0}
                y={0}
                width={Math.min(cuadrosAncho, previewCols) * cellSize}
                height={Math.min(cuadrosAlto, previewRows) * cellSize}
                fill="rgba(59, 130, 246, 0.15)"
                stroke="#3b82f6"
                strokeWidth={1.5}
                rx={2}
              />
              <text
                x={Math.min(cuadrosAncho, previewCols) * cellSize / 2}
                y={Math.min(cuadrosAlto, previewRows) * cellSize / 2 + 4}
                textAnchor="middle"
                className="text-[10px] fill-blue-600 font-medium"
              >
                {cuadrosAncho}×{cuadrosAlto}
              </text>
            </svg>
          </div>
          <p className="text-[10px] text-gray-400 mt-1 text-center">
            {descripcion || "Sin nombre"} — {cuadrosAncho} col × {cuadrosAlto} filas
          </p>
        </div>
      </div>

      {/* Lista de pautas en biblioteca */}
      <div>
        <h2 className="text-sm font-semibold text-gray-600 mb-3">Pautas disponibles (sin asignar)</h2>

        {loading && <p className="text-sm text-gray-400">Cargando...</p>}

        {!loading && pautas.length === 0 && (
          <p className="text-sm text-gray-400">No hay pautas en la biblioteca. Crea una arriba.</p>
        )}

        {!loading && pautas.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
            {pautas.map((p) => (
              <div key={p.id} className="border border-gray-200 rounded-lg p-3 bg-white">
                <div className="flex items-center gap-2 mb-2">
                  <div
                    className="bg-blue-100 border border-blue-300 rounded flex items-center justify-center text-[9px] text-blue-600 font-mono"
                    style={{ width: 32, height: 32 }}
                  >
                    {p.cuadros_ancho}×{p.cuadros_alto}
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-gray-700 truncate">{p.descripcion_pauta}</p>
                    <p className="text-[10px] text-gray-400">{p.cuadros_ancho} col × {p.cuadros_alto} filas</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
