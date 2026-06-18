import { DiagramacionDTO } from "../dto/DiagramacionDTO";
import { SaveLayoutDTO } from "../dto/SaveLayoutDTO";

export interface DiagramacionRepository {
  getLayoutByEditionId(editionId: number): Promise<DiagramacionDTO>;
  getAllLayouts(): Promise<DiagramacionDTO[]>;
  //saveLayout(editionId: number, layout: SaveLayoutDTO): Promise<void>;
}
