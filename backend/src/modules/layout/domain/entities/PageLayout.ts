import { FrameLayout } from "./FrameLayout";


export class PageLayout {

    constructor(
        public readonly paginaId: number,
        public readonly no_pagina: number,
        public readonly frames: FrameLayout[]
    ){}
}