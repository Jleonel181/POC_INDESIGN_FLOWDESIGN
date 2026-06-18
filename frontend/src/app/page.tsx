"use client";

import { useDashboard } from "@/modules/dashboard/presentation/hooks/useDashboard";
import { StatCard } from "@/modules/dashboard/presentation/components/StatCard";
import { EditionCard } from "@/modules/dashboard/presentation/components/EditionCard";

export default function DashboardPage() {
  const { editions, loading, error, totalPages, facingCount } = useDashboard();

  return (
    <div className="p-6 space-y-8">
      <div>
        <h1 className="text-xl font-semibold text-gray-800">Dashboard</h1>
        <p className="text-sm text-gray-500">Resumen general de ediciones</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <StatCard icon="auto_stories" label="Ediciones" value={loading ? "—" : editions.length} />
        <StatCard icon="description" label="Total páginas" value={loading ? "—" : totalPages} />
        <StatCard icon="menu_book" label="Facing pages" value={loading ? "—" : facingCount} />
        <StatCard icon="grid_view" label="Sin facing" value={loading ? "—" : editions.length - facingCount} />
      </div>

      {/* Editions grid */}
      <div>
        <h2 className="text-sm font-semibold text-gray-600 mb-3">Ediciones disponibles</h2>

        {loading && (
          <p className="text-sm text-gray-400">Cargando ediciones...</p>
        )}

        {error && (
          <p className="text-sm text-red-500">{error}</p>
        )}

        {!loading && !error && editions.length === 0 && (
          <p className="text-sm text-gray-400">No se encontraron ediciones.</p>
        )}

        {!loading && editions.length > 0 && (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
            {editions.map((edition) => (
              <EditionCard key={edition.id} edition={edition} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
