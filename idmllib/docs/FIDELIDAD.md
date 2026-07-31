# Fidelidad de lectura y escritura

Este documento registra **cuánta información sobrevive** al ciclo de abrir un IDML y volver a escribirlo, cómo se mide, y qué se ha corregido.

Existe porque la afirmación «roundtrip perfecto» que había en el `README.md` no era cierta, y no había forma de saberlo: no existía ninguna medición. Ahora existe, y las cifras están aquí.

## Resumen en una frase

La librería **no pierde nada de contenido** —ni un atributo, ni un elemento, ni un carácter— pero en algunos tipos **reagrupa los hijos por tipo** en lugar de conservar su orden documental.

## Estado actual

Sobre el corpus de cinco documentos, 136 archivos XML:

| Categoría de diferencia | Cantidad | Estado |
|---|---|---|
| `atributo-ausente` | 0 | cerrada con guarda |
| `atributo-valor-distinto` | 0 | cerrada con guarda |
| `atributo-sobrante` | 0 | cerrada con guarda |
| `elemento-ausente` | 0 | cerrada con guarda |
| `elemento-sobrante` | 0 | cerrada con guarda |
| `etiqueta-distinta` | 0 | cerrada con guarda |
| `namespace-distinto` | 0 | cerrada con guarda |
| `texto-distinto` | 0 | cerrada con guarda |
| **`orden-elementos-distinto`** | **71** | **abierta** |

Cero fallos de parseo o serialización. 66 de 136 archivos sin ninguna diferencia, y dos documentos completos —`plain.idml` y `archivo_evidencia_imagenes.idml`— sin una sola.

**«Cerrada con guarda»** significa que `assertCategoriasCerradas`, en `pkg/idml/golden_test.go`, **hace fallar el arnés** si esa categoría vuelve a aparecer. Cada guarda se verificó rompiendo el modelo a propósito y comprobando que el test falla, porque un test que nunca se ha visto fallar no es una guarda.

Las categorías se van cerrando **a medida que llegan a cero**, no todas de golpe al final. Llevar los atributos de 13.995 a 0 costó cuatro tareas, y durante ese tiempo no existía nada que fallara si se rompía: el arnés solo escribía cifras en un log.

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

### Recetas de medición

Estas tres son las que se han vuelto a necesitar una y otra vez. Quedan aquí escritas para no rederivarlas.

**Quién tiene las diferencias de orden que quedan.** Agrupa los desórdenes por elemento padre, que es lo que dice qué tarea los cierra:

```bash
IDMLLIB_MAX_DIFFS=0 go test ./pkg/idml/ -run TestGoldenRoundtrip_ExampleIDML -v -count=1 2>&1 \
  | grep 'orden-elementos-distinto\]' \
  | sed 's/.*\. //;s/ \[orden.*//;s/\[[0-9]*\]//g' \
  | sort | uniq -c | sort -rn
```

**Si un atributo puede faltar, antes de quitarle `omitempty`.** La decisión depende de dos cifras distintas: cuántas veces viene con valor vacío, y cuántas veces **no viene**. Si puede faltar, `omitempty` no se toca:

```bash
python3 - testdata/documento_referencia TextFrame ItemTransform <<'PY'
import sys, glob, xml.etree.ElementTree as ET
from collections import Counter
d, elem, attr = sys.argv[1], sys.argv[2], sys.argv[3]
c = Counter()
for f in glob.glob(d + "/**/*.xml", recursive=True):
    try: raiz = ET.parse(f).getroot()
    except ET.ParseError: continue
    for e in raiz.iter():
        if e.tag.split('}')[-1] != elem: continue
        c["total"] += 1
        c["ausente" if attr not in e.attrib else ("vacio" if e.attrib[attr] == "" else "con-valor")] += 1
print(elem, attr, dict(c))
PY
```

Ejemplos ya medidos, útiles como control de que la receta funciona: `TextFrame/ItemTransform` da 33 total y 33 con valor, nunca falta. `Story/StoryTitle` da 90 total, 30 con valor y **60 ausente**: a ese no se le quita `omitempty`.

