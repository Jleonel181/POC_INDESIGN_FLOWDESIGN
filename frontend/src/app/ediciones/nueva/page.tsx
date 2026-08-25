"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/shared/infraestructure/http/apiClient";

type Preset = "nuestro-diario" | "custom";

const PRESETS = {
  "nuestro-diario": {
    label: "Nuestro Diario",
    noPaginas: 28,
    anchoMm: 272.99,
    altoMm: 336.55,
    cuadrosAncho: 5,
    cuadrosAlto: 8,
    facingPages: true,
    margenSup: 9.53,
    margenInf: 9.53,
    margenIzq: 9.53,
    margenDer: 9.53,
    columns: 1,
  },
} as const;

export default function NuevaEdicionPage() {
  const router = useRouter();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [preset, setPreset] = useState<Preset>("nuestro-diario");

  const [noPaginas, setNoPaginas] = useState<number>(PRESETS["nuestro-diario"].noPaginas);
  const [anchoMm, setAnchoMm] = useState<number>(PRESETS["nuestro-diario"].anchoMm);
  const [altoMm, setAltoMm] = useState<number>(PRESETS["nuestro-diario"].altoMm);
  const [cuadrosAncho, setCuadrosAncho] = useState<number>(PRESETS["nuestro-diario"].cuadrosAncho);
  const [cuadrosAlto, setCuadrosAlto] = useState<number>(PRESETS["nuestro-diario"].cuadrosAlto);
  const [facingPages, setFacingPages] = useState<boolean>(PRESETS["nuestro-diario"].facingPages);
  const [margenSup, setMargenSup] = useState<number>(PRESETS["nuestro-diario"].margenSup);
  const [margenInf, setMargenInf] = useState<number>(PRESETS["nuestro-diario"].margenInf);
  const [margenIzq, setMargenIzq] = useState<number>(PRESETS["nuestro-diario"].margenIzq);
  const [margenDer, setMargenDer] = useState<number>(PRESETS["nuestro-diario"].margenDer);

  const applyPreset = (key: Preset) => {
    setPreset(key);
    if (key === "custom") return;
    const p = PRESETS[key];
    setNoPaginas(p.noPaginas);
    setAnchoMm(p.anchoMm);
    setAltoMm(p.altoMm);
    setCuadrosAncho(p.cuadrosAncho);
    setCuadrosAlto(p.cuadrosAlto);
    setFacingPages(p.facingPages);
    setMargenSup(p.margenSup);
    setMargenInf(p.margenInf);
    setMargenIzq(p.margenIzq);
    setMargenDer(p.margenDer);
  };

  const handleSubmit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      const result = await apiClient.post<{ edition: { id: number } }>("/editions", {
        no_paginas: noPaginas,
        ancho_mm: anchoMm,
        alto_mm: altoMm,
        cuadros_ancho: cuadrosAncho,
        cuadros_alto: cuadrosAlto,
        facing_pages: facingPages,
        margen_superior_mm: margenSup,
        margen_inferior_mm: margenInf,
        margen_izquierdo_mm: margenIzq,
        margen_derecho_mm: margenDer,
      });
      router.push(`/diagramador/${result.edition.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error desconocido");
    } finally {
      setSubmitting(false);
    }
  };

  // --- Canvas de previsualización ---
  const canvasWidth = 260;
  const scale = canvasWidth / anchoMm;
  const canvasHeight = altoMm * scale;

  const mTop = margenSup * scale;
  const mBottom = margenInf * scale;
  const mLeft = margenIzq * scale;
  const mRight = margenDer * scale;

  const contentW = canvasWidth - mLeft - mRight;
  const contentH = canvasHeight - mTop - mBottom;

  const cellW = cuadrosAncho > 0 ? contentW / cuadrosAncho : contentW;
  const cellH = cuadrosAlto > 0 ? contentH / cuadrosAlto : contentH;

  const isPreset = preset !== "custom";

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-gray-800">Crear nueva edición</h1>
        <p className="text-sm text-gray-500">Selecciona un formato o personaliza las dimensiones</p>
      </div>

      {/* Selector de preset */}
      <div className="flex gap-3">
        <button
          onClick={() => applyPreset("nuestro-diario")}
          className={`px-4 py-2 text-sm rounded border transition-colors ${
            preset === "nuestro-diario"
              ? "bg-blue-600 text-white border-blue-600"
              : "bg-white text-gray-700 border-gray-300 hover:border-blue-400"
          }`}
        >
          Nuestro Diario
        </button>
        <button
          onClick={() => applyPreset("custom")}
          className={`px-4 py-2 text-sm rounded border transition-colors ${
            preset === "custom"
              ? "bg-blue-600 text-white border-blue-600"
              : "bg-white text-gray-700 border-gray-300 hover:border-blue-400"
          }`}
        >
          Personalizado
        </button>
      </div>

      <div className="flex gap-8">
        {/* Formulario */}
        <div className="flex-1 space-y-4">
          {isPreset && (
            <div className="bg-blue-50 border border-blue-200 rounded p-3">
              <p className="text-xs text-blue-700">
                Formato Nuestro Diario: 272.99×336.55 mm, grilla 5×8, facing pages, márgenes 9.53 mm
              </p>
            </div>
          )}

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Ancho (mm)</label>
              <input
                type="number"
                min={50}
                step={0.01}
                value={anchoMm}
                onChange={(e) => { setAnchoMm(Number(e.target.value)); setPreset("custom"); }}
                disabled={isPreset}
                className="text-sm border border-gray-300 rounded px-2 py-1.5 disabled:bg-gray-50 disabled:text-gray-500"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Alto (mm)</label>
              <input
                type="number"
                min={50}
                step={0.01}
                value={altoMm}
                onChange={(e) => { setAltoMm(Number(e.target.value)); setPreset("custom"); }}
                disabled={isPreset}
                className="text-sm border border-gray-300 rounded px-2 py-1.5 disabled:bg-gray-50 disabled:text-gray-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Columnas (grilla)</label>
              <input
                type="number"
                min={1}
                value={cuadrosAncho}
                onChange={(e) => { setCuadrosAncho(Number(e.target.value)); setPreset("custom"); }}
                disabled={isPreset}
                className="text-sm border border-gray-300 rounded px-2 py-1.5 disabled:bg-gray-50 disabled:text-gray-500"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500">Filas (grilla)</label>
              <input
                type="number"
                min={1}
                value={cuadrosAlto}
                onChange={(e) => { setCuadrosAlto(Number(e.target.value)); setPreset("custom"); }}
                disabled={isPreset}
                className="text-sm border border-gray-300 rounded px-2 py-1.5 disabled:bg-gray-50 disabled:text-gray-500"
              />
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-gray-500">Páginas</label>
            <input
              type="number"
              min={1}
              value={noPaginas}
              onChange={(e) => setNoPaginas(Number(e.target.value))}
              className="text-sm border border-gray-300 rounded px-2 py-1.5"
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="facing"
              checked={facingPages}
              onChange={(e) => { setFacingPages(e.target.checked); setPreset("custom"); }}
              disabled={isPreset}
              className="rounded border-gray-300"
            />
            <label htmlFor="facing" className="text-sm text-gray-600">Facing pages (spreads)</label>
          </div>

          <fieldset className="border border-gray-200 rounded p-3 space-y-2">
            <legend className="text-xs text-gray-500 px-1">Márgenes (mm)</legend>
            <div className="grid grid-cols-2 gap-2">
              <div className="flex flex-col gap-1">
                <label className="text-xs text-gray-400">Superior</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  value={margenSup}
                  onChange={(e) => { setMargenSup(Number(e.target.value)); setPreset("custom"); }}
                  disabled={isPreset}
                  className="text-sm border border-gray-300 rounded px-2 py-1 disabled:bg-gray-50 disabled:text-gray-500"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs text-gray-400">Inferior</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  value={margenInf}
                  onChange={(e) => { setMargenInf(Number(e.target.value)); setPreset("custom"); }}
                  disabled={isPreset}
                  className="text-sm border border-gray-300 rounded px-2 py-1 disabled:bg-gray-50 disabled:text-gray-500"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs text-gray-400">Izquierdo</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  value={margenIzq}
                  onChange={(e) => { setMargenIzq(Number(e.target.value)); setPreset("custom"); }}
                  disabled={isPreset}
                  className="text-sm border border-gray-300 rounded px-2 py-1 disabled:bg-gray-50 disabled:text-gray-500"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="text-xs text-gray-400">Derecho</label>
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  value={margenDer}
                  onChange={(e) => { setMargenDer(Number(e.target.value)); setPreset("custom"); }}
                  disabled={isPreset}
                  className="text-sm border border-gray-300 rounded px-2 py-1 disabled:bg-gray-50 disabled:text-gray-500"
                />
              </div>
            </div>
          </fieldset>

          {error && <p className="text-xs text-red-500">{error}</p>}

          <button
            onClick={handleSubmit}
            disabled={submitting || anchoMm <= 0 || altoMm <= 0 || noPaginas <= 0}
            className="w-full text-sm px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {submitting ? "Creando..." : "Crear edición"}
          </button>
        </div>

        {/* Canvas de previsualización */}
        <div className="flex-shrink-0">
          <p className="text-xs text-gray-500 mb-2">Previsualización de la página</p>
          <div
            className="relative border border-gray-300 bg-white shadow-sm"
            style={{ width: canvasWidth, height: canvasHeight }}
          >
            <svg
              className="absolute inset-0"
              width={canvasWidth}
              height={canvasHeight}
            >
              {/* Márgenes */}
              <rect
                x={mLeft}
                y={mTop}
                width={contentW}
                height={contentH}
                fill="none"
                stroke="#097CC8"
                strokeWidth={0.75}
                strokeDasharray="4 2"
              />

              {/* Grilla vertical */}
              {Array.from({ length: cuadrosAncho + 1 }, (_, i) => (
                <line
                  key={`v-${i}`}
                  x1={mLeft + i * cellW}
                  y1={mTop}
                  x2={mLeft + i * cellW}
                  y2={mTop + contentH}
                  stroke="#097CC8"
                  strokeWidth={i === 0 || i === cuadrosAncho ? 0 : 0.4}
                  opacity={0.5}
                />
              ))}

              {/* Grilla horizontal */}
              {Array.from({ length: cuadrosAlto + 1 }, (_, i) => (
                <line
                  key={`h-${i}`}
                  x1={mLeft}
                  y1={mTop + i * cellH}
                  x2={mLeft + contentW}
                  y2={mTop + i * cellH}
                  stroke="#097CC8"
                  strokeWidth={i === 0 || i === cuadrosAlto ? 0 : 0.4}
                  opacity={0.5}
                />
              ))}

              {/* Labels */}
              <text x={canvasWidth / 2} y={canvasHeight + 16} textAnchor="middle" className="text-[9px] fill-gray-400">
                {anchoMm} mm
              </text>
            </svg>
          </div>
          <div className="mt-2 text-center">
            <span className="text-[10px] text-gray-400">
              {anchoMm}×{altoMm} mm | {cuadrosAncho}×{cuadrosAlto} grilla | {noPaginas} págs
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
