import { DiagramacionRepository } from "../ports/DiagramacionRepository";
import { DiagramacionDTO } from "../dto/DiagramacionDTO";

export class GetAllLayouts {
  constructor(private readonly repository: DiagramacionRepository) {}

  execute(): Promise<DiagramacionDTO[]> {
    return this.repository.getAllLayouts();
  }
}
