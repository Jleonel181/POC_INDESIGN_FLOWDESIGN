import path from "path";
import { DataSource } from "typeorm";
import { EnvironmentConfig } from "../config/environment.config";
import { HttpEsbAdapter } from "../modules/ventas/infraestructure/http/HttpEsbAdapter";
import { VentasController } from "../modules/ventas/infraestructure/http/VentasController";
import { GenerateEditionLayoutUseCase } from "../modules/layout/application/use-cases/GenerateEditionLayoutUseCase";
import { GetAllEditionsLayoutUseCase } from "../modules/layout/application/use-cases/GetAllEditionsLayoutUseCase";
import { GenerateIdmlUseCase } from "../modules/layout/application/use-cases/GenerateIdmlUseCase";
import { GridLayoutCalculator } from "../modules/layout/domain/services/GridLayoutCalculator";
import { LayoutValidator } from "../modules/layout/domain/services/LayoutValidator";
import { IdmlgenProcessAdapter } from "../modules/layout/infraestructure/idml/IdmlgenProcessAdapter";
import { LayoutController } from "../modules/layout/infraestructure/http/LayoutController";
import { CreateEditionUseCase } from "../modules/editions/application/use-cases/CreateEditionUseCase";
import { EditionController } from "../modules/editions/infraestructure/http/EditionController";
import { CreatePautaUseCase } from "../modules/pautas/application/use-cases/CreatePautaUseCase";
import { PautaController } from "../modules/pautas/infraestructure/http/PautaController";
import { PostgresEditionRepository } from "../modules/editions/infraestructure/persistence/PostgresEditionRepository";
import { PostgresPageRepository } from "../modules/pages/infraestructure/persistence/PostgresPageRepository";
import { PostgresPautaRepository } from "../modules/pautas/infraestructure/persistence/PostgresPautaRepository";
import { EditionEntity } from "../modules/editions/infraestructure/persistence/entities/EditionEntity";
import { PageEntity } from "../modules/pages/infraestructure/persistence/entities/PageEntity";
import { PautaEntity } from "../modules/pautas/infraestructure/persistence/entities/PautaEntity";

/**
 * Dependency Container
 * Applies Dependency Injection and Inversion of Control principles
 * Following SOLID principles for maintainable and testable code
 */
export function createDependencyContainer(dataSource: DataSource) {
    // Infrastructure layer - Repository implementations
    const editionRepository = new PostgresEditionRepository(
        dataSource.getRepository(EditionEntity)
    );
    const pageRepository = new PostgresPageRepository(
        dataSource.getRepository(PageEntity)
    );
    const pautaRepository = new PostgresPautaRepository(
        dataSource.getRepository(PautaEntity)
    );
    
    // Domain layer - Services
    const gridLayoutCalculator = new GridLayoutCalculator();
    const layoutValidator = new LayoutValidator();

    // Infrastructure layer - Adapters
    const idmlgenPath = process.env.IDMLGEN_PATH || path.resolve(__dirname, "../../idmllib/bin/idmlgen");
    const idmlGenerator = new IdmlgenProcessAdapter(idmlgenPath);

    // Application layer - Use Cases
    const generateEditionLayoutUseCase = new GenerateEditionLayoutUseCase(
        editionRepository,
        pageRepository,
        pautaRepository,
        gridLayoutCalculator,
        layoutValidator
    );

    const getAllEditionsLayoutUseCase = new GetAllEditionsLayoutUseCase(
        editionRepository,
        generateEditionLayoutUseCase
    );

    const generateIdmlUseCase = new GenerateIdmlUseCase(
        generateEditionLayoutUseCase,
        idmlGenerator
    );

    const createEditionUseCase = new CreateEditionUseCase(
        editionRepository,
        pageRepository
    );

    const createPautaUseCase = new CreatePautaUseCase(
        pautaRepository
    );

    // Infrastructure layer - Controllers
    const layoutController = new LayoutController(
        generateEditionLayoutUseCase,
        getAllEditionsLayoutUseCase,
        generateIdmlUseCase
    );
    const editionController = new EditionController(createEditionUseCase);
    const pautaController = new PautaController(createPautaUseCase, pautaRepository);

    // Ventas module — consumes ESB via HTTP, no direct DB credentials here
    const esbAdapter = new HttpEsbAdapter(EnvironmentConfig.getInstance().getAppConfig().esbUrl);
    const ventasController = new VentasController(esbAdapter);

    return {
        layoutController,
        editionController,
        pautaController,
        ventasController
    };
}