**Dónde están los atributos con prefijo de namespace.** Importa porque `encoding/xml` los corrompe al reemitirlos, así que el comodín `xml:",any,attr"` no debe alcanzarlos:

```bash
python3 - testdata/documento_referencia <<'PY'
import sys, glob, re
from collections import Counter
d = sys.argv[1]
tag_re = re.compile(r'<([A-Za-z_][\w.\-]*)([^>]*)>', re.S)
pref_re = re.compile(r'(?<![\w:.\-])([A-Za-z_][\w.\-]*):([A-Za-z_][\w.\-]*)\s*=\s*"')
por_attr, por_arch = Counter(), Counter()
for f in sorted(glob.glob(d + "/**/*.xml", recursive=True)):
    for _, attrs in tag_re.findall(open(f, encoding="utf-8", errors="replace").read()):
        for p, local in pref_re.findall(attrs):
            if p == "xmlns": continue          # una declaracion no es un atributo
            por_attr[p + ":" + local] += 1; por_arch[f.split("/")[-1]] += 1
print("total:", sum(por_attr.values()))
for t, c in (("por atributo", por_attr), ("por archivo", por_arch)):
    print("\n" + t + ":"); [print(f"  {v:5d}  {k}") for k, v in c.most_common(8)]
PY
```

Cuidado con dos trampas de esta última, que ya dieron una cifra mal: hay que **excluir las declaraciones `xmlns:`**, que no son atributos, y el resultado **no está solo en `META-INF/metadata.xml`**. Lo medido en el Documento_Referencia: 501 atributos con prefijo, todos `rdf:parseType`, `xml:lang`, `x:xmptk` y `rdf:about`; 345 en `metadata.xml` y **156 repartidos por spreads y stories**.

Esos 156 no son un riesgo, y la razón conviene saberla: están dentro de un `<![CDATA[ ]]>`, en `MetadataPacketPreference > Properties > Contents`, donde InDesign incrusta el paquete XMP de cada imagen. Para `encoding/xml` eso es texto, no atributos, así que ningún comodín los alcanza. **El CDATA sobrevive al ciclo completo**, comprobado sobre `Spread_uce7.xml`: 286.839 bytes de entrada y 286.830 de salida, con el bloque intacto y sin escapar.

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

## Progreso medido

Con `IDMLLIB_MAX_DIFFS=0`. La columna «inicio» es la primera medición, cuando el arnés empezó a correr:

| Origen | Archivos | atributos: inicio → ahora | texto: inicio → ahora | orden: inicio → ahora |
|---|---|---|---|---|
| `documento_referencia` | 43 | 3868 → **0** | 2 → **0** | 35 → 31 |
| `archivo_evidencia_imagenes` | 11 | 949 → **0** | 0 | 0 |
| `plain` | 12 | 730 → **0** | 0 | 0 |
| `example` | 26 | 4031 → **0** | 0 | 10 → 6 |
| `tripple` | 44 | 4417 → **0** | 0 | 26 → 17 |
| **TOTAL** | **136** | **13.995 → 0** | **2 → 0** | **71 → 54** |

Archivos sin ninguna diferencia: **53 → 74** de 136. `archivo_evidencia_imagenes` y `plain` quedan completos.

Los 17 desórdenes cerrados hasta ahora son 5 de los elementos de página de los spreads (Tarea 7) y 12 de los tipos compartidos `Properties` y `ObjectStyleGroup` (Tarea 2e).

**Cero fallos** de parseo o serialización en los cinco documentos, antes y ahora.

## Los defectos

### 1. Atributos descartados — RESUELTO

`encoding/xml` asigna a los campos del struct los atributos que el struct declara y **descarta el resto en silencio**. Un `Rectangle` del Documento_Referencia trae 33 atributos y el modelo declaraba 21, con 22 emitidos tras aplicar `omitempty`: 11 se perdían en cada ciclo.

