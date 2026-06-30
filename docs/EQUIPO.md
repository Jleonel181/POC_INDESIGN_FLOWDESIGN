# Equipo — Permisos y Colaboración

## Cómo Agregar Usuarios a GitHub

### Paso 1: Invitar al repositorio

1. Ve a tu repositorio en GitHub
2. **Settings** → **Collaborators and teams**
3. Click en **Add people**
4. Escribe el username de GitHub del usuario
5. Selecciona el nivel de acceso:

| Rol | Permisos |
|---|---|
| **Pull requests only** | Pueden hacer fork y PR (sin acceso directo) |
| **Triage** | Pueden revisar PRs pero no mergear |
| **Write** | Pueden mergear PRs y hacer push directo |
| **Admin** | Acceso total (crear branches, secrets, etc.) |

### Paso 2: Configurar Reviewers Requeridos

Para que los cambios requieran aprobación antes de mergear:

1. **Settings** → **Branches**
2. Selecciona rama (`develop` o `main`)
3. **Add rule**
4. En **Require a pull request before merging:**
   - Checkear **Require approvals**
   - Seleccionar número de revisores (recomendado: 1)
   - Checkear **Require review from Code Owners** (opcional)

### Paso 3: Crear CODEOWNERS (recomendado)

Archivo `.github/CODEOWNERS` (repositorio raíz):

```
# Estructura de propietarios por módulo

# Backend
/backend/                     @leonardo
/esb/                         @leonardo

# Frontend
/frontend/                    @designer

# CI/CD
.github/                      @leonardo

# Documentación
docs/                         @equipo

# Pull request requests
*                             @leonardo
```

Cuando alguien abre un PR que toca estos archivos, automáticamente se asignan los reviewers.

---

## Flujo de Revisión de PRs

### 1. Abrir un PR

Después de hacer push a una rama feature:

```bash
git push origin feature/mi-cambio
```

GitHub sugiere automáticamente abrir un PR. Hacer click en **Compare & pull request**.

### 2. Descripción del PR

Incluir:
- Qué cambio se hizo
- Por qué se hizo
- Cómo testear

**Template (crear `.github/pull_request_template.md`):**

```markdown
## Descripción
Explica brevemente qué hizo este PR

## Tipo de cambio
- [ ] Bug fix
- [ ] Feature nueva
- [ ] Breaking change
- [ ] Documentación

## Cómo testear
Pasos para validar el cambio

## Screenshots (si aplica)
Imágenes o videos

## Checklist
- [ ] Mi código sigue el style del proyecto
- [ ] He testeado localmente
- [ ] No hay console.log o debug code
- [ ] La documentación está actualizada
```

### 3. CI/CD Automático

GitHub Actions se ejecuta automáticamente:

```
Abrir PR
  ↓
docker-ci.yml se dispara
  ↓
Valida que el build funciona
  ↓
Si pasa: permite mergear
Si falla: bloquea mergear
```

### 4. Revisión de Código

Los reviewers pueden:
- Hacer comentarios en líneas específicas
- Solicitar cambios
- Aprobar

### 5. Mergear

Una vez aprobado y CI pasa:
- Click en **Merge pull request**
- Seleccionar estrategia:
  - **Merge commit** (crea merge commit, recomendado)
  - **Squash and merge** (aplasta commits)
  - **Rebase and merge** (rebase limpio)

---

## Gestión de Branches

### Convención de Nombres

| Tipo | Ejemplo | Propósito |
|---|---|---|
| Feature | `feature/agregar-diagramador` | Trabajo nuevo |
| Bugfix | `bugfix/corregir-layout` | Correcciones |
| Hotfix | `hotfix/emergencia-cierre` | Producción urgente |
| Release | `release/1.0.0` | Preparar release |

### Eliminar Branch Después del Merge

Después de mergear un PR, GitHub ofrece eliminar el branch automáticamente. **Aceptar siempre.**

Si quieres eliminar local:
```bash
git branch -d feature/mi-cambio
```

---

## Protecciones de Rama

Configurar en **Settings** → **Branches** → **Branch protection rules**:

### main
- ✅ Require pull request reviews before merging (1 revisor)
- ✅ Require status checks to pass before merging (docker-ci.yml, docker-production.yml)
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require branches to be up to date before merging
- ✅ Include administrators (los admins también cumplen las reglas)

### develop
- ✅ Require pull request reviews before merging (1 revisor)
- ✅ Require status checks to pass before merging (docker-ci.yml, docker-staging.yml)
- ⚪ Dismiss stale pull request approvals (opcional)
- ⚪ Require branches to be up to date (menos estricto que main)

---

## Secretos y Credenciales

### Agregar Secretos en GitHub

