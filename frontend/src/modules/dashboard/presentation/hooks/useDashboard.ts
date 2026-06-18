"use client";

import { useEffect, useState } from "react";
import { DiagramacionDTO } from "@/modules/diagramacion/application/dto/DiagramacionDTO";
import { HttpDiagramacionRepository } from "@/modules/diagramacion/infraestructure/repositories/HttpDiagramacionRepository";
import { GetAllLayouts } from "@/modules/diagramacion/application/use-cases/GetAllLayouts";

const repository = new HttpDiagramacionRepository();
const getAllLayouts = new GetAllLayouts(repository);

interface DashboardState {
  editions: DiagramacionDTO["edition"][];
  loading: boolean;
  error: string | null;
}

export function useDashboard() {
  const [state, setState] = useState<DashboardState>({
    editions: [],
    loading: true,
    error: null,
  });

  useEffect(() => {
    getAllLayouts
      .execute()
      .then((dtos) => {
        setState({ editions: dtos.map((d) => d.edition), loading: false, error: null });
      })
      .catch(() => {
        setState({ editions: [], loading: false, error: "Error al cargar ediciones" });
      });
  }, []);

  const totalPages  = state.editions.reduce((acc, e) => acc + e.no_paginas, 0);
  const facingCount = state.editions.filter((e) => e.facing_pages).length;

  return { ...state, totalPages, facingCount };
}
