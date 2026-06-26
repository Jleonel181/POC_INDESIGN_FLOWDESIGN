export interface DatabaseConfig {
    host: string;
    port: number;
    username: string;
    password: string;
    database: string;
    synchronize: boolean;
    logging: boolean;
}

export interface CorsConfig {
    allowedOrigins: string[];
    allowedMethods: string[];
    allowedHeaders: string[];
    credentials: boolean;
}

export interface AppConfig {
    port: number;
    database: DatabaseConfig;
    esbUrl: string;
    cors: CorsConfig;
}

export class EnvironmentConfig {
    private static instance: EnvironmentConfig;

    private constructor() {}

    static getInstance(): EnvironmentConfig {
        if (!EnvironmentConfig.instance) {
            EnvironmentConfig.instance = new EnvironmentConfig();
        }
        return EnvironmentConfig.instance;
    }

    getAppConfig(): AppConfig {
        return {
            port: this.getNumber('PORT', 3000),
            database: this.getDatabaseConfig(),
            esbUrl: this.getString('ESB_URL', 'http://localhost:4000'),
            cors: this.getCorsConfig()
        };
    }

    private getCorsConfig(): CorsConfig {
        const origins = this.getString('CORS_ALLOWED_ORIGINS', 'http://localhost:3000');
        return {
            allowedOrigins: origins.split(',').map(o => o.trim()),
            allowedMethods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
            allowedHeaders: ['Content-Type', 'Authorization'],
            credentials: this.getBoolean('CORS_CREDENTIALS', true),
        };
    }

    private getDatabaseConfig(): DatabaseConfig {
        return {
            host: this.getString('DB_HOST', 'localhost'),
            port: this.getNumber('DB_PORT', 5432),
            username: this.getString('DB_USER', 'designflow'),
            password: this.getString('DB_PASSWORD', 'designflow123'),
            database: this.getString('DB_NAME', 'designflow_db'),
            synchronize: this.getBoolean('DB_SYNCHRONIZE', false),
            logging: this.getBoolean('DB_LOGGING', false)
        };
    }

    private getString(key: string, defaultValue: string): string {
        return process.env[key] || defaultValue;
    }

    private getNumber(key: string, defaultValue: number): number {
        const value = process.env[key];
        return value ? parseInt(value, 10) : defaultValue;
    }

    private getBoolean(key: string, defaultValue: boolean): boolean {
        const value = process.env[key];
        if (value === undefined) return defaultValue;
        return value.toLowerCase() === 'true';
    }
}
