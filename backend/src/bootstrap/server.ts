import { DatabaseConnection } from "../config/database.config";
import { createApp } from "./app";

export async function startServer(): Promise<void> {
    try {
        // Initialize database connection
        const dbConnection = DatabaseConnection.getInstance();
        await dbConnection.initialize();
        
        const dataSource = dbConnection.getDataSource();
        
        // Create and start Express app
        const port = Number(process.env.PORT ?? 3000);
        const app = createApp(dataSource);
        
        app.listen(port, () => {
            console.log(`✅ Server is running on port ${port}`);
            console.log(`🔗 Health check: http://localhost:${port}/health`);
        });

        // Graceful shutdown
        process.on('SIGINT', async () => {
            console.log('\n🛑 Shutting down gracefully...');
            await dbConnection.close();
            process.exit(0);
        });

        process.on('SIGTERM', async () => {
            console.log('\n🛑 SIGTERM received, shutting down...');
            await dbConnection.close();
            process.exit(0);
        });

    } catch (error) {
        console.error('❌ Failed to start server:', error);
        process.exit(1);
    }
}