La solución es un campo comodín. Resultó que **`encoding/xml` ya lo implementa** con la etiqueta `xml:",any,attr"`, en las dos direcciones: al leer recoge solo los atributos que no encajaron en otro campo, en su orden; al escribir los emite después de los declarados. Está aplicado a **78 tipos**.

**Límite conocido:** `encoding/xml` corrompe los atributos con **prefijo de namespace** al re-emitirlos, convirtiendo `xmlns:idPkg="..."` en `xmlns:_xmlns="xmlns" _xmlns:idPkg="..."`. No aplica al modelo: se inspeccionaron 24.565 elementos de 351 tipos en el corpus completo y los 826 atributos con prefijo que existen están todos en el XMP de `META-INF/metadata.xml`, que no pasa por structs tipados. Si algún día un tipo tipado recibe un atributo con prefijo, hay que sustituir la etiqueta por `xmlutil.UnmarshalAttrs`/`MarshalAttrs`, que sí lo conservan.

Los tipos con `UnmarshalXML` propio no reciben la etiqueta, porque `encoding/xml` no la aplica a un tipo con lectura propia: ahí se usan esas dos funciones de `internal/xmlutil/attrs.go`.

### 2. `omitempty` con atributos de valor vacío — RESUELTO

Defecto distinto y menos evidente, que apareció tras poner los comodines. InDesign emite atributos con **valor vacío**, como `OverriddenPageItemProps=""`. Con un campo `string` y `omitempty`, Go no distingue «ausente» de «presente y vacío», así que los omitía al escribir.

Se resolvió quitando `omitempty` de los pares tipo/atributo afectados, **solo tras medir** que el atributo aparece con valor vacío en el corpus y que **nunca falta** en un elemento de su tipo. Se acotó por struct y no globalmente: `Name`, por ejemplo, es un nombre de atributo común y en otros tipos sí puede omitirse.

### 3. Instrucciones de proceso dentro de `<Content>` — RESUELTO

El texto de una story puede llevar marcadores de carácter especial de InDesign, como `<?ACE 18?>`, intercalados con el texto. `Content` se modelaba solo como datos de carácter, así que `<Content><?ACE 18?>.</Content>` se re-emitía como `<Content>.</Content>`: **un carácter perdido del documento**.

`Content` guarda ahora dos formas: `Text` con `,chardata`, que es el texto legible y la fuente de verdad, y `Raw` con `,innerxml`, el contenido literal. Al escribir reemite el literal **solo si `Text` no ha cambiado** desde que se parseó; si cambió, gana `Text`. Sin esa comprobación, un programa de reemplazo de texto vería su cambio descartado en silencio, que es peor que perder el marcador.

### 4. Hijos reagrupados por tipo — PARCIAL, es lo que queda

Un struct de Go declara sus hijos en campos por tipo, y al serializar los emite en el orden de esos campos. Pero en el XML de entrada venían intercalados. El resultado tiene los mismos elementos en otro orden.

La solución es `internal/xmlutil/childorder.go`: un registro de la secuencia leída que se reproduce al emitir. **No es un contenedor**: el contenido sigue viviendo en su campo por tipo, que es su fuente de verdad; el registro solo guarda el orden. Eso permitió arreglarlo sin tocar los 158 accesos a los campos de `Document`.

`ChildOrder.Replay` emite en dos pasadas: primero el orden registrado, y después los hijos que el registro no menciona. La segunda pasada es lo que hace el mecanismo seguro: sin ella, agregar un hijo sin actualizar el registro lo haría **desaparecer** al serializar, y perder un elemento es peor que emitirlo en una posición rara. Tiene un efecto secundario útil: un modelo construido desde cero, sin registro, cae entero en la segunda pasada y sale en el orden de los campos, que es el comportamiento anterior. No hay dos caminos que mantener.

**Aplicado a:** `designmap.xml` (el elemento `Document`), `Resources/Styles.xml`, `Resources/Graphic.xml` y, desde la Tarea 7, los **spreads** (el elemento `Spread`).

