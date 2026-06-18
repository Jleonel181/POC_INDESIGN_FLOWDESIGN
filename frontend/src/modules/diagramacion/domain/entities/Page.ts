import { Pauta } from './Pauta';

export class Page {
  constructor(
    public readonly id: number,
    public readonly noPagina: number,
    public readonly pautas: Pauta[]
  ) {}
}
