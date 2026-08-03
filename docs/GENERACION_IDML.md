# Generación IDML — Arquitectura

## Flujo de datos

```
┌──────────────────────────────────────────────────────────────────┐
│  PostgreSQL                                                       │
│  (ediciones, páginas, pautas)                                     │
└───────────────────────────────┬──────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│  Backend Node.js (:3001)                                          │
│                                                                   │
│  GenerateEditionLayoutUseCase                                     │
│    → GridLayoutCalculator (grilla → bounds mm)                    │
│    → LayoutValidator (sin solapamientos, dentro de grilla)        │
│    → assertLayoutContract (validación del contrato)               │
│                                                                   │
│  Resultado: Layout Contract JSON                                  │
│  GET /api/layout/:editionId/idml                                  │
│    → spawn: idmlgen [-folio] -out - < contract.json               │
│    → responde con el IDML como descarga                           │
└───────────────────────────────┬──────────────────────────────────┘
                                │ stdin: JSON
                                │ stdout: bytes IDML
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│  cmd/idmlgen (Go, binario estático)                               │
│                                                                   │
│  1. Lee Layout Contract JSON por stdin                            │
│  2. Valida: dimensiones > 0, unit=mm, origin=top-left             │
│  3. NewFromTemplate (dimensiones, márgenes, columnas)             │
│  4. Por cada pauta: AddTextFrame con bounds mm→pt                 │
│  5. Si -folio: agrega TextFrame de pie con <?ACE 18?>             │
│  6. Registra stories en designmap.xml                             │
│  7. WriteTo(stdout) o Write(archivo)                              │
│                                                                   │
│  Códigos de salida: 0=OK, 1=entrada inválida, 2=fallo interno    │
└───────────────────────────────┬──────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│  Archivo IDML                                                     │
│  → Se abre en Adobe InDesign o Affinity Publisher                 │
│  → Contiene los marcos posicionados y el folio automático         │
└──────────────────────────────────────────────────────────────────┘
```

## Contratos

### Layout Contract JSON (`contracts/layout.schema.json`)

Es el formato intermedio entre el dominio editorial (pautas, cuadros, ediciones) y el
generador IDML genérico. Lo produce el backend y lo consume `idmlgen`.

Campos clave de `edition`:

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `ancho_mm` | number | Ancho de página en mm |
| `alto_mm` | number | Alto de página en mm |
| `cuadros_ancho` | number | Columnas de la grilla |
| `cuadros_alto` | number | Filas de la grilla |
| `facing_pages` | boolean | Páginas encaradas (spreads) |
| `margen_*_mm` | number | Los cuatro márgenes en mm |

Cada `pauta` incluye `indesignBounds` con `topMm`, `leftMm`, `bottomMm`, `rightMm` ya
calculados por `GridLayoutCalculator`.

### JSON de idmlgen (implícito)

`cmd/idmlgen` consume directamente el Layout Contract. No hay un segundo schema; la
traducción de mm a puntos InDesign (1pt = 25.4/72 mm) se hace dentro del binario.

Si en el futuro idmllib se usara desde otro sistema editorial que no hable de "pautas",
se introduciría un schema genérico intermedio (páginas, marcos, bounds en pt). Hoy no
hace falta porque el único consumidor es este proyecto.

## Decisiones de diseño

1. **Desacoplamiento por proceso.** El backend invoca a `idmlgen` como un subproceso con
   stdin/stdout. No hay binding de Go en Node.js, ni servicio HTTP en Go. Ventajas:
   - El binario se puede actualizar o recompilar sin tocar el backend.
   - Se puede probar en aislamiento: `cat layout.json | idmlgen > test.idml`.
   - El backend no necesita saber de IDML, XML ni ZIP.

2. **Folio determinista.** El encabezado/folio se genera como código, no como instrucción
   a una IA. El marcador `<?ACE 18?>` es un auto page number de InDesign que se resuelve
   al abrir el archivo.

3. **Una página por ahora.** Multi-página requiere crear spreads adicionales en el
   paquete IDML. La Tarea 18 del plan de idmllib define `AddSpread`; cuando se cierre,
   `idmlgen` iterará las páginas del contrato.

4. **Sin IA para la diagramación base.** La IA queda como opción complementaria para
   sugerencias de layout cuando no hay pauta predefinida. La ejecución del layout ya
   definido es 100% determinista.

## Cómo compilar el binario

```bash
cd idmllib
go build -o bin/idmlgen ./cmd/idmlgen
```

## Cómo probar manualmente

```bash
# Generar con folio a archivo
cat /tmp/layout_test.json | ./bin/idmlgen -folio -out /tmp/edicion.idml

# Generar sin folio a stdout
curl http://localhost:3001/api/layout/1 | ./bin/idmlgen > edicion.idml

# Verificar
unzip -l /tmp/edicion.idml
```

## Limitaciones conocidas

- Solo genera la primera página del contrato (multi-página pendiente de Tarea 18).
- No soporta imágenes (pendiente de Tarea 20).
- Los estilos tipográficos del folio (Popular Bold 10pt) no se aplican desde código;
  se necesitaría declarar un ParagraphStyle en Resources/Styles.xml.
- `TextFramePreference` se emite como XML crudo via comodín `OtherElements`, no como
  struct tipado (pendiente de Tarea 13).
