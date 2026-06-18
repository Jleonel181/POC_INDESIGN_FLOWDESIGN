import { PageLayout } from "./PageLayout";


export class EditionLayout {
    constructor(
        public readonly editionId: number,
        public readonly no_paginas: number,
        public readonly ancho_mm: number,
        public readonly alto_mm: number,
        public readonly cuadros_ancho: number,
        public readonly cuadros_alto: number,
        public readonly pageLayouts: PageLayout[]
    ){}
}