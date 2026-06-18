export class GridPosition {
    constructor(public readonly x: number, public readonly y: number) {

        if(x < 0){
            throw new Error ('La posición x no puede ser negativa')
        }
        if(y < 0){
            throw new Error ('La posición y no puede ser negativa')
        }

    }
}