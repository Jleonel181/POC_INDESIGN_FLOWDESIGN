import { DiagramacionRepository } from "../ports/DiagramacionRepository";
import { DiagramacionDTO } from "../dto/DiagramacionDTO";

export class GetDiagramacionByEdition {
  constructor(private readonly repository: DiagramacionRepository) {}

  async execute(editionId: number): Promise<DiagramacionDTO> {
    return this.repository.getLayoutByEditionId(editionId);
  }
}
