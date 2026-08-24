import { GridPosition } from '../value-objects/GridPosition';
import { GridSize } from '../value-objects/GridSize';

export interface InDesignBounds {
  topMm: number;
  leftMm: number;
  bottomMm: number;
  rightMm: number;
}

export class Pauta {
  constructor(
    public readonly id: number,
    public readonly descripcion: string,
    private position: GridPosition,
    private size: GridSize,
    public readonly indesignBounds?: InDesignBounds,
    public readonly contentType: "text" | "image" = "text",
    public readonly imageBase64?: string
  ) {}

  getPosition(): GridPosition {
    return this.position;
  }

  getSize(): GridSize {
    return this.size;
  }

  moveTo(position: GridPosition): void {
    this.position = position;
  }

  resize(size: GridSize): void {
    this.size = size;
  }

  get isImage(): boolean {
    return this.contentType === "image" && !!this.imageBase64;
  }
}
