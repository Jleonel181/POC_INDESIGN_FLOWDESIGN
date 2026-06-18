import { DiagramacionRepository } from "../../application/ports/DiagramacionRepository";
import { DiagramacionDTO } from "../../application/dto/DiagramacionDTO";
import { SaveLayoutDTO } from "../../application/dto/SaveLayoutDTO";
import { apiClient } from "../../../../shared/infraestructure/http/apiClient";

export class HttpDiagramacionRepository implements DiagramacionRepository {
  async getLayoutByEditionId(editionId: number): Promise<DiagramacionDTO> {
    return apiClient.get<DiagramacionDTO>(`/layout/${editionId}`);
  }

  async getAllLayouts(): Promise<DiagramacionDTO[]> {
    return apiClient.get<DiagramacionDTO[]>(`/layout`);
  }
}