En los spreads el registro va acompañado de `SpreadElement.Items`, la secuencia de elementos de página en orden documental con punteros a los elementos guardados en los campos por tipo. El reparto de responsabilidades es deliberado y viene de un conflicto medido: **`Items` manda el orden y los campos por tipo mandan el contenido**, porque hay código que borra elementos escribiendo directamente en esos campos —`removeItemFromSpread` de `pkg/idml`— y si el contenido saliera de `Items` ese borrado no llegaría al archivo escrito, sin que ningún test lo delatara. Las fases 2 y 3 migran esas escrituras a `Append` y `Remove`, y entonces `Items` pasa a ser también el contenido.

Para decidir la forma se midieron los **54** `Spread` y `MasterSpread` de los cinco documentos: en **0** casos aparece un hijo que no es elemento de página después de un elemento de página. La forma es siempre `[FlattenerPreference o Properties][Page…][elementos de página…]`, así que el prefijo no necesita contenedor.

Desde la Tarea 2e se aplica también a los dos tipos compartidos que se descubrieron al ampliar el corpus: **`common.Properties`**, usado en 32 sitios, y **`ObjectStyleGroup`**.

Antes de escribirlo se evaluó un atajo mucho más barato: **reordenar los campos del struct** poniendo el comodín primero, dos líneas en lugar de dos serializadores. La medición lo descartó, y merece la pena saberlo porque es tentador. Sobre los 2224 elementos `<Properties>` del corpus, 10 traen un hijo tipado **después** de los del comodín y **5 lo traen antes**, con la forma `Label, AppliedMathMLSwatch`. Poner el comodín primero arreglaría 10 y **rompería 5**. En `ObjectStyleGroup` pasa lo mismo: 2 casos con el grupo anidado delante y 2 con el estilo delante. **Ninguna disposición fija de los campos sirve**, hace falta recordar el orden de cada elemento. Hay un test por disposición para que el atajo no se reintente.

Ese cambio obligó a mover el registro de orden. `pkg/common` lo necesitaba, pero `internal/xmlutil` importa `pkg/common` para sus helpers de error, así que había un ciclo. `childorder.go` —que no importa nada fuera de la biblioteca estándar— vive ahora en **`internal/xmlorder`**, y `xmlutil` conserva dos alias de tipo para que las 24 referencias existentes no cambiaran.

**Pendiente en:** las stories, en sus dos niveles, `Story` y `ParagraphStyleRange`.

Los 71 desórdenes que quedan, por elemento padre. Medido con `IDMLLIB_MAX_DIFFS=0`; la receta para reproducir esta tabla está más arriba, en «Recetas de medición»:

| Elemento padre | Desórdenes | Qué se reordena | Quién lo cierra |
|---|---|---|---|
| `root/Story` | 30 | `MetadataPacketPreference` se emite al final | Tarea 10, criterio 4 lo dice con esta cifra |
| `root/Story/ParagraphStyleRange` | 23 | `Change` entre dos `CharacterStyleRange` se va al final | **hueco**, ver abajo |
| `root/Spread/GraphicLine` | 1 | `ObjectExportOption` y `TextWrapPreference` se intercambian | **hueco**, ver abajo |
| ~~`root/Spread`~~ | ~~5~~ → **0** | los elementos de página, reagrupados por tipo | **cerrado por la Tarea 7** |
| ~~`root/Spread/Page/Properties`~~ | ~~8~~ → **0** | `Label` se adelantaba a lo del comodín | **cerrado por la Tarea 2e** |
| ~~`root/Section/Properties`~~ | ~~2~~ → **0** | igual que `Page/Properties` | **cerrado por la Tarea 2e** |
| ~~`root/RootObjectStyleGroup`~~ | ~~2~~ → **0** | los estilos se adelantaban a los grupos anidados | **cerrado por la Tarea 2e** |

Dos lecturas de esta tabla.

**53 de las 54 que quedan están en las stories.** La Tarea 10 y el hueco de `ParagraphStyleRange` son prácticamente todo lo que falta. Los spreads los cerró la Tarea 7 y los tipos compartidos la 2e.

