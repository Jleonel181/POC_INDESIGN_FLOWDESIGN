import { Pauta } from '../entities/Pauta';

export class LayoutCollisionService {
  hasCollision(pautaA: Pauta, pautaB: Pauta): boolean {
    const posA = pautaA.getPosition();
    const sizeA = pautaA.getSize();
    const posB = pautaB.getPosition();
    const sizeB = pautaB.getSize();

    const aRight = posA.x + sizeA.width;
    const aBottom = posA.y + sizeA.height;
    const bRight = posB.x + sizeB.width;
    const bBottom = posB.y + sizeB.height;

    return posA.x < bRight && aRight > posB.x && posA.y < bBottom && aBottom > posB.y;
  }

  findCollisions(pautas: Pauta[]): [Pauta, Pauta][] {
    const collisions: [Pauta, Pauta][] = [];
    for (let i = 0; i < pautas.length; i++) {
      for (let j = i + 1; j < pautas.length; j++) {
        if (this.hasCollision(pautas[i], pautas[j])) {
          collisions.push([pautas[i], pautas[j]]);
        }
      }
    }
    return collisions;
  }
}
