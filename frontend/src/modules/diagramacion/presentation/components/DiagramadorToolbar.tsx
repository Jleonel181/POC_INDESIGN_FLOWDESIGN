"use client";

interface DiagramadorToolbarProps {
  totalPages: number;
  currentPageIndex: number;
  onPageChange: (index: number) => void;
  onSave: () => void;
}

export function DiagramadorToolbar({ totalPages, currentPageIndex, onPageChange, onSave }: DiagramadorToolbarProps) {
  return (
    <div className="flex items-center gap-4 p-3 bg-gray-50 border border-gray-200 rounded-md">
      <div className="flex items-center gap-2 text-gray-700">
        <button
          disabled={currentPageIndex === 0}
          onClick={() => onPageChange(currentPageIndex - 1)}
          className="px-2 py-1 text-sm border rounded disabled:opacity-40"
        >
          ← Anterior
        </button>
        <span className="text-sm text-gray-900">
          Página {currentPageIndex + 1} de {totalPages}
        </span>
        <button
          disabled={currentPageIndex >= totalPages - 1}
          onClick={() => onPageChange(currentPageIndex + 1)}
          className="px-2 py-1 text-sm border rounded disabled:opacity-40"
        >
          Siguiente →
        </button>
      </div>

      <div className="ml-auto">
        <button onClick={onSave} className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700">
          Guardar Layout
        </button>
      </div>
    </div>
  );
}
