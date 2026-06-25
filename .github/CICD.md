# CI/CD Pipeline — POC InDesign FlowDesign

Documentación del flujo de integración y entrega continua implementado con GitHub Actions y Docker Hub.

---

## Tabla de contenidos

1. [Resumen del flujo](#1-resumen-del-flujo)
2. [Estructura de ramas](#2-estructura-de-ramas)
3. [Pipelines](#3-pipelines)
   - [docker-ci.yml — Validación de build](#31-docker-ciyml--validación-de-build)
   - [docker-staging.yml — Entrega a Staging](#32-docker-stagingyml--entrega-a-staging)
   - [docker-production.yml — Entrega a Producción](#33-docker-productionyml--entrega-a-producción)
4. [Configuración inicial en GitHub](#4-configuración-inicial-en-github)
   - [Secrets](#41-secrets)
   - [Environment de producción con aprobación manual](#42-environment-de-producción-con-aprobación-manual)
5. [Flujo de trabajo diario](#5-flujo-de-trabajo-diario)
6. [Imágenes Docker generadas](#6-imágenes-docker-generadas)
7. [Rollback](#7-rollback)

---

## 1. Resumen del flujo

```
feature/mi-cambio
       │
       │  git push + PR hacia develop
       ▼
   develop  ◄──── merge ────  [CI valida build]
       │
       │  merge automático dispara staging
       ▼
  Docker Hub: staging-latest / staging-<sha>
       │
       │  PR de develop → main
       ▼
    main  ◄──── merge ────  [CI valida build]
       │
       │  merge dispara producción (requiere aprobación manual)
       ▼
  Docker Hub: production-latest / production-<sha>
```

---

## 2. Estructura de ramas

| Rama | Propósito |
|---|---|
| `main` | Código en producción. Solo recibe merges desde `develop`. |
| `develop` | Integración y staging. Recibe merges desde ramas `feature/`. |
| `feature/*` | Desarrollo de funcionalidades o correcciones. Se crean desde `develop`. |

> Las ramas `feature/` nunca se mergean directamente a `main`.

---

## 3. Pipelines

### 3.1 `docker-ci.yml` — Validación de build

**Archivo:** `.github/workflows/docker-ci.yml`

**Se ejecuta cuando:** Se abre o actualiza un Pull Request hacia `develop` o `main`.

**Qué hace:**
1. Hace checkout del código de la rama origen del PR.
2. Verifica que `backend/Dockerfile` exista. Si no existe, falla con mensaje de error claro.
3. Ejecuta `docker build` usando el contexto `./backend`.
4. No hace push de ninguna imagen.

**Objetivo:** Garantizar que cualquier código que quiera entrar a `develop` o `main` pueda construirse correctamente como imagen Docker, antes de hacer merge.

**Comportamiento ante builds duplicados:** Si se hace un nuevo push a la misma rama mientras el pipeline está corriendo, el pipeline anterior se cancela automáticamente (`concurrency: cancel-in-progress: true`).

---

### 3.2 `docker-staging.yml` — Entrega a Staging

**Archivo:** `.github/workflows/docker-staging.yml`

**Se ejecuta cuando:** Se hace push (o merge de PR) a la rama `develop`.

**Qué hace:**
1. Hace checkout del código.
2. Verifica que `backend/Dockerfile` exista.
3. Hace login en Docker Hub usando los secrets `DOCKER_USERNAME` y `DOCKER_PASSWORD`.
4. Normaliza el nombre de la imagen a minúsculas.
5. Construye y publica la imagen con dos tags:
   - `staging-latest` → siempre apunta a la última entrega de staging.
   - `staging-<sha>` → tag inmutable que identifica exactamente qué commit generó esa imagen.

**Imagen publicada:**
```
<docker-user>/<nombre-repo>:staging-latest
<docker-user>/<nombre-repo>:staging-a3f1c9d...
```

---

### 3.3 `docker-production.yml` — Entrega a Producción

**Archivo:** `.github/workflows/docker-production.yml`

**Se ejecuta cuando:** Se hace push (o merge de PR) a la rama `main`.

**Qué hace:**
1. Hace checkout del código.
2. Verifica que `backend/Dockerfile` exista.
3. Espera aprobación manual si el Environment `production` está configurado en GitHub.
4. Hace login en Docker Hub usando los secrets `DOCKER_USERNAME` y `DOCKER_PASSWORD`.
5. Normaliza el nombre de la imagen a minúsculas.
6. Construye y publica la imagen con dos tags:
   - `production-latest` → siempre apunta a la última entrega de producción.
   - `production-<sha>` → tag inmutable que identifica exactamente qué commit generó esa imagen.

**Imagen publicada:**
```
<docker-user>/<nombre-repo>:production-latest
<docker-user>/<nombre-repo>:production-a3f1c9d...
```

**Comportamiento ante builds duplicados:** A diferencia de staging, `cancel-in-progress` está en `false` para producción. Si hay un deploy en curso, el nuevo esperará en cola en lugar de cancelar el anterior.

---

## 4. Configuración inicial en GitHub

### 4.1 Secrets

Los secrets almacenan las credenciales de Docker Hub de forma segura. Se configuran una sola vez.

1. Ir al repositorio en GitHub.
2. Navegar a **Settings → Secrets and variables → Actions**.
3. Crear los siguientes secrets con **New repository secret**:

| Secret | Valor |
|---|---|
| `DOCKER_USERNAME` | Tu nombre de usuario de Docker Hub. |
| `DOCKER_PASSWORD` | Tu Access Token de Docker Hub (recomendado) o tu contraseña. |

> Para generar un Access Token en Docker Hub: **Account Settings → Security → New Access Token**. Es más seguro que usar la contraseña directamente.

---

### 4.2 Environment de producción con aprobación manual

El pipeline de producción usa un GitHub Environment llamado `production`. Cuando está configurado con revisores requeridos, el pipeline se pausa antes de ejecutarse y espera aprobación explícita.

**Cómo configurarlo:**

1. Ir al repositorio en GitHub.
2. Navegar a **Settings → Environments → New environment**.
3. Nombre: `production` (debe ser exactamente este nombre).
4. Activar **Required reviewers**.
5. Agregar los usuarios o equipos que deben aprobar cada deploy a producción.
6. Guardar.

**Cómo se ve en la práctica:**

Cuando se hace merge a `main`, el pipeline aparece en estado de espera en la pestaña **Actions** del repositorio. Los revisores configurados reciben una notificación por email y deben aprobar desde GitHub para que el deploy continúe.

---

## 5. Flujo de trabajo diario

### Iniciar un cambio nuevo

```bash
# Siempre partir desde develop actualizado
git checkout develop
git pull origin develop

# Crear la rama de trabajo
git checkout -b feature/nombre-descriptivo
```

### Subir cambios y abrir PR

```bash
git add .
git commit -m "feat: descripción del cambio"
git push origin feature/nombre-descriptivo
```

Luego en GitHub: abrir Pull Request de `feature/nombre-descriptivo` → `develop`.

- El pipeline `docker-ci.yml` se ejecuta automáticamente.
- Si el build pasa, el PR puede mergearse.
- Si el build falla, se debe corregir antes de hacer merge.

### Merge a develop → Staging

Al hacer merge del PR a `develop`:

- El pipeline `docker-staging.yml` se ejecuta automáticamente.
- La imagen de staging queda publicada en Docker Hub.
- Validar que la imagen funciona correctamente en el ambiente de staging.

### Promover a producción

Cuando staging está validado:

```bash
# Abrir PR de develop → main en GitHub
```

- El pipeline `docker-ci.yml` valida el build nuevamente.
- Al hacer merge, `docker-production.yml` se dispara.
- Si el Environment `production` tiene aprobación manual configurada, el pipeline espera.
- El revisor aprueba desde GitHub → la imagen se publica en producción.

---

## 6. Imágenes Docker generadas

Cada entrega genera dos tags en Docker Hub:

| Tag | Descripción |
|---|---|
| `staging-latest` | Última imagen de staging. Se sobreescribe en cada entrega. |
| `staging-<sha>` | Imagen inmutable ligada al commit exacto. Nunca se sobreescribe. |
| `production-latest` | Última imagen de producción. Se sobreescribe en cada entrega. |
| `production-<sha>` | Imagen inmutable ligada al commit exacto. Nunca se sobreescribe. |

**Cómo identificar qué imagen corresponde a qué código:**

El `<sha>` es el SHA del commit de Git. Puedes encontrarlo en:
- La lista de commits del repositorio en GitHub.
- La pestaña **Actions** → detalle del workflow run.
- Docker Hub → pestaña **Tags** de tu repositorio.

Ejemplo:
```
mi-usuario/poc-indesign-flowdesign:staging-a3f1c9d
```

---

## 7. Rollback

Si una imagen de producción presenta problemas, se puede revertir a cualquier imagen anterior usando su tag con SHA.

### Identificar la imagen anterior

1. Ir a Docker Hub → tu repositorio → pestaña **Tags**.
2. Buscar el tag `production-<sha>` de la entrega anterior a la que quieres volver.
3. Alternativamente, ir al historial de commits en GitHub e identificar el SHA del commit estable anterior.

### Ejecutar el rollback

```bash
# Descargar la imagen de la versión anterior
docker pull <docker-user>/<nombre-repo>:production-<sha-anterior>

# Redeployar con esa imagen en tu servidor o plataforma
docker run <docker-user>/<nombre-repo>:production-<sha-anterior>
```

Si usas `docker-compose`, actualizar el tag de la imagen en el archivo y aplicar:

```bash
docker compose up -d
```

> El tag `production-latest` no es confiable para rollback porque apunta siempre a la última versión. Usar siempre el tag con SHA para revertir a una versión específica.
