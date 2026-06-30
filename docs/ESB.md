# ESB — Enterprise Service Bus

## Visión General

El ESB es un servicio independiente que integra sistemas externos con FlowDesign. Su responsabilidad es consultar datos de sistemas legacy y transformarlos al dominio de FlowDesign.

**Stack:** Node.js 22 + Express + TypeScript

**Puerto:** 4000 (local)

**Responsabilidad:** Lectura de datos externos, transformación, exposición por API HTTP

## Arquitectura

```
ESB Service
├── Domain Layer
│   └── AdImport.ts              # Entidad de dominio
│
├── Application Layer
│   ├── Ports/
│   │   └── AdDataSourcePort.ts  # Interfaz (contrato)
│   └── Use Cases/
│       └── GetImportedAdsUseCase.ts
│
└── Infrastructure Layer
    ├── Adapters/
    │   └── MysqlAdDataSourceAdapter.ts  # MySQL real
    ├── Transformers/
    │   └── AdImportTransformer.ts
    └── HTTP/
        ├── VentasController.ts
        └── VentasRoutes.ts
```

## Módulos

### Ventas

Integración con el sistema de ventas (MySQL on-premise).

**Domain:**
- `AdImport` — Entidad que representa un anuncio importado

**Application:**
- `AdDataSourcePort` — Interfaz que abstrae la fuente de datos
- `GetImportedAdsUseCase` — Caso de uso: obtener anuncios por fecha

**Infrastructure:**
- `MysqlAdDataSourceAdapter` — Implementación que lee de MySQL
- `AdImportTransformer` — Transforma datos crudos a `AdImport`
- `VentasController` — Expone endpoint HTTP
- `VentasRoutes` — Define rutas

## Endpoints

### Health Check

```
GET /health

Respuesta:
{
  "status": "ok",
  "service": "flowdesign-esb"
}
```

### Obtener Anuncios por Fecha

```
GET /api/ventas/ads?date=YYYY-MM-DD

Query params:
- date (requerido): Fecha en formato YYYY-MM-DD

Respuesta:
{
  "date": "2026-06-11",
  "total": 45,
  "ads": [
    {
      "fid": "3681453",
      "customer": "Diarios Modernos, S.a.",
      "product": "Centro Occidente",
      "slogan": "NUESTRO TRABAJO",
      "coverDate": "2026-06-11",
      "formatFid": "6.00x8.00-MN",
      "cuadrosAncho": 6,
      "cuadrosAlto": 8,
      "widthMm": 254,
      "heightMm": 317.5,
      "colorFid": "Full",
      "placementComment": "Desplegado",
      "statusFid": "Por emitir",
      "printSystemFid": "32542600042",
      "regionalName": "Region Centro Occidente",
      "productCategory": "Departamental",
      "observations": ""
    }
  ]
}
```

## Flujo de una Request

```
1. Cliente solicita: GET /api/ventas/ads?date=2026-06-11
   ↓
2. VentasController.getAdsByDate()
   - Valida que date está presente y tiene formato YYYY-MM-DD
   ↓
3. GetImportedAdsUseCase.execute(date)
   - Delega a AdDataSourcePort.fetchByDate(date)
   ↓
4. MysqlAdDataSourceAdapter.fetchByDate(date)
   - Abre conexión a MySQL 192.168.28.33
   - Ejecuta: SELECT ... FROM vw_pauta_dataplan WHERE publication_date = '2026-06-11'
   - Retorna array de registros crudos
   ↓
5. AdImportTransformer.transform(raw) [para cada registro]
   - Convierte Width/Height de pulgadas a milímetros
   - Asigna grilla fija 6x8
   - Mapea campos al dominio AdImport
   ↓
6. Retorna array de AdImport al caso de uso
   ↓
7. VentasController formatea respuesta JSON
   ↓
8. Cliente recibe respuesta
```

## Configuración

### Variables de Entorno

**Archivo:** `esb/.env`

```bash
# Puerto
PORT=4000

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3001

# MySQL Ventas (on-premise)
DB_VENTAS_SERVER=jdbc:mysql://192.168.28.33/ventas?characterEncoding=latin1
DB_VENTAS_USER=dataplan
DB_VENTAS_PASSWORD=D@taP1an
```

### Parsing de JDBC URL

El ESB parsea automáticamente el JDBC URL para extraer host y base de datos:

```typescript
// De: jdbc:mysql://192.168.28.33/ventas?characterEncoding=latin1
// A:
{
  host: "192.168.28.33",
  port: 3306,
  database: "ventas",
  username: "dataplan",
  password: "D@taP1an"
}
```

## Adaptadores

El punto clave del ESB es que los adaptadores son **reemplazables** sin cambiar la lógica.

### AdDataSourcePort (Interfaz)

