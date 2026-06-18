import { UseCase } from "../../../../shared/application/UseCase";
import { EditionRepository } from "../../../editions/domain/repositories/EditionRepository";
import { GenerateEditionLayoutUseCase } from "./GenerateEditionLayoutUseCase";
import { EditionLayoutResponseDTO } from "../dto/EditionLayoutResponseDTO";

export class GetAllEditionsLayoutUseCase implements UseCase<void, EditionLayoutResponseDTO[]> {
    constructor(
        private readonly editionRepository: EditionRepository,
        private readonly generateEditionLayoutUseCase: GenerateEditionLayoutUseCase
    ) {}

    async execute(): Promise<EditionLayoutResponseDTO[]> {
        const editions = await this.editionRepository.findAll();

        const layouts = await Promise.all(
            editions.map((edition) =>
                this.generateEditionLayoutUseCase.execute({ editionId: edition.id })
            )
        );

        return layouts;
    }
}
