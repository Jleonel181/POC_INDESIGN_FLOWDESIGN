import { DomainError } from "../../../../shared/domain/DomainError";
import { LayoutContract } from "./LayoutContract";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isFiniteNumber(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value);
}

export function assertLayoutContract(payload: unknown): asserts payload is LayoutContract {
  if (!isRecord(payload)) {
    throw new DomainError("Layout inválido: payload debe ser un objeto.");
  }

  const metadata = payload.metadata;
  const edition = payload.edition;
  const pages = payload.pages;

  if (!isRecord(metadata)) {
    throw new DomainError("Layout inválido: metadata es obligatorio.");
  }

  if (
    typeof metadata.version !== "string" ||
    metadata.unit !== "mm" ||
    metadata.coordinateSystem !== "grid" ||
    metadata.origin !== "top-left"
  ) {
    throw new DomainError("Layout inválido: metadata no cumple el contrato.");
  }

  if (!isRecord(edition)) {
    throw new DomainError("Layout inválido: edition es obligatorio.");
  }

  const editionNumericKeys = [
    "id",
    "no_paginas",
    "ancho_mm",
    "alto_mm",
    "cuadros_ancho",
    "cuadros_alto"
  ] as const;

  for (const key of editionNumericKeys) {
    if (!isFiniteNumber(edition[key])) {
      throw new DomainError(`Layout inválido: edition.${key} debe ser numérico.`);
    }
  }

  if (!Array.isArray(pages)) {
    throw new DomainError("Layout inválido: pages debe ser un arreglo.");
  }

  for (const [pageIndex, page] of pages.entries()) {
    if (!isRecord(page)) {
      throw new DomainError(`Layout inválido: pages[${pageIndex}] debe ser un objeto.`);
    }

    if (!isFiniteNumber(page.id) || !isFiniteNumber(page.no_pagina)) {
      throw new DomainError(`Layout inválido: pages[${pageIndex}] no cumple campos numéricos.`);
    }

    if (!Array.isArray(page.pautas)) {
      throw new DomainError(`Layout inválido: pages[${pageIndex}].pautas debe ser un arreglo.`);
    }

    for (const [pautaIndex, pauta] of page.pautas.entries()) {
      if (!isRecord(pauta)) {
        throw new DomainError(
          `Layout inválido: pages[${pageIndex}].pautas[${pautaIndex}] debe ser un objeto.`
        );
      }

      if (
        !isFiniteNumber(pauta.id) ||
        typeof pauta.descripcion_pauta !== "string" ||
        !isFiniteNumber(pauta.cuadros_alto) ||
        !isFiniteNumber(pauta.cuadros_ancho) ||
        !isFiniteNumber(pauta.ubicacion_cuadros_x) ||
        !isFiniteNumber(pauta.ubicacion_cuadros_y)
      ) {
        throw new DomainError(
          `Layout inválido: pages[${pageIndex}].pautas[${pautaIndex}] no cumple campos requeridos.`
        );
      }

      if (!isRecord(pauta.indesignBounds)) {
        throw new DomainError(
          `Layout inválido: pages[${pageIndex}].pautas[${pautaIndex}].indesignBounds es obligatorio.`
        );
      }

      const { topMm, leftMm, bottomMm, rightMm } = pauta.indesignBounds;
      if (
        !isFiniteNumber(topMm) ||
        !isFiniteNumber(leftMm) ||
        !isFiniteNumber(bottomMm) ||
        !isFiniteNumber(rightMm)
      ) {
        throw new DomainError(
          `Layout inválido: pages[${pageIndex}].pautas[${pautaIndex}].indesignBounds debe tener números válidos.`
        );
      }
    }
  }
}
