# Fidelidad de lectura y escritura

Este documento registra **cuánta información sobrevive** al ciclo de abrir un IDML y volver a escribirlo, cómo se mide, y qué se ha corregido.

Existe porque la afirmación «roundtrip perfecto» que había en el `README.md` no era cierta, y no había forma de saberlo: no existía ninguna medición. Ahora existe, y las cifras están aquí.

## Resumen en una frase

La librería **no corrompe valores ni pierde elementos**, pero **descarta atributos que no modela** y, en algunos tipos, **reagrupa los hijos por tipo** en lugar de conservar su orden.

## Cómo estaba antes

Antes de introducir el arnés de fidelidad, el estado real era desconocido. Concretamente:

- El único test de fidelidad, `TestGoldenRoundtrip_ExampleIDML`, tenía un `t.Skip` incondicional: nunca se ejecutaba.
- Los documentos de referencia de InDesign no estaban versionados, así que ni siquiera había con qué comparar en un clon del repositorio.
- `ARCHITECTURE.md` describía las pruebas de golden files como «preservación byte a byte del ZIP». Eso no es lo que hacen ni lo que pueden hacer: reescribir un IDML nunca da un ZIP idéntico, porque cambian las marcas de tiempo y el orden de atributos.
- `NewFromTemplate()` emitía un paquete de 9 archivos cuyo `designmap.xml` no declaraba ninguna referencia `idPkg:Spread`, es decir, **un documento sin ninguna página**. Nadie había comprobado si InDesign abría su salida.

## El arnés

`pkg/idml/golden_test.go` aplica a cada archivo XML del corpus el ciclo **parseo → serialización → comparación estructural** y reporta las diferencias por categoría.

Las diferencias **se registran, no hacen fallar el test**. Es deliberado: hoy son la medición que las tareas de modelo tienen que llevar a cero. La tarea de cierre del plan es la que convierte una diferencia en fallo, cuando la cifra ya sea cero.

### Ejecutarlo

```bash
# Resumen por origen y total agregado
go test ./pkg/idml/ -run TestGoldenRoundtrip_ExampleIDML -v -count=1 2>&1 | grep 'resumen \['

# Cifras de referencia: sin el tope de 100 diferencias por archivo
IDMLLIB_MAX_DIFFS=0 go test ./pkg/idml/ -run TestGoldenRoundtrip_ExampleIDML -v -count=1 2>&1 | grep 'resumen \['

# Detalle de un archivo concreto
IDMLLIB_MAX_DIFFS=0 go test ./pkg/idml/ -run TestGoldenRoundtrip_ExampleIDML -v -count=1 2>&1 | grep -A6 'Spread_uce7'
```

**Importante al leer las cifras:** con el tope de 100 por archivo puesto, el reparto por categoría de un archivo truncado depende de en qué punto cortó, así que **no es comparable entre ejecuciones**. Cualquier comparación se hace con `IDMLLIB_MAX_DIFFS=0`. El propio arnés lo advierte cuando hay archivos truncados.

### Variables de entorno

Cada elemento del corpus se resuelve desde una constante del paquete de prueba y se puede apuntar a otra copia sin recompilar:

| Variable | Qué sustituye |
|---|---|
| `IDMLLIB_REFERENCE_DIR` | el Documento_Referencia descomprimido |
| `IDMLLIB_IMAGES_FIXTURE` | el Archivo_Evidencia_Imagenes |
| `IDMLLIB_PLAIN_IDML` | `testdata/plain.idml` |
| `IDMLLIB_EXAMPLE_IDML` | `testdata/example.idml` |
| `IDMLLIB_TRIPPLE_IDML` | `testdata/tripple.idml` |
| `IDMLLIB_MAX_DIFFS` | el tope de diferencias por archivo; `0` = sin tope |

Para medir contra **cualquier IDML propio**, con ruta absoluta:

```bash
IDMLLIB_MAX_DIFFS=0 IDMLLIB_IMAGES_FIXTURE="/ruta/absoluta/a/tu.idml" \
  go test ./pkg/idml/ -run 'TestGoldenRoundtrip_ExampleIDML/archivo_evidencia_imagenes' -v -count=1
```

Con una ruta **relativa** el test se salta en silencio y termina en `ok`, porque la resuelve desde `pkg/idml/`. Parece que pasó cuando no ha medido nada.

### Comportamiento ante un corpus incompleto