**La fila que queda fuera de las stories no tiene dueño.** El desorden de `GraphicLine` es el de los hijos de un **elemento de página**: las Tareas 13 y 14 nombran `ObjectExportOption` y `TextWrapPreference` pero para **añadirlos cuando faltan**, no para ordenarlos, y aquí ya están los dos, solo intercambiados. Está anotado en la Tarea 10. El candidato natural es un registro de orden en `PageItemBase`, que es donde se declaran esos campos.

**Dos filas no las cubre ninguna tarea del plan.** Los 23 de `ParagraphStyleRange` son el mismo patrón y el mismo archivo que la Tarea 10, pero un nivel más abajo: la Tarea 10 da contenedor ordenado a `StoryElement`, y esto lo necesita `ParagraphStyleRange` para intercalar `Change` entre los `CharacterStyleRange`. Está anotado como criterio adicional de la Tarea 10 en lugar de tarea nueva, porque es el mismo archivo y el mismo patrón. El de `GraphicLine` es el orden de los hijos de un elemento de página, que tampoco tiene dueño: las Tareas 13 y 14 hablan de `ObjectExportOption` y `TextWrapPreference` pero para **añadirlos cuando faltan**, no para ordenarlos, y aquí ya están los dos, solo intercambiados.

**Las cifras por origen no se reparten igual** entre estas filas (35 en `documento_referencia`, 26 en `tripple`, 10 en `example`), porque cada documento tiene una mezcla distinta. Al cerrar una fila, comprobar el total, no el de un origen.

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

## Dos tests dormidos por el orden documental

`pkg/idml/roundtrip_test.go` tiene dos tests desactivados **desde antes de este trabajo**, y sus mensajes de omisión decían «element order changes due to Go's xml.Marshal»:

- `TestRoundtrip_ByteComparison`
- `TestRoundtripStructure_StructuralComparison`

No están dormidos por diseño: están dormidos **por el único defecto que queda**. Y el diagnóstico que llevaban era incorrecto: no es una limitación de `xml.Marshal`, es que el modelo emitía los hijos agrupados por el orden de los campos del struct.

**Cuándo se levantan:** cuando `orden-elementos-distinto` llegue a 0. Entonces hay que **intentar** reactivarlos, porque comparan a nivel de contenido de archivo y pueden delatar algo que el arnés no ve. Si tras el orden siguen fallando, lo que reporten es información nueva y hay que investigarla, no volver a desactivarlos.

Dos datos medidos que acotan lo que cabe esperar al reactivarlos. El primero: sobre `Spread_uce7.xml`, el archivo más grande del corpus, la salida ya sale a **286.830 bytes contra 286.839 de entrada**, nueve de diferencia. Con la comparación byte a byte no hay margen para que sobrevivan defectos grandes escondidos. El segundo: **el bloque `<![CDATA[ ]]>` del paquete XMP se conserva intacto y sin escapar**, que era el riesgo obvio de una comparación byte a byte y resultó no serlo.

Es la lección más útil de todo este trabajo: **un `t.Skip` sin condición de reactivación es deuda que se disfraza de suite en verde.** Estos dos ocultaron el defecto del orden durante todo el desarrollo previo, y el arnés existe precisamente porque nadie tenía una cifra que mirar.

## Lo que sigue sin estar verificado

- El **orden documental** de 71 nodos: los elementos de página de los spreads, los hijos de las stories, y los tipos `Properties` y `ObjectStyleGroup`.
- Los archivos bajo `MasterSpreads/` **no se parsean**: se copian tal cual. Son los 11 «copiados sin parsear» del resumen.
- Solo se ha comprobado que InDesign abre un documento **generado desde cero**. No se ha comprobado que abra un documento **existente reescrito** por la librería.
- El corpus son cinco documentos de una misma versión de InDesign. No hay especificación oficial vigente contra la que validar: Adobe retiró la de IDML de su sitio.
- El orden de los hijos de los `ObjectStyle` se reordena y el arnés **no lo ve**, porque `ObjectStyle` está en la lista `SortElements` del comparador. Queda abierto si InDesign comparte ese criterio.
