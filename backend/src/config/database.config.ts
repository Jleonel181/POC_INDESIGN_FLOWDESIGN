import "reflect-metadata";
import { DataSource } from "typeorm";
import { EnvironmentConfig } from "./environment.config";
import { EditionEntity } from "../modules/editions/infraestructure/persistence/entities/EditionEntity";
import { PageEntity } from "../modules/pages/infraestructure/persistence/entities/PageEntity";
import { PautaEntity } from "../modules/pautas/infraestructure/persistence/entities/PautaEntity";

export class DatabaseConnection {
    private static instance: DatabaseConnection;
    private dataSource: DataSource;

    private constructor() {
        const config = EnvironmentConfig.getInstance().getAppConfig();
        
        this.dataSource = new DataSource({
            type: "postgres",
            host: config.database.host,
            port: config.database.port,
            username: config.database.username,
            password: config.database.password,
            database: config.database.database,
            synchronize: config.database.synchronize,
            logging: config.database.logging,
            entities: [EditionEntity, PageEntity, PautaEntity],
            subscribers: [],
            migrations: [],
        });
    }

    static getInstance(): DatabaseConnection {
        if (!DatabaseConnection.instance) {
            DatabaseConnection.instance = new DatabaseConnection();
        }
        return DatabaseConnection.instance;
    }

    async initialize(): Promise<void> {
        if (!this.dataSource.isInitialized) {
            await this.dataSource.initialize();
            console.log("Database connection established successfully");
        }
    }

    async close(): Promise<void> {
        if (this.dataSource.isInitialized) {
            await this.dataSource.destroy();
            console.log("Database connection closed");
        }
    }

    getDataSource(): DataSource {
        if (!this.dataSource.isInitialized) {
            throw new Error("Database connection is not initialized. Call initialize() first.");
        }
        return this.dataSource;
    }
}
