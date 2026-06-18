import { Pauta } from "../../domain/entities/Pauta";
import { Page } from "../../domain/entities/Page";
import { LayoutCollisionService } from "../../domain/services/LayoutCollisionService";
import { GridPosition } from "../../domain/value-objects/GridPosition";
import { GridSize } from "../../domain/value-objects/GridSize";

interface AddPautaInput {
  page: Page;
  pauta: Pauta;
  gridColumns: number;
  gridRows: number;
}

export class AddPautaToPage {
  private collisionService = new LayoutCollisionService();

  execute(input: AddPautaInput): { success: boolean; error?: string } {
    const { page, pauta, gridColumns, gridRows } = input;
    const pos = pauta.getPosition();
    const size = pauta.getSize();

    // Validate bounds
    if (pos.x + size.width > gridColumns || pos.y + size.height > gridRows) {
      return { success: false, error: "La pauta excede los límites de la grilla" };
    }

    // Validate collisions
    const collisions = this.collisionService.findCollisions([...page.pautas, pauta]);
    if (collisions.length > 0) {
      return { success: false, error: "La pauta colisiona con otra existente" };
    }

    return { success: true };
  }
}
