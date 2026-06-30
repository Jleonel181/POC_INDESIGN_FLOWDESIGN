# POC_INDESIGN_FLOWDESIGN

Prueba de concepto de FlowDesign: backend, frontend y ESB para integración con sistemas de ventas.

## Visión General

Este proyecto implementa una arquitectura de microservicios desacoplada que integra un sistema de diagramación editorial (FlowDesign) con un sistema externo de gestión de ventas mediante un Enterprise Service Bus (ESB).

```
Frontend (Next.js)
      ↓
Backend (Node.js/Express)
      ↓
ESB (Node.js/Express)
      ↓
MySQL Ventas (on-premise)
```

## Documentación

- **[ARQUITECTURA.md](./docs/ARQUITECTURA.md)** — Visión general, componentes, decisiones técnicas
- **[BACKEND.md](./docs/BACKEND.md)** — API, endpoints, estructura hexagonal
- **[FRONTEND.md](./docs/FRONTEND.md)** — Next.js, configuración, módulos
- **[ESB.md](./docs/ESB.md)** — Enterprise Service Bus, adaptadores, transformadores
- **[DATABASE.md](./docs/DATABASE.md)** — Schema PostgreSQL, migraciones, datos
- **[CI_CD.md](./.github/CICD.md)** — GitHub Actions, pipelines, estrategia de branching
- **[EQUIPO.md](./docs/EQUIPO.md)** — Cómo agregar usuarios, permisos, revisores

## Quick Start

### Requisitos

- Docker y Docker Compose
- Node.js 22+
- MySQL accesible en `192.168.28.33` (para el ESB)

### Levantar en local

```bash
# 1. Clonar el repo
git clone <repo-url>
cd POC_InDesign_FlowDesign

# 2. Copiar archivos de configuración
cp backend/.env.example backend/.env
cp esb/.env.example esb/.env

# 3. Levantar servicios con Docker Compose
docker compose up

# En otra terminal, levantar frontend
cd frontend
npm install
npm run dev
```

Endpoints disponibles:
- Frontend: http://localhost:3000
- Backend: http://localhost:3001
- ESB: http://localhost:4000
- PostgreSQL: localhost:5432

### Verificar que todo funciona

```bash
# ESB health
curl http://localhost:4000/health

# Backend health
curl http://localhost:3001/health

# Obtener anuncios de un día específico
curl "http://localhost:3001/api/ventas/ads?date=2026-06-11"
```

## Estructura del Proyecto

```
POC_InDesign_FlowDesign/
├── backend/              # API principal (Node.js/Express)
│   ├── src/
│   │   ├── modules/      # Arquitectura hexagonal
│   │   ├── config/
│   │   └── bootstrap/
│   ├── Dockerfile
│   └── package.json
│
├── esb/                  # Enterprise Service Bus (Node.js/Express)
│   ├── src/
│   │   ├── modules/ventas/
│   │   ├── config/
│   │   └── bootstrap/
│   ├── Dockerfile
│   └── package.json
│
├── frontend/             # UI (Next.js)
│   ├── src/app/
│   ├── package.json
│   └── next.config.ts
│
├── .github/
│   ├── workflows/        # CI/CD pipelines
│   └── CICD.md
│
├── docs/                 # Documentación
│   ├── ARQUITECTURA.md
│   ├── BACKEND.md
│   ├── FRONTEND.md
│   ├── ESB.md
│   ├── DATABASE.md
│   └── EQUIPO.md
│
└── docker-compose.yml    # Orquestación local
```

## Stack Tecnológico

| Componente | Tecnología | Versión |
|---|---|---|
| Frontend | Next.js | 14+ |
| Backend | Node.js/Express | 22-alpine |
| ESB | Node.js/Express | 22-alpine |
| DB Principal | PostgreSQL | 16-alpine |
| DB Ventas | MySQL | on-premise |
| Orquestación | Docker Compose | 3.8 |
| CI/CD | GitHub Actions | - |

## Flujo de Desarrollo

1. Crear rama feature desde `develop`
2. Hacer push y abrir PR hacia `develop`
3. CI valida que el build funciona (docker-ci.yml)
4. Hacer merge a `develop` → Se publica imagen staging en Docker Hub
5. Validar en staging
6. Abrir PR de `develop` → `main`
7. Hacer merge a `main` → Se publica imagen production (requiere aprobación)

Ver [CI_CD.md](./.github/CICD.md) para detalles completos.

## Despliegue a Producción (AWS)

Cuando escales a AWS:
- Backend: ECS Task
- ESB: Lambda + API Gateway (costo bajo)
- DB: RDS PostgreSQL
- Secretos: AWS Secrets Manager
- Red: VPC con Site-to-Site VPN a on-premise

Ver [ARQUITECTURA.md](./docs/ARQUITECTURA.md) para detalles técnicos.

## Seguridad

- Variables sensibles en `.env` (local) → Secrets Manager (AWS)
- `.env` nunca se commitea (está en .gitignore)
- Base de datos de ventas: acceso de solo lectura
- CORS configurado por ambiente
- Dockerfile multi-stage para optimizar imagen

Ver [EQUIPO.md](./docs/EQUIPO.md) para permisos y autenticación.

## Contacto y Soporte

Para preguntas sobre la arquitectura, revisar la documentación en `docs/` o contactar al equipo de desarrollo.

---

**Versión:** 1.0.0-poc
