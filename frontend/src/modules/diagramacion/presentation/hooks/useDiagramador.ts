"use client";

import { useState, useEffect, useCallback } from "react";
import { Edition } from "../../domain/entities/Edition";
import { Page } from "../../domain/entities/Page";
import { Pauta } from "../../domain/entities/Pauta";
import { GridPosition } from "../../domain/value-objects/GridPosition";
import { GridSize } from "../../domain/value-objects/GridSize";
import { LayoutCollisionService } from "../../domain/services/LayoutCollisionService";
import { HttpDiagramacionRepository } from "../../infraestructure/repositories/HttpDiagramacionRepository";
import { DiagramacionMapper } from "../../infraestructure/mappers/DiagramacionMapper";
import { GetDiagramacionByEdition } from "../../application/use-cases/GetDiagramacionByEdition";
import { SaveDiagramacionLayout } from "../../application/use-cases/SaveDiagramacionLayout";
import { DiagramacionDTO } from "../../application/dto/DiagramacionDTO";

interface UseDiagramadorState {
  edition: Edition | null;
  pages: Page[];
  currentPageIndex: number;
  loading: boolean;
  error: string | null;
  rawDTO: DiagramacionDTO | null;
}

const repository = new HttpDiagramacionRepository();
const getDiagramacion = new GetDiagramacionByEdition(repository);
const saveDiagramacion = new SaveDiagramacionLayout(repository);
const collisionService = new LayoutCollisionService();

export function useDiagramador(editionId: number) {
  const [state, setState] = useState<UseDiagramadorState>({
    edition: null,
    pages: [],
    currentPageIndex: 0,
    loading: true,
    error: null,
    rawDTO: null,
  });

  // Fetch layout data
  useEffect(() => {
    async function load() {
      try {
        const dto = await getDiagramacion.execute(editionId);
        const { edition, pages } = DiagramacionMapper.toDomain(dto);
        setState((prev) => ({ ...prev, edition, pages, rawDTO: dto, loading: false }));
      } catch (err) {
        setState((prev) => ({
          ...prev,
          loading: false,
          error: err instanceof Error ? err.message : "Error desconocido",
        }));
      }
    }
    load();
  }, [editionId]);

  const currentPage = state.pages[state.currentPageIndex] ?? null;

  const setCurrentPageIndex = useCallback((index: number) => {
    setState((prev) => ({ ...prev, currentPageIndex: index }));
  }, []);

  const movePauta = useCallback(
    (pautaId: number, newPosition: GridPosition) => {
      setState((prev) => {
        const pages = prev.pages.map((page) => {
          const pautaIndex = page.pautas.findIndex((p) => p.id === pautaId);
          if (pautaIndex === -1) return page;

          const updatedPautas = [...page.pautas];
          const pauta = updatedPautas[pautaIndex];
          pauta.moveTo(newPosition);

          // Validate collision
          const collisions = collisionService.findCollisions(updatedPautas);
          if (collisions.length > 0) return page; // Reject move

          return new Page(page.id, page.noPagina, updatedPautas);
        });
        return { ...prev, pages };
      });
    },
    []
  );

  const resizePauta = useCallback(
    (pautaId: number, newSize: GridSize) => {
      setState((prev) => {
        const pages = prev.pages.map((page) => {
          const pautaIndex = page.pautas.findIndex((p) => p.id === pautaId);
          if (pautaIndex === -1) return page;

          const updatedPautas = [...page.pautas];
          const pauta = updatedPautas[pautaIndex];
          pauta.resize(newSize);

          const collisions = collisionService.findCollisions(updatedPautas);
          if (collisions.length > 0) return page; // Reject resize

          return new Page(page.id, page.noPagina, updatedPautas);
        });
        return { ...prev, pages };
      });
    },
    []
  );

  const saveLayout = useCallback(async () => {
    if (!state.edition) return;

    const items = state.pages.flatMap((page) =>
      page.pautas.map((pauta) => ({
        pautaId: pauta.id,
        pageId: page.id,
        ubicacion_cuadros_x: pauta.getPosition().x,
        ubicacion_cuadros_y: pauta.getPosition().y,
        cuadros_ancho: pauta.getSize().width,
        cuadros_alto: pauta.getSize().height,
      }))
    );

    await saveDiagramacion.execute(editionId, { items });
  }, [state.edition, state.pages, editionId]);

  return {
    edition: state.edition,
    pages: state.pages,
    currentPage,
    currentPageIndex: state.currentPageIndex,
    loading: state.loading,
    error: state.error,
    rawDTO: state.rawDTO,
    setCurrentPageIndex,
    movePauta,
    resizePauta,
    saveLayout,
  };
}
