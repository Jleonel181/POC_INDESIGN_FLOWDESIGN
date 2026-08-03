import { spawn } from "child_process";
import { IdmlGenerator } from "../../domain/ports/IdmlGenerator";
import { IdmlDocumentDTO } from "../../application/dto/IdmlDocumentDTO";

export class IdmlgenProcessAdapter implements IdmlGenerator {
    constructor(private readonly binPath: string) {}

    generate(document: IdmlDocumentDTO): Promise<Buffer> {
        return new Promise((resolve, reject) => {
            const proc = spawn(this.binPath, [], { stdio: ["pipe", "pipe", "pipe"] });

            const chunks: Buffer[] = [];
            const stderrChunks: Buffer[] = [];

            proc.stdout.on("data", (chunk: Buffer) => chunks.push(chunk));
            proc.stderr.on("data", (chunk: Buffer) => stderrChunks.push(chunk));

            proc.on("close", (code) => {
                if (code !== 0) {
                    const stderr = Buffer.concat(stderrChunks).toString().trim();
                    const err = new Error(`idmlgen falló (exit ${code}): ${stderr}`);
                    (err as any).exitCode = code;
                    reject(err);
                    return;
                }
                resolve(Buffer.concat(chunks));
            });

            proc.on("error", (err) => {
                reject(new Error(`No se pudo ejecutar idmlgen: ${err.message}`));
            });

            proc.stdin.write(JSON.stringify(document));
            proc.stdin.end();
        });
    }
}
