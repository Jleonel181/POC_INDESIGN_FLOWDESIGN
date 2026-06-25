import { DataSource } from "typeorm";
import { EnvironmentConfig } from "../config/environment.config";
import { MysqlAdDataSourceAdapter } from "../modules/esb/infraestructure/adapters/MysqlAdDataSourceAdapter";
import { GetImportedAdsUseCase } from "../modules/esb/application/use-cases/GetImportedAdsUseCase";
import { EsbController } from "../modules/esb/infraestructure/http/EsbController";
import { GenerateEditionLayoutUseCase } from "../modules/layout/application/use-cases/GenerateEditionLayoutUseCase";
import { GetAllEditionsLayoutUseCase } from "../modules/layout/application/use-cases/GetAllEditionsLayoutUseCase";
import { GridLayoutCalculator } from "../modules/layout/domain/services/GridLayoutCalculator";
import { LayoutValidator } from "../modules/layout/domain/services/LayoutValidator";
import { LayoutController } from "../modules/layout/infraestructure/http/LayoutController";
import { CreateEditionUseCase } from "../modules/editions/application/use-cases/CreateEditionUseCase";
import { EditionController } from "../modules/editions/infraestructure/http/EditionController";
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

    const createEditionUseCase = new CreateEditionUseCase(
        editionRepository,
        pageRepository
    );

    // Infrastructure layer - Controllers
    const layoutController = new LayoutController(generateEditionLayoutUseCase, getAllEditionsLayoutUseCase);
    const editionController = new EditionController(createEditionUseCase);

    // ESB module — swap MysqlAdDataSourceAdapter for another adapter when needed
    const esbDataSource = new MysqlAdDataSourceAdapter(
        EnvironmentConfig.getInstance().getAppConfig().ventasDatabase
    );
    const getImportedAdsUseCase = new GetImportedAdsUseCase(esbDataSource);
    const esbController = new EsbController(getImportedAdsUseCase);

    return {
        layoutController,
        editionController,
        esbController
    };
}