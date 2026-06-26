import "dotenv/config";
import { createApp } from "./bootstrap/app";
import { EnvironmentConfig } from "./config/environment.config";

const port = EnvironmentConfig.getInstance().getConfig().port;
const app = createApp();

app.listen(port, () => {
    console.log(`✅ ESB running on port ${port}`);
    console.log(`🔗 Health: http://localhost:${port}/health`);
    console.log(`📦 Ventas: http://localhost:${port}/api/ventas/ads?date=YYYY-MM-DD`);
});
