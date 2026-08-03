import { IdmlDocumentDTO } from "../../application/dto/IdmlDocumentDTO";

export interface IdmlGenerator {
    generate(document: IdmlDocumentDTO): Promise<Buffer>;
}