```typescript
export interface AdDataSourcePort {
    fetchByDate(date: string): Promise<AdImport[]>;
}
```

### Adaptador Actual: MysqlAdDataSourceAdapter

Conecta a MySQL on-premise.

```typescript
const adapter = new MysqlAdDataSourceAdapter(config);
const ads = await adapter.fetchByDate("2026-06-11");
```

### Adaptadores Futuros

Sin cambiar el caso de uso ni el controlador, podrías agregar:

**RestVentasAdapter** — Si el sistema de ventas expone una API REST
```typescript
const adapter = new RestVentasAdapter("https://ventas-api.com");
```

**FileVentasAdapter** — Si los datos llegan en archivos CSV/JSON
```typescript
const adapter = new FileVentasAdapter("/data/ventas.json");
```

**LambdaVentasAdapter** — Si quieres consultarle a una Lambda en AWS
```typescript
const adapter = new LambdaVentasAdapter("arn:aws:lambda:...");
```

Todo sin tocar `GetImportedAdsUseCase` ni `VentasController`.

## Transformadores

### AdImportTransformer

Convierte datos crudos de MySQL al dominio `AdImport`.

**Transformaciones:**
- Pulgadas → Milímetros: `inches * 25.4`
- Asigna grilla fija: `6 × 8` (configurable)
- Mapea campos: `Width` → `widthMm`, etc.

**Ejemplo:**

Input (de MySQL):
```json
{
  "FID": "3681453",
  "Width": 10.0,
  "Height": 12.5,
  "FormatFID": "6.00x8.00-MN"
}
```

Output (AdImport):
```json
{
  "fid": "3681453",
  "widthMm": 254.0,
  "heightMm": 317.5,
  "cuadrosAncho": 6,
  "cuadrosAlto": 8,
  "formatFid": "6.00x8.00-MN"
}
```

## Desarrollo

### Levantar en local

```bash
cd esb
npm install
npm run dev
```

El servidor corre en `http://localhost:4000` con hot reload.

### Testing Manual

```bash
# Health
curl http://localhost:4000/health

# Obtener anuncios
curl "http://localhost:4000/api/ventas/ads?date=2026-06-11"

# Con jq para pretty-print
curl "http://localhost:4000/api/ventas/ads?date=2026-06-11" | jq
```

## Escalabilidad

### Costo Bajo: Lambda + API Gateway

Cuando el ESB escale a AWS con bajo presupuesto:

```typescript
// esb/src/lambda.ts
import { APIGatewayProxyHandler } from 'aws-lambda';

export const handler: APIGatewayProxyHandler = async (event) => {
  const date = event.queryStringParameters?.date;
  const useCase = new GetImportedAdsUseCase(
    new MysqlAdDataSourceAdapter(config)
  );
  const ads = await useCase.execute(date);
  return { statusCode: 200, body: JSON.stringify({ ads }) };
};
```

El código es 95% igual, solo cambias el entry point de Express a Lambda.

### Alto Volumen: ECS

Si los datos explotan, migra a ECS:

```yaml
# Task Definition en AWS
esb:
  image: 123456.dkr.ecr.us-east-1.amazonaws.com/esb:latest
  cpu: 512
  memory: 1024
  environment:
    - DB_VENTAS_SERVER=...
  secrets:
    - name: DB_VENTAS_PASSWORD
      valueFrom: arn:aws:secretsmanager:...
```

El código sigue siendo el mismo Express app.

## Troubleshooting

### Error: "Cannot connect to MySQL"

Checklist:
1. ¿MySQL está corriendo en `192.168.28.33:3306`?
   ```bash
   mysql -h 192.168.28.33 -u dataplan -p
   ```

2. ¿El usuario `dataplan` existe y tiene permisos?
   ```sql
   SELECT * FROM vw_pauta_dataplan LIMIT 1;
   ```

3. ¿La VPN/conexión a on-premise está activa?
   ```bash
   ping 192.168.28.33
   ```

4. ¿El charset es correcto?
   - El ESB usa `LATIN1`
   - MySQL debe tener la vista `vw_pauta_dataplan` con ese charset

### Error: "Query timeout after 10000ms"

- MySQL on-premise está lento
- Aumentar timeout en `MysqlAdDataSourceAdapter`:
  ```typescript
  connectTimeout: 30000  // 30 segundos
  ```

### Error: "Invalid date format"

- La fecha debe ser `YYYY-MM-DD`
- ✅ Correcto: `2026-06-11`
- ❌ Incorrecto: `06-11-2026`, `2026/06/11`

## Monitoreo (futuro)

- Logs estructurados en CloudWatch
- Métricas de latencia a MySQL
- Alertas si la conexión falla

---

**Última actualización:** 2026-01-10