- **Elemento ausente:** se omite ese origen con la ruta absoluta esperada, sin fallar y sin arrastrar a los demás. Es una red de diagnóstico, no el caso normal.
- **Directorio presente pero sin XML:** falla indicando la ruta recorrida. Eso no es «no lo tengo», es «algo está mal configurado», y pasar de largo haría creer que la prueba midió algo.

## El corpus

Cinco documentos, todos versionados en `testdata/`:

| Elemento | Forma | Por qué está |
|---|---|---|
| `documento_referencia/` | descomprimido, 43 XML | Página de periódico real. Es el documento que define el alcance. Se versiona en texto porque no existe un `.idml` original suyo, y así git guarda diferencias y se puede hacer `grep` |
| `archivo_evidencia_imagenes.idml` | paquete | **Única fuente de verdad del formato de imagen embebida.** Su spread lleva dos imágenes embebidas y una enlazada de control |
| `plain.idml` | paquete | Documento casi vacío. Es el **oráculo del conjunto mínimo de archivos**: 13 entradas, 1 página, 1 marco de texto |
| `example.idml` | paquete | Más recursos y estilos |
| `tripple.idml` | paquete | Tres spreads |

Los dos primeros se incorporaron al crear el arnés; los tres últimos ya estaban en `testdata/` y **el arnés no los recorría**.

Ese detalle costó dos lecciones: medir contra dos documentos resultó insuficiente **dos veces**, porque cada documento contiene formas que los otros no. Un criterio del tipo «cero diferencias de categoría X» es falso si el corpus no contiene la forma que falla. De ahí la regla: **antes de dar por bueno un cero, comprobar que el corpus contiene el caso**.

El Documento_Referencia y el Archivo_Evidencia_Imagenes son regenerables solo en la máquina del usuario: su documento fuente de InDesign se conserva fuera de git.

## Línea base medida

Con `IDMLLIB_MAX_DIFFS=0`:

| Origen | Archivos | Limpios | Con diferencias | Sin parsear | atributo-ausente | orden-elementos | texto-distinto |
|---|---|---|---|---|---|---|---|
| `documento_referencia` | 43 | 6 | 34 | 3 | 3868 | 35 | 2 |
| `archivo_evidencia_imagenes` | 11 | 6 | 4 | 1 | 949 | 0 | 0 |
| `plain` | 12 | 7 | 4 | 1 | 730 | 0 | 0 |
| `example` | 26 | 14 | 9 | 3 | 4031 | 10 | 0 |
| `tripple` | 44 | 20 | 21 | 3 | 4417 | 26 | 0 |
| **TOTAL** | **136** | **53** | **72** | **11** | **13995** | **71** | **2** |

**Cero fallos** de parseo o serialización en los cinco documentos.

Y lo que **no** aparece importa tanto como lo que aparece. Estas categorías están a **0** en todo el corpus:

- `atributo-valor-distinto` — ningún valor se corrompe
- `atributo-sobrante` — no se inventan atributos
- `elemento-ausente` y `elemento-sobrante` — no se pierde ni se duplica ningún elemento
- `etiqueta-distinta` — ningún elemento se renombra

De ahí la conclusión: **solo hay dos defectos**, y el grande es el de los atributos. 13995 frente a 71.

## Los dos defectos

### 1. Atributos descartados

`encoding/xml` asigna a los campos del struct los atributos que el struct declara y **descarta el resto en silencio**. Un `Rectangle` del Documento_Referencia trae 33 atributos y el modelo declara 21, con 22 emitidos tras aplicar `omitempty`: **11 se pierden en cada ciclo**.

La solución es un campo comodín `OtherAttrs []xml.Attr` donde va lo que no corresponde a ningún campo declarado. El reparto lo hacen dos funciones genéricas en `internal/xmlutil/attrs.go`, `UnmarshalAttrs` y `MarshalAttrs`, que leen del propio struct qué campos declara. Escribirlas una vez evita unas 300 ramas escritas a mano en los doce tipos que lo necesitan.

**Estado:** las funciones existen y están probadas. Conectarlas a los tipos del modelo es trabajo pendiente, así que la cifra de 13995 sigue intacta.

### 2. Hijos reagrupados por tipo

Un struct de Go declara sus hijos en campos por tipo, y al serializar los emite en el orden de esos campos. Pero en el XML de entrada venían intercalados. El resultado tiene los mismos elementos en otro orden.

La solución es `internal/xmlutil/childorder.go`: un registro de la secuencia leída que se reproduce al emitir. **No es un contenedor**: el contenido sigue viviendo en su campo por tipo, que es su fuente de verdad; el registro solo guarda el orden. Eso permitió arreglarlo sin tocar los 158 accesos a los campos de `Document`.

