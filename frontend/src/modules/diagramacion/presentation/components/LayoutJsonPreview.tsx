"use client";

import { DiagramacionDTO } from "../../application/dto/DiagramacionDTO";

interface LayoutJsonPreviewProps {
  dto: DiagramacionDTO | null;
}

export function LayoutJsonPreview({ dto }: LayoutJsonPreviewProps) {
  if (!dto) return null;

  return (
    <details className="mt-4">
      <summary className="text-sm text-gray-500 cursor-pointer">Ver JSON del Layout</summary>
      <pre className="mt-2 p-3 bg-gray-900 text-green-400 text-xs rounded overflow-auto max-h-80">
        {JSON.stringify(dto, null, 2)}
      </pre>
    </details>
  );
}
