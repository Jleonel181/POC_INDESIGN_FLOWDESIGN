# Frontend — Interfaz de Usuario

## Visión General

El frontend es la interfaz web para diagramadores. Permite visualizar ediciones, páginas y anuncios, y gestionar la posición de estos en las publicaciones.

**Stack:** Next.js 14+ + React + TypeScript

**Puerto:** 3000 (local)

## Estructura

```
frontend/
├── src/
│   ├── app/                   # App Router (Next.js 14)
│   │   ├── api/               # API routes (edge functions)
│   │   ├── diagramador/       # Página principal
│   │   ├── dashboard/         # Dashboard
│   │   ├── layout.tsx         # Root layout
│   │   ├── page.tsx           # Home
│   │   └── globals.css        # Estilos globales
│   │
│   ├── config/
│   │   └── appConfig.ts       # Configuración de la app
│   │
│   ├── modules/               # Módulos por feature
│   │   ├── dashboard/
│   │   │   ├── components/
│   │   │   ├── hooks/
│   │   │   └── services/
│   │   └── diagramacion/
│   │       ├── components/
│   │       ├── hooks/
│   │       └── services/
│   │
│   └── shared/
│       ├── application/
│       ├── domain/
│       ├── infraestructure/
│       └── presentation/
│
├── public/                    # Assets estáticos
├── package.json
├── next.config.ts
└── tsconfig.json
```

## Módulos

### Dashboard

Página de inicio que muestra:
- Ediciones disponibles
- Estadísticas
- Acciones rápidas

### Diagramación

Herramienta principal para diagramar.

**Funcionalidades:**
- Visualizar grilla de página
- Ver anuncios disponibles
- Arrastrar y soltar anuncios
- Guardar posiciones

## Configuración

### Archivo: `src/config/appConfig.ts`

```typescript
export const appConfig = {
  apiBaseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001',
  esbUrl: process.env.NEXT_PUBLIC_ESB_URL || 'http://localhost:4000',
};
```

### Variables de Entorno

**Archivo:** `frontend/.env.local`

```bash
# Backend
NEXT_PUBLIC_API_URL=http://localhost:3001

# ESB (opcional, para consumir directo)
NEXT_PUBLIC_ESB_URL=http://localhost:4000
```

> `NEXT_PUBLIC_*` hace las variables accesibles en el cliente.

## Desarrollo

### Levantar en local

```bash
cd frontend
npm install
npm run dev
```

El servidor corre en `http://localhost:3000` con hot reload.

### Build para producción

```bash
npm run build
npm start
```

## Llamadas al Backend

### Obtener ediciones

```typescript
// frontend/src/shared/infraestructure/api.ts
export async function getEditions() {
  const response = await fetch(`${appConfig.apiBaseUrl}/api/editions`);
  return response.json();
}
```

### Obtener anuncios de un día

```typescript
export async function getAdsByDate(date: string) {
  const response = await fetch(
    `${appConfig.apiBaseUrl}/api/ventas/ads?date=${date}`
  );
  return response.json();
}
```

### Crear una edición

```typescript
export async function createEdition(data: {
  no_paginas: number;
  ancho_mm: number;
  alto_mm: number;
  cuadros_ancho: number;
  cuadros_alto: number;
}) {
  const response = await fetch(`${appConfig.apiBaseUrl}/api/editions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return response.json();
}
```

## Comunicación Backend-Frontend

```
Frontend Browser
      ↓
  (HTTP/CORS)
      ↓
Backend API :3001
      ↓
  Valida CORS_ALLOWED_ORIGINS
      ↓
  Responde con datos
      ↓
Frontend recibe y renderiza
```

### CORS Configuration

El backend está configurado para aceptar requests del frontend:

**Backend .env:**
```
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_CREDENTIALS=true
```

## Estructura de Componentes

Recomendado usar esta estructura para nuevos componentes:

```
src/modules/feature/
├── components/
│   ├── FeatureName.tsx
│   ├── FeatureName.module.css
│   └── subcomponents/
├── hooks/
│   └── useFeatureName.ts
└── services/
    └── featureService.ts
```

## Performance

- **Image Optimization:** Next.js `<Image>` component
- **Code Splitting:** Automático por ruta
- **Static Generation:** Para páginas estáticas
- **API Routes:** Para funciones sin servidor

## Despliegue

### Local
```bash
npm run dev
```

### Docker
```bash
docker build -t flowdesign-frontend .
docker run -p 3000:3000 flowdesign-frontend
```

### AWS (futuro)
- Opción 1: CloudFront + S3 (estático)
- Opción 2: App Runner (full)
- Opción 3: Vercel (recomendado para Next.js)

## Troubleshooting

### Error: "Failed to fetch from API"
- Verificar que backend está corriendo en `:3001`
- Verificar que `NEXT_PUBLIC_API_URL` es correcto
- Verificar CORS en backend

### Error: "Cannot find module"
- Correr `npm install`
- Verificar imports (case-sensitive en Linux)

### Slow performance en desarrollo
- Next.js dev mode es lento en Docker
- Correr localmente con `npm run dev`

---

**Última actualización:** 2026-01-10
