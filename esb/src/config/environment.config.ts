export interface VentasDbConfig {
    host: string;
    port: number;
    username: string;
    password: string;
    database: string;
}

export interface EsbConfig {
    port: number;
    corsAllowedOrigins: string[];
    ventasDatabase: VentasDbConfig;
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

    getConfig(): EsbConfig {
        return {
            port: this.getNumber("PORT", 4000),
            corsAllowedOrigins: this.getString("CORS_ALLOWED_ORIGINS", "http://localhost:3001")
                .split(",")
                .map(o => o.trim()),
            ventasDatabase: this.getVentasDbConfig(),
        };
    }

    private getVentasDbConfig(): VentasDbConfig {
        const jdbcUrl = this.getString("DB_VENTAS_SERVER", "");
        const match = jdbcUrl.match(/jdbc:mysql:\/\/([^/]+)\/([^?]+)/);
        return {
            host: match?.[1] ?? "192.168.28.33",
            port: 3306,
            username: this.getString("DB_VENTAS_USER", ""),
            password: this.getString("DB_VENTAS_PASSWORD", ""),
            database: match?.[2] ?? "ventas",
        };
    }

    private getString(key: string, defaultValue: string): string {
        return process.env[key] || defaultValue;
    }

    private getNumber(key: string, defaultValue: number): number {
        const value = process.env[key];
        return value ? parseInt(value, 10) : defaultValue;
    }
}