`ChildOrder.Replay` emite en dos pasadas: primero el orden registrado, y después los hijos que el registro no menciona. La segunda pasada es lo que hace el mecanismo seguro: sin ella, agregar un hijo sin actualizar el registro lo haría **desaparecer** al serializar, y perder un elemento es peor que emitirlo en una posición rara. Tiene un efecto secundario útil: un modelo construido desde cero, sin registro, cae entero en la segunda pasada y sale en el orden de los campos, que es el comportamiento anterior. No hay dos caminos que mantener.

**Aplicado a:** `designmap.xml` (el elemento `Document`), `Resources/Styles.xml` y `Resources/Graphic.xml`.

**Pendiente en:** los spreads, las stories, y dos tipos más —`Properties`, que es un tipo compartido usado en 32 sitios, y `ObjectStyleGroup`— que se descubrieron al ampliar el corpus.

**Techo declarado:** quien agregue un elemento a un campo por tipo sin actualizar el registro verá ese elemento emitido al final de su grupo, no en una posición arbitraria. Está marcado con un comentario `ponytail:` en el código, con su vía de mejora.

## Las categorías de diferencia

`internal/xmlutil/compare.go` clasifica cada diferencia en una categoría con nombre, de modo que se pueda contar y filtrar una por una:

`atributo-ausente`, `atributo-valor-distinto`, `atributo-sobrante`, `elemento-ausente`, `elemento-sobrante`, `orden-elementos-distinto`, `texto-distinto`, `etiqueta-distinta`, `namespace-distinto`.

Dos decisiones del comparador que conviene conocer:

**El reordenamiento se detecta como tal.** Si los hijos de un elemento son los mismos en otro orden, se emite **una** diferencia en el nodo padre con las dos secuencias, no una por cada posición desplazada. Antes, un desfase de una posición encadenaba una diferencia por cada hijo posterior: el `designmap.xml` del Archivo_Evidencia_Imagenes producía 96 diferencias, de las que 50 eran etiquetas desplazadas. Un defecto, 96 líneas de reporte, y ninguna nombraba el problema real.

**Los hijos se emparejan por nombre y ordinal, no por posición.** Es lo que hace fiables las cifras de atributos: sin ese emparejamiento, tras un reordenamiento la comparación queda desalineada y las diferencias de atributo que reporta están comparando elementos que no se corresponden.

**`SortElements`** lista los elementos cuyos hijos se ordenan antes de comparar, porque ahí el orden se declara irrelevante: `FontFamily`, `Color`, `ParagraphStyle`, `CharacterStyle` y `ObjectStyle`. Consecuencia práctica: el desorden de los hijos de los `ObjectStyle` existe en el corpus pero el arnés no lo ve. Queda abierto si InDesign comparte ese criterio.

## Generar un documento abrible

`NewFromTemplate()` produce un paquete de **13 entradas**, el mismo conjunto que `testdata/plain.idml`, que es una exportación de InDesign de la misma forma.

**Verificado a mano en Adobe InDesign:** dos sondas A4, una de 1 columna y otra de 5, abren sin aviso de daño ni petición de recuperación, y muestran la página con su marco de texto dentro de los márgenes. Abren también en **Affinity Publisher**, así que la forma emitida no depende de una tolerancia particular de InDesign.

De esa verificación salió un dato que cerraba una incógnita: **InDesign respeta el `ColumnsPositions` emitido y no lo recalcula.** Con 5 columnas y medianil 12 se emitieron huecos de 12 y se miden 12. Por lo tanto el cálculo directo a partir del ancho útil, el número de columnas y el medianil es la derivación correcta.

Para regenerar las sondas:

```bash
IDMLLIB_PROBE_DIR=~/Desktop/sondas_idml go test ./pkg/idml/ -run SondaParaInDesign -v -count=1
```

## Lo que sigue sin estar verificado

- La fidelidad de un IDML **existente** está lejos de cero: 13995 atributos y 71 desórdenes.
- Los archivos bajo `MasterSpreads/` **no se parsean**: se copian tal cual. Son los 11 «copiados sin parsear» del resumen.
- Solo se ha comprobado que InDesign abre un documento **generado desde cero**. No se ha comprobado que abra un documento **existente reescrito** por la librería.
- El corpus son cinco documentos de una misma versión de InDesign. No hay especificación oficial vigente contra la que validar: Adobe retiró la de IDML de su sitio.
