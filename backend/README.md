# DesignFlow Backend - PostgreSQL con Docker

Backend con arquitectura hexagonal, TypeORM y PostgreSQL.

## 🚀 Inicio Rápido

### 1. Levantar PostgreSQL con Docker

```bash
docker-compose up -d
```

Esto iniciará PostgreSQL en el puerto 5432 con la base de datos inicializada.

### 2. Verificar que PostgreSQL esté corriendo

```bash
docker ps
```

Deberías ver el contenedor `designflow-postgres` en estado "healthy".

### 3. Instalar dependencias (si no lo has hecho)

```bash
cd backend
npm install
```

### 4. Iniciar el servidor

```bash
npm run dev
```

El servidor estará corriendo en `http://localhost:3000`

## 🧪 Verificar funcionamiento

### Health Check
```bash
curl http://localhost:3000/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "service": "designflow-backend",
  "database": "connected"
}
```

### Generar Layout de Edición
```bash
curl http://localhost:3000/api/layout/1
```

## 🗄️ Gestión de Base de Datos

### Conectarse a PostgreSQL
```bash
docker exec -it designflow-postgres psql -U designflow -d designflow_db
```

### Ver datos de las tablas
```sql
SELECT * FROM editions;
SELECT * FROM pages;
SELECT * FROM pautas;
```

### Detener contenedor
```bash
docker-compose down
```

### Detener y eliminar datos
```bash
docker-compose down -v
```

## 📁 Estructura del Proyecto

```
backend/
├── src/
│   ├── bootstrap/           # Inicialización y DI
│   ├── config/              # Configuraciones
│   ├── modules/
│   │   ├── editions/
│   │   │   ├── domain/
│   │   │   │   ├── entities/
│   │   │   │   └── repositories/
│   │   │   └── infraestructure/
│   │   │       └── persistence/
│   │   │           ├── entities/     # TypeORM Entities
│   │   │           ├── mappers/      # Entity ↔ Domain
│   │   │           └── PostgresEditionRepository.ts
│   │   ├── pages/
│   │   └── pautas/
│   └── shared/
└── database/
    └── init.sql             # Schema y datos iniciales
```

## 🛠️ Variables de Entorno

Archivo `.env`:
```
PORT=3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=designflow
DB_PASSWORD=designflow123
DB_NAME=designflow_db
DB_SYNCHRONIZE=false
DB_LOGGING=true
```

## 🏗️ Arquitectura

### Principios SOLID aplicados:

- **Single Responsibility**: Cada clase tiene una única responsabilidad
- **Open/Closed**: Repositorios implementan interfaces, fácil extensión
- **Liskov Substitution**: Cualquier implementación de repository funciona
- **Interface Segregation**: Interfaces específicas por módulo
- **Dependency Inversion**: Dependemos de abstracciones (interfaces)

### Capas:

1. **Domain**: Entidades y contratos (independiente de infraestructura)
2. **Application**: Casos de uso
3. **Infrastructure**: TypeORM, PostgreSQL, HTTP controllers

## 🔍 Troubleshooting

### Error de conexión a base de datos

1. Verificar que Docker esté corriendo: `docker ps`
2. Ver logs: `docker logs designflow-postgres`
3. Reiniciar contenedor: `docker-compose restart`

### Puerto 5432 ocupado

Cambiar en `.env`:
```
DB_PORT=5433
```

Y en `docker-compose.yml`:
```yaml
ports:
  - "5433:5432"
```

## 📝 Notas

- Los datos se persisten en un volumen Docker (`postgres_data`)
- El script `init.sql` se ejecuta solo la primera vez
- Para resetear datos: `docker-compose down -v && docker-compose up -d`
