import { DiagramacionRepository } from "../ports/DiagramacionRepository";
import { SaveLayoutDTO } from "../dto/SaveLayoutDTO";

export class SaveDiagramacionLayout {
  constructor(private readonly repository: DiagramacionRepository) {}

  async execute(editionId: number, layout: SaveLayoutDTO): Promise<void> {
    return this.repository.saveLayout(editionId, layout);
  }
}
