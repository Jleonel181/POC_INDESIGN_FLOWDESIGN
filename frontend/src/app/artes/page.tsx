"use client";

import { useState, useEffect, useRef } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001/api";

/** Pauta tal como viene del ESB (sistema de ventas). */
interface VentasAd {
  fid: string;
  customer: string;
  product: string;
  slogan: string;
  coverDate: string;
  formatFid: string;
  cuadrosAncho: number;
  cuadrosAlto: number;
  widthMm: number;
  heightMm: number;
  colorFid: string;
  placementComment: string;
  statusFid: string;
  printSystemFid: string;
  regionalName: string;
  productCategory: string;
  observations: string | null;
}

/** Pauta local (DB FlowDesign) con su imagen si la tiene. */
interface LocalPauta {
  id: number;
  descripcion_pauta: string;
  content_type: "text" | "image";
  image_base64: string | null;
  cover_date: string | null;
}

export default function ArtesPage() {
  const [ads, setAds] = useState<VentasAd[]>([]);
  const [localPautas, setLocalPautas] = useState<LocalPauta[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState<string | null>(null);
  const [filterDate, setFilterDate] = useState(() => new Date().toISOString().split("T")[0]);

  /** Cargar pautas de ventas (ESB) por fecha. */
  const loadAds = async () => {
    if (!filterDate) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/ventas/ads?date=${filterDate}`);
      const data = await res.json();
      setAds(data.ads || []);
    } catch {
      setAds([]);
    } finally {
      setLoading(false);
    }
  };

  /** Cargar pautas locales por cover_date para saber cuáles ya tienen imagen. */
  const loadLocal = async () => {
    try {
      const res = await fetch(`${API_BASE}/pautas?date=${filterDate}`);
      const data = await res.json();
      setLocalPautas(Array.isArray(data) ? data : []);
    } catch {
      setLocalPautas([]);
    }
  };

  useEffect(() => {
    if (filterDate) {
      loadAds();
      loadLocal();
    }
  }, [filterDate]);

  /** Busca si hay una pauta local con imagen para este anuncio (por descripción). */
  const findLocalImage = (ad: VentasAd): LocalPauta | undefined => {
    const desc = [ad.customer, ad.product, ad.formatFid].filter(Boolean).join(" – ") || `Pauta ${ad.fid}`;
    return localPautas.find(
      (p) => p.descripcion_pauta === desc && p.content_type === "image"
    );
  };

  /** Sube imagen a la pauta local por su ID directo. Si no existe localmente, importa primero. */
  const handleUpload = async (ad: VentasAd, file: File) => {
    const desc = [ad.customer, ad.product, ad.formatFid].filter(Boolean).join(" – ") || `Pauta ${ad.fid}`;
    setUploading(ad.fid);

    try {
      // Buscar la pauta local por descripción
      let local = localPautas.find((p) => p.descripcion_pauta === desc);

      // Si no existe localmente, importar las pautas de esta fecha
      if (!local) {
        await fetch(`${API_BASE}/ventas/import`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ date: filterDate }),
        });
        const res = await fetch(`${API_BASE}/pautas?date=${filterDate}`);
        const data = await res.json();
        const freshLocal: LocalPauta[] = Array.isArray(data) ? data : [];
        setLocalPautas(freshLocal);
        local = freshLocal.find((p) => p.descripcion_pauta === desc);
      }

      if (!local) {
        alert("No se pudo asociar la pauta. Intenta de nuevo.");
        return;
      }

      // Subir imagen por ID específico
      const formData = new FormData();
      formData.append("file", file);

      const res = await fetch(`${API_BASE}/pautas/${local.id}/image`, {
        method: "PUT",
        body: formData,
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: "Error desconocido" }));
        alert(`Error: ${err.error}`);
        return;
      }

      await loadLocal();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Error al subir imagen");
    } finally {
      setUploading(null);
    }
  };

  const handleRemove = async (localId: number) => {
    if (!confirm("¿Quitar la imagen de esta pauta?")) return;
    try {
      await fetch(`${API_BASE}/pautas/${localId}/image`, { method: "DELETE" });
      await loadLocal();
    } catch {
      /* ignore */
    }
  };

  return (
    <div className="max-w-5xl mx-auto p-8">
      <h1 className="text-2xl font-bold text-gray-800 mb-2">Gestión de Artes</h1>
      <p className="text-sm text-gray-500 mb-6">
        Pautas del sistema de ventas. Asigna el arte final (imagen) a cada una.
        Al exportar IDML, las que tengan imagen se generan como marco de imagen.
      </p>

      {/* Filtro por fecha de publicación */}
      <div className="flex items-center gap-3 mb-6">
        <label className="text-sm text-gray-600 font-medium">Fecha de publicación:</label>
        <input
          type="date"
          value={filterDate}
          onChange={(e) => setFilterDate(e.target.value)}
          className="border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <span className="ml-auto text-xs text-gray-400">
          {ads.length} pauta{ads.length !== 1 ? "s" : ""} en ventas
        </span>
      </div>

      {loading ? (
        <p className="text-gray-500 py-8 text-center">Consultando ventas...</p>
      ) : ads.length === 0 ? (
        <p className="text-gray-400 text-center py-12">
          No hay pautas en ventas para esta fecha.
        </p>
      ) : (
        <div className="grid gap-3">
          {ads.map((ad) => (
            <AdArteCard
              key={ad.fid}
              ad={ad}
              localPauta={findLocalImage(ad)}
              uploading={uploading === ad.fid}
              onUpload={(file) => handleUpload(ad, file)}
              onRemove={(localId) => handleRemove(localId)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function AdArteCard({
  ad,
  localPauta,
  uploading,
  onUpload,
  onRemove,
}: {
  ad: VentasAd;
  localPauta: LocalPauta | undefined;
  uploading: boolean;
  onUpload: (file: File) => void;
  onRemove: (localId: number) => void;
}) {
  const fileRef = useRef<HTMLInputElement>(null);
  const hasImage = !!localPauta;
  const desc = [ad.customer, ad.product, ad.formatFid].filter(Boolean).join(" – ");

  return (
    <div className="flex items-center gap-4 border border-gray-200 rounded-lg p-4 bg-white">
      {/* Thumbnail */}
      <div className="w-16 h-16 flex-shrink-0 bg-gray-100 rounded border border-gray-200 flex items-center justify-center overflow-hidden">
        {hasImage ? (
          <img
            src={`${API_BASE}/pautas/${localPauta.id}/image`}
            alt={desc}
            className="w-full h-full object-cover"
          />
        ) : (
          <span className="text-[10px] text-gray-400 text-center px-1">Sin arte</span>
        )}
      </div>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <p className="font-medium text-gray-800 truncate text-sm">{desc || `FID ${ad.fid}`}</p>
        <p className="text-xs text-gray-500">
          {ad.cuadrosAncho}×{ad.cuadrosAlto} cuadros · {ad.colorFid} · {ad.statusFid}
        </p>
        {ad.observations && (
          <p className="text-xs text-gray-400 truncate">{ad.observations}</p>
        )}
        <p className={`text-xs mt-0.5 ${hasImage ? "text-green-600" : "text-amber-600"}`}>
          {hasImage ? "Arte asignado" : "Sin arte — dummy en IDML"}
        </p>
      </div>

      {/* Actions */}
      <div className="flex gap-2 flex-shrink-0">
        {hasImage && (
          <button
            onClick={() => onRemove(localPauta.id)}
            className="text-xs px-3 py-1.5 border border-red-200 text-red-600 rounded hover:bg-red-50 transition-colors"
          >
            Quitar
          </button>
        )}
        <button
          onClick={() => fileRef.current?.click()}
          disabled={uploading}
          className="text-xs px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
        >
          {uploading ? "Subiendo..." : hasImage ? "Cambiar" : "Subir arte"}
        </button>
        <input
          ref={fileRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (file) onUpload(file);
            e.target.value = "";
          }}
        />
      </div>
    </div>
  );
}