1. **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Nombre: `DOCKER_USERNAME`
4. Valor: tu usuario de Docker Hub

**Secretos necesarios:**
- `DOCKER_USERNAME` — Usuario Docker Hub
- `DOCKER_PASSWORD` — Token de acceso Docker Hub (no contraseña)

Los secretos se inyectan en los workflows y no son visibles en los logs.

### En Workflows

```yaml
- name: Login to Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKER_USERNAME }}
    password: ${{ secrets.DOCKER_PASSWORD }}
```

---

## Environment y Aprobaciones en Producción

### Crear Environment: production

1. **Settings** → **Environments** → **New environment**
2. Nombre: `production`
3. En **Deployment branches and tags:**
   - Seleccionar `main` (solo main puede desplegar a prod)
4. En **Required reviewers:**
   - Agregar usuarios que deben aprobar cada deploy
5. Guardar

### Cómo funciona en el workflow

```yaml
jobs:
  deploy-production:
    environment: production  # Espera aprobación
    steps:
      - name: Build and Push
        ...
```

Cuando se hace merge a `main`:
1. El workflow se ejecuta pero se pausa en `deploy-production`
2. GitHub notifica a los reviewers por email
3. Ellos van a **Actions** → **workflow run** y aprueban
4. El deploy continúa

---

## Notificaciones

### Configurar Notificaciones de GitHub

1. Click tu perfil → **Settings** → **Notifications**
2. Email preferences:
   - ✅ Comments on issues and pull requests
   - ✅ Pull request reviews
   - ✅ Pull request pushes
   - ✅ Status checks
3. Guardar

Así recibirás alertas cuando:
- Alguien comenta en tu PR
- Alguien solicita cambios
- Los tests fallan

---

## Buenas Prácticas

### ✅ Hacer

- Commits atómicos y con mensajes claros
  ```bash
  git commit -m "feat: agregar endpoint de diagramador"
  ```

- PRs pequeños (< 500 líneas idealmente)

- Descripción clara en cada PR

- Esperar aprobación antes de mergear

- Correr tests localmente antes de pushear

### ❌ No Hacer

- Commits con mensajes genéricos ("fix", "update")
- PRs gigantes con 10 cambios distintos
- Mergear sin aprobación
- Pushear código con `console.log()` o comentarios de debug
- Hacer merge a `main` sin que pase CI

---

## Flujo Completo de un Feature

```
1. Crear rama feature desde develop
   git checkout develop
   git pull origin develop
   git checkout -b feature/mi-cambio

2. Hacer cambios, commits locales
   git add .
   git commit -m "feat: descripción"

3. Pushear a GitHub
   git push origin feature/mi-cambio

4. Abrir PR en GitHub
   - Seleccionar: feature/mi-cambio → develop
   - Llenar descripción
   - Agregar labels si existen

5. CI corre automáticamente (docker-ci.yml)
   - Si pasa: permite mergear
   - Si falla: pedir cambios

6. Esperar revisión
   - Reviewer revisa código
   - Solicita cambios o aprueba

7. Si hay cambios, hacer fix y pushear
   git add .
   git commit -m "fix: comentarios del revisor"
   git push origin feature/mi-cambio

8. Merger después de aprobación
   - Click "Merge pull request" en GitHub
   - Seleccionar "Merge commit"
   - Confirmar

9. Eliminar branch
   - GitHub ofrece "Delete branch"
   - Aceptar

10. Después del merge a develop
    - docker-staging.yml se dispara
    - Imagen staging se publica en Docker Hub
    - Validar en staging

11. Cuando staging está listo
    - Abrir PR: develop → main
    - Mismo proceso
    - Al mergear, docker-production.yml se dispara
    - Requiere aprobación de reviewer de producción
    - Si se aprueba, imagen production se publica
```

---

## Roles Recomendados

| Rol | Usuarios | Permisos |
|---|---|---|
| **Owner** | @leonardo | Admin total, gestión de equipo |
| **Core** | - | Write, pueden mergear PRs, manejan releases |
| **Reviewer** | - | Triage, revisar código, no mergear |
| **Contributor** | - | Pull requests only, sin acceso directo |

---

## Troubleshooting

### Error: "You don't have permission to push to this repository"

- No tienes acceso de Write al repo
- Contactar al Owner para que te agregue
- O hacer fork y abrir PR desde el fork

### Error: "Branch is protected, requires reviews before merging"

- Esperar a que un revisor apruebe el PR
- No puedes mergear tu propio PR sin otra aprobación

### ¿Cómo revierto un merge accidental?

```bash
# Si fue hace poco (aún en main)
git revert -m 1 <commit-hash>
git push origin main
```

---

**Última actualización:** 2026-01-10
