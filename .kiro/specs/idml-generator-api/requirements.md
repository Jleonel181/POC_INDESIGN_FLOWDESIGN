# Documento de Requisitos

## Introducción

`idmllib` es hoy un *round-tripper*: abre un paquete IDML existente, parsea un subconjunto de sus archivos, permite modificarlos y los reescribe copiando byte a byte todo lo que no entendió. Esta funcionalidad convierte la librería en un **generador**: producir un archivo IDML válido y abrible por Adobe InDesign a partir de una entrada JSON que un binario de línea de comandos lee de la entrada estándar, sin partir de ningún paquete previo.

El alcance funcional se define por un archivo de referencia real exportado por InDesign, descomprimido en `idmllib/testdata/documento_referencia/` (página de periódico: 1 spread, 3 master spreads, 30 stories, texto multicolumna, imágenes enlazadas, grupos anidados, líneas gráficas, capas, secciones, guías, control de cambios). Todo lo que ese archivo modela debe poder generarse; todo lo que ese archivo contiene debe poder parsearse y re-serializarse sin pérdida estructural.

La estrategia elegida es **constructor completo desde cero**: el paquete se ensambla íntegro desde structs de Go, sin plantilla-esqueleto que se parchee.

Dos deudas estructurales del modelo actual son prerequisito de todo lo demás y por eso aparecen como requisitos 2 y 3: el modelo pierde el **orden documental** (que en IDML es el orden de apilamiento) y descarta los **atributos XML no declarados** en el struct. Construir el generador antes de resolverlas produciría documentos incompletos.

## Glosario

- **IDML**: formato de intercambio de Adobe InDesign; paquete ZIP de archivos XML con `mimetype` como primera entrada sin comprimir.
- **IDMS**: formato de *snippet* de InDesign; un único archivo XML. Ya soportado por `pkg/idms`.
- **Paquete_IDML**: el tipo `idml.Package` (`pkg/idml/package.go`), coordinador de archivos, caché de objetos parseados y orden de entradas del ZIP.
- **Modelo_Spread**: los tipos de `pkg/spread` (`Spread`, `SpreadElement`, `Page`, `Rectangle`, `SpreadTextFrame`, `Oval`, `Polygon`, `GraphicLine`, `Group`, `Image`, `PDF`, `Guide`).
- **Modelo_Story**: los tipos de `pkg/story` (`Story`, `StoryElement`, `ParagraphStyleRange`, `CharacterStyleRange`).
- **Serializador_XML**: el conjunto de funciones `Marshal*`/`Unmarshal*` de los paquetes de dominio más los helpers de `internal/xmlutil`.
- **Arnes_Fidelidad**: la prueba automatizada que parsea, re-serializa y compara estructuralmente los archivos XML del archivo de referencia.
- **Comparación_Estructural**: comparación de dos documentos XML por árbol (elementos, atributos y texto), insensible al orden de atributos y al espacio en blanco no significativo, implementada en `internal/xmlutil/compare.go` e `internal/testutil/comparison.go`.
- **Orden_Documental**: el orden de aparición de los elementos hijo dentro de un elemento XML. En un spread IDML determina el orden de apilamiento (z-order); en una story determina la secuencia del texto.
- **OtherAttrs**: campo `[]xml.Attr` que almacena los atributos presentes en el XML de entrada que no corresponden a ningún campo declarado del struct.
- **Ensamblador_Paquete**: la API pública de `pkg/idml` que agrega archivos al Paquete_IDML y mantiene coherentes el `designmap.xml`, el `StoryList` y el orden de entradas del ZIP.
- **Generador_IDs**: el componente que produce identificadores `Self` únicos con el formato que emite InDesign (por ejemplo `u1a2b`, `Rectangle/u3c4d`).
- **Constructor_Documento**: la API de alto nivel que construye un documento completo desde structs de opciones de Go (`NewDocument` → `AddMasterSpread` → `AddPage` → `AddTextFrame`/`AddImageFrame`/`AddShape`/`AddGroup` → `Build`).
- **Resolvedor_Imagenes**: el componente que dada una ruta local o una carga base64 obtiene los bytes de la imagen, sus dimensiones en píxeles y su resolución.
- **CLI_Generador**: el binario `cmd/idmlgen` que expone el Constructor_Documento sobre la entrada y la salida estándar.
- **Validador_Entrada**: el componente que valida el JSON recibido antes de invocar al Constructor_Documento.
- **Documento_Referencia**: el IDML exportado por InDesign descomprimido en `idmllib/testdata/documento_referencia/` (43 archivos `.xml` más `mimetype`). Está **versionado en el repositorio**, de modo que la medición de fidelidad es reproducible en cualquier clon. Respecto de las preferencias cumple además el papel de **oráculo de diagnóstico**: su `Resources/Preferences.xml` sirve para comparar e identificar un atributo ausente cuando InDesign rechaza el archivo generado, nunca como plantilla que se copie.
- **Subconjunto_Curado**: el conjunto de elementos que el generador emite en `Resources/Preferences.xml`: los 10 elementos de primer nivel de la plantilla `pkg/idml/templates/minimal/Preferences.xml` más `DocumentPreference`, `MarginPreference` y `ViewPreference`, construidos con los valores que recibe el Constructor_Documento.
- **Archivo_Evidencia_Imagenes**: el IDML `idmllib/testdata/archivo_evidencia_imagenes.idml` exportado por InDesign, cuyo `Spreads/Spread_ud1.xml` contiene dos imágenes embebidas (`Self="ue8"`, PNG de 8.339 bytes; `Self="uee"`, JPEG de 60.516 bytes) y una imagen enlazada de control (`Self="uf9"`, JPEG de 21.412 bytes). Es la única fuente de verdad del formato de imagen embebida. Igual que el Documento_Referencia, está **versionado en el repositorio**, en su forma de paquete `.idml` comprimido. **Regenerable solo en la máquina del usuario**: el documento fuente `imagenes_idml.indd` se conserva en la raíz del workspace pero queda fuera de git, así que producir una variante exige exportar de nuevo desde InDesign en esa máquina y no es posible desde un clon del repositorio.

---

## Requisitos

### Requisito 1: Medición de fidelidad estructural

**User Story:** Como mantenedor de la librería, quiero una prueba que mida cuánto del archivo de referencia sobrevive un ciclo parseo → serialización, para saber con datos qué falta implementar y detectar regresiones.

#### Acceptance Criteria

1. THE Arnes_Fidelidad SHALL recorrer recursivamente el directorio del Documento_Referencia, seleccionar los 43 archivos cuyo nombre termina en `.xml`, excluir el archivo `mimetype`, y por cada archivo seleccionado parsearlo, re-serializarlo con el Serializador_XML y verificarlo contra el archivo original mediante Comparación_Estructural.
2. WHEN una Comparación_Estructural detecta una diferencia, THE Arnes_Fidelidad SHALL reportar la ruta del archivo, la ruta del nodo afectado, la categoría de la diferencia entre las categorías: atributo ausente, atributo con valor distinto, elemento ausente, elemento sobrante, orden de elementos distinto y contenido de texto distinto, el valor esperado y el valor obtenido de esa diferencia, hasta un máximo de 100 diferencias por archivo, marcando el reporte de ese archivo como truncado al alcanzar ese máximo.
3. THE Arnes_Fidelidad SHALL emitir el número de archivos sin diferencias y el número de archivos con diferencias sobre el total de archivos XML del Documento_Referencia.
4. THE Arnes_Fidelidad SHALL usar `internal/xmlutil/compare.go` e `internal/testutil/comparison.go` como único mecanismo de Comparación_Estructural.
5. IF un archivo del Documento_Referencia falla al parsear o falla al re-serializar, THEN THE Arnes_Fidelidad SHALL registrar ese archivo como fallo indicando la etapa en que ocurrió el fallo entre parseo y re-serialización, indicando el mensaje de error, y SHALL continuar con los archivos restantes.
6. IF al menos un archivo presenta diferencias de Comparación_Estructural o queda registrado como fallo de parseo o de re-serialización, THEN THE Arnes_Fidelidad SHALL terminar con resultado de fallo.
7. IF el directorio del Documento_Referencia está presente en la ruta esperada y su recorrido selecciona cero archivos `.xml`, THEN THE Arnes_Fidelidad SHALL terminar con resultado de fallo indicando la ruta recorrida.
8. WHERE un archivo seleccionado corresponde a un tipo de archivo que el Paquete_IDML copia sin parsear, THE Arnes_Fidelidad SHALL contarlo en la categoría «copiados sin parsear», separada del número de archivos sin diferencias, y SHALL mantener el resultado del arnés sin fallo por causa de ese archivo.
9. THE Arnes_Fidelidad SHALL incluir en su corpus, además del Documento_Referencia, el Archivo_Evidencia_Imagenes, aplicándole el mismo ciclo de parseo, re-serialización y Comparación_Estructural, de modo que el ciclo de la imagen embebida quede cubierto por la misma medición.
10. IF el directorio del Documento_Referencia o el Archivo_Evidencia_Imagenes no está presente en la ruta esperada, condición anómala en una copia de trabajo íntegra porque ambos elementos están versionados, THEN THE Arnes_Fidelidad SHALL omitir su ejecución registrando el motivo y la ruta esperada, y SHALL NOT reportar un fallo, siguiendo el patrón `t.Skipf` que el repositorio ya usa en `pkg/idml/fonts_test.go` y `pkg/idml/xml_test.go`, de modo que una copia de trabajo con el dato borrado o con la ruta mal configurada produzca un diagnóstico en lugar de un fallo opaco.
11. THE Arnes_Fidelidad SHALL localizar cada elemento del corpus en una ruta documentada, `idmllib/testdata/documento_referencia/` y `idmllib/testdata/archivo_evidencia_imagenes.idml`, configurable mediante variable de entorno o constante del paquete de prueba, y SHALL indicar en el mensaje de omisión del criterio 10 cuál era la ruta esperada, de modo que la ausencia sea diagnosticable.

---

### Requisito 2: Preservación de atributos XML no modelados

**User Story:** Como desarrollador que usa la librería, quiero que los atributos XML que la librería todavía no modela sobrevivan el ciclo de lectura y escritura, para que abrir y guardar un documento no borre información de InDesign.

#### Acceptance Criteria

1. THE Serializador_XML SHALL proveer en `internal/xmlutil` una función única que, dado un elemento XML y un puntero a struct, asigne los atributos declarados a sus campos y almacene los atributos restantes en el campo OtherAttrs del struct, conservando de cada atributo restante su prefijo de namespace, su nombre local, su valor literal y su orden de aparición, para al menos 128 atributos por elemento.
2. WHEN el Serializador_XML decodifica un elemento en su struct destino, THE Serializador_XML SHALL ejecutar el método `UnmarshalXML` de ese struct exactamente una vez por elemento, decodificando el struct mediante un tipo alias para que ese método no se invoque a sí mismo.
3. THE Serializador_XML SHALL proveer la función recíproca que, al serializar, emita los atributos de los campos declarados seguidos de los atributos almacenados en OtherAttrs, emita cada nombre de atributo una sola vez cuando un atributo de OtherAttrs coincida con el nombre de un campo declarado, y emita el elemento sin atributos adicionales cuando OtherAttrs está vacío.
4. THE Modelo_Spread SHALL exponer OtherAttrs en los tipos `Rectangle`, `SpreadTextFrame`, `Oval`, `Polygon`, `GraphicLine`, `Group`, `Image`, `PDF`, `Page`, `Guide` y `SpreadElement`.
5. THE Modelo_Story SHALL exponer OtherAttrs en los tipos `ParagraphStyleRange` y `StoryElement`.
6. WHEN el Arnes_Fidelidad procesa el Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias de categoría «atributo ausente» y cero diferencias de categoría «atributo con valor distinto».
7. THE Modelo_Spread SHALL preservar en `Rectangle` los atributos `FillColor`, `FillTint`, `StrokeWeight`, `StrokeColor`, `TopLeftCornerOption`, `TopLeftCornerRadius`, `CornerOption`, `CornerRadius`, `FlexItemWidthMode` y `BeforeGroupingLayerPosition` presentes en el Documento_Referencia.
8. WHEN un elemento de entrada se almacena en un campo comodín `OtherElements`, THE Serializador_XML SHALL emitirlo conservando sus propios atributos y sus elementos hijo con los atributos de cada hijo.
9. IF el valor de un atributo de entrada no es convertible al tipo del campo declarado que le corresponde, THEN THE Serializador_XML SHALL devolver un error que identifique el nombre del atributo, el valor recibido y el tipo esperado, y SHALL dejar el struct destino sin asignar ese campo.
10. WHEN un elemento se somete a dos ciclos consecutivos de parseo y serialización, THE Serializador_XML SHALL producir en el segundo ciclo el mismo resultado que en el primero, con el mismo conjunto de atributos en OtherAttrs y en el mismo orden.

---

### Requisito 3: Preservación del orden documental

**User Story:** Como usuario del generador, quiero que el orden de los elementos se conserve, porque en un spread ese orden es el orden de apilamiento visual y en una story es la secuencia del texto.

#### Acceptance Criteria

1. THE Modelo_Spread SHALL almacenar los elementos de página de un `SpreadElement` en un contenedor único que conserve el Orden_Documental de entrada, en lugar de un slice independiente por tipo de elemento.
2. WHEN se parsea el `SpreadElement` de `Spreads/Spread_uce7.xml`, cuyos **32** elementos de página hijos directos de `<Spread>` forman la secuencia `Rectangle`, `Rectangle`, `Group`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `Rectangle`, `Rectangle`, `Rectangle`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `Rectangle`, `Rectangle`, `Rectangle`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `Rectangle`, `TextFrame`, `GraphicLine`, `Group` —20 `TextFrame`, 9 `Rectangle`, 2 `Group` y 1 `GraphicLine`—, THE Modelo_Spread SHALL almacenar esos elementos en esa misma secuencia y SHALL conservar el número total de elementos de página de la entrada, que es 32.
3. WHEN se serializa un `SpreadElement`, THE Serializador_XML SHALL emitir los elementos de página en el mismo Orden_Documental en que fueron almacenados, sin reagruparlos por tipo de elemento.
4. THE Modelo_Spread SHALL exponer accesores por tipo (marcos de texto, rectángulos, imágenes, óvalos, polígonos, líneas gráficas, grupos) que devuelvan en Orden_Documental **punteros** a los elementos de ese tipo contenidos en el nivel inmediato del `SpreadElement`, sin descender al contenido de los grupos, y que devuelvan una colección vacía de longitud 0 cuando ese nivel no contiene ningún elemento de ese tipo. THE Modelo_Spread SHALL NOT devolver copias de los elementos almacenados.
5. THE Modelo_Story SHALL almacenar los hijos de `StoryElement`, incluidos `ParagraphStyleRange`, los elementos de control de cambios y los elementos que el Modelo_Story no modela, en un contenedor único que conserve el Orden_Documental y la posición de cada hijo respecto de los demás, siguiendo el mismo patrón ya implementado en `CharacterStyleRange` (`Children []CharacterChild` con `MarshalXML`/`UnmarshalXML` propios en `pkg/story/parse.go`).
6. WHEN el contenedor ordenado reemplaza los slices por tipo, THE Modelo_Spread SHALL mantener compilables los seis paquetes que hoy acceden a esos slices —`pkg/spread` (`spread.go`, `doc.go`), `pkg/idml` (`selection.go`, `package_modifications.go`, `index.go`, `resourcemgr_cleanup.go`, `resourcemgr_validation.go`), `pkg/idms` (`exporter_build.go`, `exporter.go`), `pkg/analysis` (`tracker.go`), `cmd/debug-export` y `cmd/cli` (`tui/export_idms.go`, `tui/textframe_selector.go`)—, incluidos los dos binarios existentes `cmd/cli` y `cmd/debug-export`, y SHALL mantener en verde las pruebas existentes de esos paquetes. WHERE una aserción existente accede al slice por tipo como campo, THE Modelo_Spread SHALL admitir su adaptación mecánica a llamada de accesor, y SHALL NOT admitir que ninguna aserción se debilite ni se elimine; THE Modelo_Spread SHALL NOT reducir el número de pruebas de cada paquete afectado por debajo de su valor actual: 175 en `pkg/idml`, 47 en `pkg/idms` y 25 en `pkg/analysis`. THE Modelo_Spread SHALL NOT migrar ni modificar los seis campos homónimos del struct `Selection` de `pkg/idml/selection.go` líneas 38 a 53, que declaran los mismos nombres (`TextFrames`, `Rectangles`, `Ovals`, `Polygons`, `GraphicLines`, `Groups`) y no pertenecen al Modelo_Spread.
7. WHEN el Arnes_Fidelidad procesa el spread, los 3 master spreads y las 30 stories del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias de categoría «orden de elementos distinto».
8. IF un elemento hijo de un `SpreadElement` o de un `StoryElement` no corresponde a ningún tipo modelado, THEN THE Serializador_XML SHALL almacenarlo como XML crudo dentro del mismo contenedor ordenado del elemento contenedor, conservando su posición respecto de los elementos modelados.
9. WHEN se agrega un elemento de página a un `SpreadElement`, THE Modelo_Spread SHALL colocarlo en la última posición del contenedor ordenado.
10. WHEN se elimina un elemento de página de un `SpreadElement`, THE Modelo_Spread SHALL conservar el Orden_Documental relativo de los elementos restantes y SHALL reducir en 1 el número total de elementos de página.
11. WHEN el Modelo_Spread almacena los elementos de página de `Spreads/Spread_uce7.xml` y los de `MasterSpreads/MasterSpread_u2dd.xml`, THE Modelo_Spread SHALL conservar la secuencia de modo que no ocurra la reagrupación por tipo, que desplaza 13 de los 32 elementos del spread y 12 de los 17 de ese master spread.
12. THE Modelo_Spread SHALL almacenar `<Image>` como hijo del marco que la contiene y NOT como hermano en el contenedor ordenado de nivel de spread, porque las 7 imágenes del Documento_Referencia tienen `<Rectangle>` como padre inmediato.
13. THE Modelo_Spread SHALL almacenar en el contenedor ordenado punteros a los elementos de página y NOT valores, y SHALL declarar la interfaz de ese contenedor con métodos de receptor puntero, de modo que un valor no la satisfaga y el compilador rechace almacenar una copia. THE Modelo_Spread SHALL declarar esa interfaz en `pkg/spread` o en `pkg/common` y SHALL NOT reutilizar la declarada en `pkg/idml/interfaces.go` línea 61, porque `pkg/spread` no importa `pkg/idml` ni puede hacerlo sin crear un ciclo de importación.
14. WHEN un consumidor obtiene un elemento de página mediante un accesor por tipo, muta uno de sus atributos y el `SpreadElement` se re-serializa, THE Serializador_XML SHALL emitir ese atributo con el valor mutado, de modo que la verificación falle si el accesor devuelve copias en lugar de punteros.

---

### Requisito 4: Grupos anidados

**User Story:** Como usuario del generador, quiero agrupar elementos y anidar grupos dentro de grupos, para reproducir composiciones como las del archivo de referencia.

#### Acceptance Criteria

1. THE Modelo_Spread SHALL almacenar el contenido de página de `Group` en el mismo contenedor ordenado que el Requisito 3 define para `SpreadElement`, admitiendo los tipos `Rectangle`, `SpreadTextFrame`, `Oval`, `Polygon`, `GraphicLine`, `Image`, `PDF` y `Group`, en lugar del comodín de elementos no modelados que usa hoy.
2. WHEN se serializa un `Group`, THE Serializador_XML SHALL emitir su contenido de página en el mismo Orden_Documental en que fue parseado o agregado, y SHALL aplicar la misma regla al contenido de cada grupo anidado en cualquier nivel.
3. THE Modelo_Spread SHALL admitir grupos anidados hasta 32 niveles de profundidad, contando el grupo de nivel superior como nivel 1.
4. WHEN el Paquete_IDML recorre los elementos de página de un `Spread` o de un `MasterSpread` para seleccionar, validar o indexar, THE Paquete_IDML SHALL visitar cada elemento contenido en un `Group` y en sus grupos anidados exactamente una vez, en recorrido en profundidad, visitando cada `Group` antes que su contenido y respetando el Orden_Documental en cada nivel.
5. WHEN el Arnes_Fidelidad procesa el spread del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias de Comparación_Estructural en los elementos `Group` y en su contenido, conservando por cada grupo su nivel de anidamiento, el número de elementos hijo y el tipo de cada elemento hijo.
6. IF la entrada contiene grupos anidados con una profundidad mayor a 32 niveles, THEN THE Modelo_Spread SHALL devolver un error que identifique el atributo `Self` del grupo más profundo y la profundidad alcanzada, sin devolver un modelo parcial del spread.
7. IF un `Group` no contiene ningún elemento de página, THEN THE Serializador_XML SHALL emitirlo conservando sus atributos y sin agregar elementos hijo que no estuvieran en la entrada.

---

### Requisito 5: Master spreads como tipo de primera clase

**User Story:** Como usuario del generador, quiero definir páginas maestras con su contenido, para que las páginas del documento hereden márgenes, guías y elementos base.

#### Acceptance Criteria

1. WHEN se parsea un archivo cuyo elemento raíz es `idPkg:MasterSpread` con elemento interno `MasterSpread`, THE Modelo_Spread SHALL producir el modelo del master spread y SHALL almacenar el valor del atributo `DOMVersion` del elemento raíz, aplicando el mismo tratamiento a los archivos cuyo elemento raíz es `idPkg:Spread` con elemento interno `Spread`.
2. THE Modelo_Spread SHALL declarar en `MasterSpread` los atributos `Self`, `Name`, `NamePrefix`, `BaseName`, `ShowMasterItems`, `PageCount`, `OverriddenPageItemProps`, `PrimaryTextFrame` e `ItemTransform`, SHALL preservar en OtherAttrs según el Requisito 2 los demás atributos presentes en la entrada, y SHALL emitir el valor de cada atributo con el mismo texto literal recibido, sin normalizarlo (por ejemplo `PrimaryTextFrame="n"` se emite como `n`).
3. THE Modelo_Spread SHALL almacenar el contenido de página de `MasterSpread` en el mismo contenedor ordenado que el Requisito 3 define para `SpreadElement`, admitiendo los mismos tipos de contenido y exponiendo los mismos accesores por tipo, y SHALL conservar el número de elementos de página hijos directos de `<MasterSpread>` medido en cada uno de los 3 master spreads del Documento_Referencia: **17** en `MasterSpreads/MasterSpread_u2dd.xml` (8 `GraphicLine`, 5 `TextFrame`, 4 `Rectangle`), **8** en `MasterSpreads/MasterSpread_u409.xml` (4 `GraphicLine`, 4 `TextFrame`) y **0** en `MasterSpreads/MasterSpread_ub0.xml`, que tiene páginas pero ningún contenido de página.
4. THE Paquete_IDML SHALL clasificar como master spread las rutas cuyo prefijo es `MasterSpreads/` y cuyo sufijo es `.xml`, SHALL clasificar como spread las rutas cuyo prefijo es `Spreads/` y cuyo sufijo es `.xml`, SHALL asignar cada ruta a lo sumo una de esas dos clasificaciones, y SHALL clasificar las rutas restantes como ninguna de las dos.
5. WHEN el Arnes_Fidelidad procesa los 3 archivos de `MasterSpreads/` del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias en las categorías definidas en el Requisito 1, criterio 2.
6. WHEN se serializa un `MasterSpread`, THE Serializador_XML SHALL emitir el elemento raíz con el nombre `idPkg:MasterSpread`, el elemento interno con el nombre `MasterSpread`, y el atributo `DOMVersion` del elemento raíz con el valor leído en el parseo.
7. IF el nombre del elemento raíz o el nombre del elemento interno de un archivo de `Spreads/` o de `MasterSpreads/` no coincide con los nombres reconocidos, THEN THE Modelo_Spread SHALL devolver un error que identifique la ruta del archivo, el nombre encontrado y el nombre esperado, sin devolver un modelo parcial.

---

### Requisito 6: Elementos de marco, forma y apariencia faltantes

**User Story:** Como usuario del generador, quiero definir marcos de texto multicolumna, transparencias y metadatos por elemento, porque el archivo de referencia los usa y sin ellos el resultado no se parece al original.

#### Acceptance Criteria

1. THE Modelo_Spread SHALL declarar `TextFramePreference` como hijo de `SpreadTextFrame` con los 19 atributos que el Documento_Referencia declara en ese elemento: `TextColumnCount`, `TextColumnGutter`, `TextColumnFixedWidth`, `TextColumnMaxWidth`, `VerticalJustification`, `ColumnRuleOverride`, `ColumnRuleOffset`, `ColumnRuleTopInset`, `ColumnRuleBottomInset`, `ColumnRuleInsetChainOverride`, `ColumnRuleStrokeWidth`, `ColumnRuleStrokeColor`, `ColumnRuleStrokeType`, `ColumnRuleStrokeTint`, `ColumnRuleOverprintOverride`, `FootnotesEnableOverrides`, `FootnotesSpanAcrossColumns`, `FootnotesMinimumSpacing` y `FootnotesSpaceBetween`.
2. THE Serializador_XML SHALL declarar el tipo `TextFramePreference` exactamente una vez en el repositorio, en `pkg/resources/styles.go` o en un paquete compartido, y SHALL hacerlo alcanzable desde `pkg/spread` y desde `pkg/resources` sin declarar un segundo tipo con ese nombre.
3. THE Modelo_Spread SHALL declarar `TransparencySetting` como hijo de los elementos de página con los tres hijos que el Documento_Referencia declara: `BlendingSetting` con el atributo `Opacity`, `DirectionalFeatherSetting` con los atributos `Applied`, `TopWidth` y `BottomWidth`, y `DropShadowSetting` con los atributos `Mode`, `BlendMode`, `Opacity`, `XOffset`, `YOffset`, `Size`, `Spread` y `EffectColor`.
4. THE Modelo_Spread SHALL declarar `MetadataPacketPreference` como hijo de los elementos de página y THE Modelo_Story SHALL declararlo como hijo de las stories, conservando el texto de su hijo `Properties/Contents` con el mismo contenido literal recibido y emitiéndolo dentro de una sección CDATA.
5. THE Modelo_Spread SHALL sustituir la definición marcadora actual de `ObjectExportOption` en `pkg/spread/page_items.go` por una definición que declare los 30 atributos que el Documento_Referencia declara en ese elemento: `AltTextSourceType`, `ActualTextSourceType`, `CustomAltText`, `CustomActualText`, `ApplyTagType`, `ImageConversionType`, `ImageExportResolution`, `GIFOptionsPalette`, `GIFOptionsInterlaced`, `JPEGOptionsQuality`, `JPEGOptionsFormat`, `ImageAlignment`, `ImageSpaceBefore`, `ImageSpaceAfter`, `UseImagePageBreak`, `ImagePageBreak`, `CustomImageAlignment`, `SpaceUnit`, `CustomLayout`, `CustomLayoutType`, `EpubType`, `SizeType`, `CustomSize`, `PreserveAppearanceFromLayout`, `EpubAriaRole`, `EpubAriaLabel`, `EpubAriaLabelSourceType`, `AltTextGenerationError`, `AltTextCropSyncRect` y `AIGeneratedAltText`, más los hijos `Properties/AltMetadataProperty` y `Properties/ActualMetadataProperty`, cada uno con los atributos `NamespacePrefix` y `PropertyPath`.
6. THE Modelo_Spread SHALL declarar en `Rectangle` los seis atributos de relleno y trazo que ya declara `Oval` en `pkg/spread/spread.go`: `FillColor`, `FillTint`, `StrokeWeight`, `StrokeType`, `StrokeColor` y `StrokeTint`.
7. THE Modelo_Spread SHALL declarar el hijo `Properties/InsetSpacing` de `TextFramePreference` con el atributo `type` de valor `list` y con 4 elementos hijo `ListItem`, cada uno con el atributo `type` de valor `unit`, conservando el Orden_Documental de los 4 valores.
8. IF el Constructor_Documento recibe un valor de `TextColumnCount` menor que 1 o un valor de `TextColumnGutter` menor que 0, THEN THE Constructor_Documento SHALL devolver un error que identifique el nombre del atributo y el valor recibido, sin generar archivo.
9. WHEN un elemento de página de entrada no declara `TextFramePreference`, `TransparencySetting`, `MetadataPacketPreference` ni `ObjectExportOption`, THE Serializador_XML SHALL emitir ese elemento sin ninguno de esos cuatro elementos hijo.

---

### Requisito 7: Formato de párrafo completo

**User Story:** Como usuario del generador, quiero controlar la composición de los párrafos, para que el texto generado se maquete como el del documento original.

#### Acceptance Criteria

1. THE Modelo_Story SHALL declarar en `ParagraphStyleRange` los 11 atributos que el Documento_Referencia declara en ese elemento: `AppliedParagraphStyle`, `Justification`, `Hyphenation`, `HyphenationZone`, `FirstLineIndent`, `LeftIndent`, `GridAlignment`, `BulletsAndNumberingListType`, `RuleAboveLineWeight`, `RuleBelowLineWeight` y `SplitColumnInsideGutter`.
2. WHEN un `ParagraphStyleRange` de entrada declara un atributo distinto de los 11 declarados en el criterio 1, THE Modelo_Story SHALL preservarlo mediante OtherAttrs según el Requisito 2.
3. WHEN el Arnes_Fidelidad procesa las 30 stories del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias en las categorías definidas en el Requisito 1, criterio 2.
4. IF el Constructor_Documento recibe un párrafo cuyo estilo de párrafo aplicado no está definido entre los estilos del documento en construcción, THEN THE Constructor_Documento SHALL devolver un error que identifique el nombre del estilo solicitado, sin generar archivo.
5. IF un `ParagraphStyleRange` de entrada no contiene ningún `CharacterStyleRange`, THEN THE Serializador_XML SHALL emitirlo conservando sus atributos y sin agregar elementos hijo que no estuvieran en la entrada.

---

### Requisito 8: Recursos y archivos auxiliares tipados

**User Story:** Como usuario del generador, quiero que el generador emita la geometría del documento, las etiquetas XML y la backing story donde InDesign las espera, para que InDesign abra el archivo con el tamaño de página correcto.

#### Acceptance Criteria

1. THE Paquete_IDML SHALL modelar los cuatro elementos que el generador necesita emitir, `DocumentPreference`, `MarginPreference`, `ViewPreference` y `TextDefault`, distinguiendo las dos formas del elemento `MarginPreference`, que aparece con dos formas distintas en el Documento_Referencia:
   - en `Resources/Preferences.xml`, el `MarginPreference` de nivel de documento, con los 7 atributos `Top`, `Bottom`, `Left`, `Right`, `ColumnCount`, `ColumnGutter` y `ColumnDirection`, **sin** el atributo `ColumnsPositions`, que ese elemento no lleva;
   - en `Spreads/*.xml` y en `MasterSpreads/*.xml`, el `MarginPreference` de nivel de página, con los mismos 7 atributos más `ColumnsPositions`, es decir 8 atributos.

   THE Paquete_IDML SHALL declarar en `DocumentPreference` los atributos `PageHeight`, `PageWidth`, `PagesPerDocument` y `FacingPages`, y en `ViewPreference` los atributos `HorizontalMeasurementUnits`, `VerticalMeasurementUnits` y `RulerOrigin`.
2. THE Paquete_IDML SHALL emitir el elemento `DocumentPreference` como hijo del elemento raíz de `Resources/Preferences.xml` y SHALL emitir `designmap.xml` sin ningún elemento `DocumentPreference`, escribiendo en `PageHeight`, `PageWidth`, `PagesPerDocument` y `FacingPages` los valores recibidos del Constructor_Documento, cuyos valores en el Documento_Referencia son `954`, `774`, `1` y `true`.
3. WHEN `NewFromTemplate()` de `pkg/idml/templates.go` construye un paquete, THE Paquete_IDML SHALL producir un `designmap.xml` que contiene cero elementos `DocumentPreference` y un `Resources/Preferences.xml` que contiene exactamente un elemento `DocumentPreference`.
4. THE Paquete_IDML SHALL modelar `XML/Tags.xml` con elementos `XMLTag` de atributos `Self` y `Name` y con su hijo `Properties/TagColor`, cuyo atributo `type` tiene el valor `enumeration` y cuyo contenido de texto es el nombre del color.
5. THE Paquete_IDML SHALL modelar `XML/BackingStory.xml` con el elemento `XmlStory` de atributos `Self`, `UserText`, `IsEndnoteStory`, `AppliedTOCStyle`, `TrackChanges`, `StoryTitle` y `AppliedNamedGrid`, y con el elemento recursivo `XMLElement` de atributos `Self` y `MarkupTag`.
6. WHEN el Arnes_Fidelidad procesa `Resources/Preferences.xml`, `XML/Tags.xml` y `XML/BackingStory.xml` del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias en las categorías definidas en el Requisito 1, criterio 2.
7. THE Paquete_IDML SHALL preservar mediante OtherAttrs según el Requisito 2 los 276 atributos que el Documento_Referencia declara en `TextDefault`, sin declarar ninguno de ellos como campo del struct.
8. IF un `Resources/Preferences.xml` de entrada contiene cero elementos `DocumentPreference`, THEN THE Paquete_IDML SHALL devolver un error que identifique la ruta del archivo y el elemento ausente, sin devolver un modelo parcial.
9. IF el Constructor_Documento recibe un ancho de página o un alto de página menor o igual que 0, THEN THE Constructor_Documento SHALL devolver un error que identifique la dimensión y el valor recibido, sin generar archivo.
10. THE Paquete_IDML SHALL emitir `Resources/Preferences.xml` a partir del subconjunto curado, formado por los 10 elementos de primer nivel que ya declara la plantilla `pkg/idml/templates/minimal/Preferences.xml` (`ButtonPreference`, `PrintPreference`, `PrintBookletOption`, `PrintBookletPrintPreference`, `PageItemDefault`, `FrameFittingOption`, `StoryPreference`, `TextFramePreference`, `TextPreference`, `TextDefault`) más los tres elementos que esa plantilla no trae y el generador necesita, `DocumentPreference`, `MarginPreference` y `ViewPreference`, construidos con los valores que recibe el Constructor_Documento, y SHALL NOT copiar el `Resources/Preferences.xml` del Documento_Referencia. Adobe InDesign aplica sus propios valores por defecto a toda preferencia ausente, según el `README.md` de la propia plantilla, de modo que el subconjunto solo necesita ser válido, no completo.
11. THE Paquete_IDML SHALL emitir `TextDefault` con los 9 atributos que declara la plantilla mínima, `FirstLineIndent`, `LeftIndent`, `RightIndent`, `SpaceBefore`, `SpaceAfter`, `Justification`, `FontStyle`, `PointSize` y `AppliedFont`, y SHALL NOT declarar los demás atributos que el Documento_Referencia lleva en ese elemento, que se preservan al leer mediante OtherAttrs según el criterio 7 pero no se emiten al generar desde cero.
12. THE Constructor_Documento SHALL recibir como opciones los atributos `Intent` y `PageBinding` de `DocumentPreference`, de modo que el documento generado no quede restringido a la intención de impresión (`Intent="PrintIntent"`) ni a la encuadernación de izquierda a derecha (`PageBinding="LeftToRight"`) que declara el Documento_Referencia, y SHALL emitir los desplazamientos de sangrado y de anotación con el valor `0` cuando las opciones no los especifican.
13. THE Constructor_Documento SHALL recibir como opciones las unidades de medida de `ViewPreference` —`HorizontalMeasurementUnits`, `VerticalMeasurementUnits`, `LineMeasurementUnits`, `TypographicMeasurementUnits`, `TextSizeMeasurementUnits`, `StrokeMeasurementUnits` y `PrintDialogMeasurementUnits`— con el valor por defecto `Points`, porque el punto es la unidad estándar del sector editorial, y SHALL NOT fijar en el código los valores `Picas` y `Millimeters` que declara el Documento_Referencia.
14. WHEN el Constructor_Documento genera un documento, THE Paquete_IDML SHALL producir un `Resources/Preferences.xml` que Adobe InDesign abre sin reportar daño ni recuperación del documento (verificación manual: requiere abrir el archivo en Adobe InDesign y registrar el resultado).
15. IF Adobe InDesign rechaza el `Resources/Preferences.xml` generado, THEN THE Paquete_IDML SHALL identificar el atributo ausente comparando contra el `Resources/Preferences.xml` del Documento_Referencia usado como oráculo de diagnóstico, y SHALL agregar únicamente ese atributo al subconjunto curado, sin copiar el resto de ese archivo.
16. THE Constructor_Documento SHALL calcular el atributo `ColumnsPositions` del `MarginPreference` de nivel de página a partir del ancho de página, los márgenes, el número de columnas y el medianil, y SHALL NOT copiar el valor que declara el Documento_Referencia, porque el generador debe servir a cualquier formato y a cualquier número de columnas.

---

### Requisito 9: API pública de ensamblado del paquete

**User Story:** Como desarrollador que usa la librería, quiero agregar spreads, master spreads y stories a un paquete y que las referencias internas queden coherentes, sin tocar campos privados.

#### Acceptance Criteria

1. THE Ensamblador_Paquete SHALL exponer operaciones públicas para agregar y para eliminar un spread, un master spread, una story y la backing story de un Paquete_IDML, sin requerir acceso a los identificadores privados `files`, `fileOrder`, `setFileData` ni `addFileFromTemplate`.
2. WHEN se agrega un spread, un master spread o una story, THE Ensamblador_Paquete SHALL escribir en el `designmap.xml` la referencia correspondiente entre `<idPkg:Spread src>`, `<idPkg:MasterSpread src>` y `<idPkg:Story src>`, con el valor `src` igual a la ruta del archivo agregado dentro del paquete.
3. WHEN se agrega una story, THE Ensamblador_Paquete SHALL agregar el atributo `Self` de esa story al `StoryList` del elemento que la referencia.
4. WHEN se elimina un spread, un master spread o una story, THE Ensamblador_Paquete SHALL eliminar del `designmap.xml` la referencia de ese archivo, SHALL eliminar del `StoryList` el identificador de esa story, y SHALL dejar cero apariciones de ese identificador y de esa ruta en el `designmap.xml`.
5. THE Generador_IDs SHALL producir cada identificador `Self` con el formato de una letra `u` en minúscula seguida de entre 1 y 8 dígitos hexadecimales en minúscula, que es el formato que emite InDesign (por ejemplo `uce7`, `u1bc8`).
6. THE Ensamblador_Paquete SHALL ordenar las entradas del ZIP con `mimetype` como primera entrada y almacenada sin compresión, y THE Ensamblador_Paquete SHALL incluir `META-INF/container.xml` y `META-INF/metadata.xml`.
7. WHEN un Paquete_IDML se construye desde cero y no contiene `META-INF/metadata.xml`, THE Ensamblador_Paquete SHALL crear ese archivo con el paquete XMP correspondiente.
8. IF una operación de ensamblado recibe un identificador ya presente en el Paquete_IDML, THEN THE Ensamblador_Paquete SHALL devolver un error que identifique la operación y el identificador duplicado.
9. THE Generador_IDs SHALL registrar los identificadores `Self` leídos al abrir el Paquete_IDML y los emitidos después para ese mismo paquete, y SHALL emitir cada identificador nuevo distinto de todos los registrados.
10. IF una operación de eliminación recibe un identificador o una ruta que no está presente en el Paquete_IDML, THEN THE Ensamblador_Paquete SHALL devolver un error que identifique la operación y el valor ausente, sin modificar el `designmap.xml`, el `StoryList` ni el orden de entradas del ZIP.

---

### Requisito 10: Constructor de documento desde cero

**User Story:** Como desarrollador que usa la librería, quiero construir un documento IDML completo desde structs de Go, para generar archivos sin partir de ningún IDML existente.

#### Acceptance Criteria

1. THE Constructor_Documento SHALL exponer un flujo de construcción compuesto por las operaciones públicas `NewDocument`, `AddMasterSpread`, `AddPage`, `AddTextFrame`, `AddImageFrame`, `AddShape`, `AddGroup` y `Build`, donde `Build` devuelve el Paquete_IDML ensamblado.
2. THE Constructor_Documento SHALL admitir como elementos de página marcos de texto, marcos de imagen, formas, líneas y grupos.
3. THE Constructor_Documento SHALL admitir la definición de capas, secciones, guías y estilos de párrafo, de carácter y de objeto.
4. THE Constructor_Documento SHALL asignar los elementos de página a la capa indicada y THE Constructor_Documento SHALL emitirlos en el Orden_Documental en que fueron agregados.
5. THE Constructor_Documento SHALL calcular `ItemTransform` y `PathGeometry` de cada elemento de página a partir de su posición y tamaño usando `pkg/spread/geometry.go`.
6. THE Constructor_Documento SHALL generalizar la construcción programática ya existente en `pkg/idms/exporter_build.go` (documento, spread, colores, swatches, estilos de trazo, grupos de estilos, capas y grupos de color por defecto) y compartirla entre la exportación IDMS y la generación IDML.
7. THE Constructor_Documento SHALL recibir su configuración mediante structs de opciones.
8. WHEN el Constructor_Documento genera un documento con los tipos de elemento del Documento_Referencia, THE Paquete_IDML SHALL producir un archivo que Adobe InDesign abre sin reportar daño ni recuperación del documento (verificación manual: requiere abrir el archivo en Adobe InDesign y registrar el resultado).
9. IF una opción obligatoria falta o tiene un valor fuera de rango, THEN THE Constructor_Documento SHALL devolver un error que identifique la opción y el valor recibido, sin generar archivo.
10. WHEN el Constructor_Documento genera un documento con los tipos de elemento del Documento_Referencia, THE Paquete_IDML SHALL producir un paquete que contiene las entradas `mimetype`, `META-INF/container.xml`, `META-INF/metadata.xml`, `designmap.xml`, `Resources/Fonts.xml`, `Resources/Graphic.xml`, `Resources/Preferences.xml`, `Resources/Styles.xml`, `XML/Tags.xml`, `XML/BackingStory.xml`, al menos un archivo bajo `MasterSpreads/`, al menos un archivo bajo `Spreads/` y un archivo bajo `Stories/` por cada story referenciada en el `designmap.xml`.
11. WHEN el paquete generado por el Constructor_Documento se vuelve a abrir y se procesa con el Arnes_Fidelidad, THE Serializador_XML SHALL producir cero diferencias en las categorías definidas en el Requisito 1, criterio 2, respecto de los archivos del paquete generado.
12. THE Paquete_IDML SHALL producir con `NewFromTemplate()` un paquete que contiene las entradas del criterio 10 y cuyo `designmap.xml` declara las referencias `idPkg:Spread`, `idPkg:Story` e `idPkg:BackingStory`, porque un paquete cuyo `designmap.xml` no declara ninguna referencia `idPkg:Spread` no contiene ninguna página.
13. THE Paquete_IDML SHALL cerrar en el paquete generado las cuatro referencias cruzadas que un documento mínimo exige: el atributo `ItemLayer` de cada elemento de página SHALL nombrar un elemento `Layer` declarado en el `designmap.xml`, el atributo `ParentStory` de cada marco de texto SHALL nombrar una story emitida bajo `Stories/`, el atributo `PageStart` de cada `Section` SHALL nombrar una página emitida, y el atributo `StoryList` del elemento `Document` SHALL contener el identificador de cada story emitida y el de la backing story.
14. WHEN el paquete producido por `NewFromTemplate()` se abre en Adobe InDesign, THE Paquete_IDML SHALL abrirse sin que Adobe InDesign reporte daño ni recuperación del documento (verificación manual: requiere abrir el archivo en Adobe InDesign y registrar el resultado). THE Paquete_IDML SHALL someterse a esta verificación antes de implementar el Constructor_Documento, de modo que actúe como comprobación temprana en lugar de la del criterio 8, que llega al final de la cadena de tareas.

    **Resultado registrado (Tarea 3): superada.** Dos sondas A4, de 1 y de 5 columnas, abren en Adobe InDesign sin aviso de daño ni petición de recuperación, y muestran la página con su marco de texto dentro de los márgenes. Abren también en Affinity Publisher. Los criterios 12 y 13 quedan verificados por tests automáticos: el conjunto de 13 entradas coincide con `testdata/plain.idml` y las cuatro referencias cruzadas se resuelven contra el elemento al que apuntan.

---

### Requisito 11: Resolución de imágenes desde ruta local o carga base64

**User Story:** Como usuario del generador, quiero indicar imágenes por ruta de archivo local o incrustando sus bytes en base64 dentro del JSON, para cubrir tanto el caso en que la imagen ya está en disco como el caso en que quien llama tiene los bytes en memoria.

El modo enlazado que este requisito describe es el **modo secundario**: sirve cuando el documento generado se abre en la misma máquina que contiene los archivos de imagen. El modo por defecto es el embebido del Requisito 12, porque `LinkResourceURI` solo admite URI `file:` que apunta a la máquina que generó el documento, y un documento que viaja tendría enlaces rotos.

El Resolvedor_Imagenes **no realiza ninguna petición de red**: las dos formas de origen admitidas son locales al proceso.

#### Acceptance Criteria

1. THE Resolvedor_Imagenes SHALL aceptar como origen de imagen únicamente dos formas: una ruta de archivo local o una carga de bytes codificada en base64 dentro de la propia entrada JSON.
2. THE Resolvedor_Imagenes SHALL obtener el ancho y alto en píxeles y el formato de la imagen usando `image.DecodeConfig` de la biblioteca estándar.
3. THE Resolvedor_Imagenes SHALL calcular `ActualPpi`, `EffectivePpi` y el `ItemTransform` de escala del elemento `Image` a partir de las dimensiones obtenidas y del tamaño del marco de destino.
4. WHERE el origen es una ruta local, THE Resolvedor_Imagenes SHALL restringir la lectura a un directorio base configurado y THE Resolvedor_Imagenes SHALL rechazar las rutas que, tras resolver enlaces simbólicos y segmentos relativos con `filepath.EvalSymlinks` y comprobar el prefijo resultante, queden fuera de ese directorio base.
5. IF el origen es una ruta local y el directorio base del criterio 4 no está configurado, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique la opción de configuración ausente, sin leer ningún archivo.
6. WHERE el origen es una carga base64, THE Resolvedor_Imagenes SHALL limitar el tamaño decodificado por imagen a un máximo con valor por defecto de 64 MiB, configurable mediante el struct de opciones del Resolvedor_Imagenes.
7. IF el tamaño decodificado de una carga base64 excede el límite del criterio 6, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique la imagen y el límite superado, sin generar archivo.
8. IF el origen recibido es una cadena vacía, o no declara ninguna de las dos formas del criterio 1, o declara ambas a la vez, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique el origen y la forma recibida, sin leer ningún archivo.
9. IF la imagen no puede leerse, no puede decodificarse desde base64 o no puede decodificarse como imagen, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique el origen y el motivo.
10. THE Resolvedor_Imagenes SHALL admitir el modo de imagen enlazada como modo secundario, seleccionable explícitamente y pensado para el uso local de la librería desde Go, generando el elemento `Image` con sus hijos `Link`, `FrameFittingOption` y `ClippingPathSettings`, y copiando el archivo resuelto a una carpeta de recursos junto al documento generado.
11. THE Resolvedor_Imagenes SHALL emitir el atributo `LinkResourceURI` del elemento `Link` exclusivamente como una URI de esquema `file:`, porque IDML no admite ningún otro esquema en los enlaces.
12. THE Resolvedor_Imagenes SHALL NOT realizar ninguna petición de red al resolver una imagen, en ninguna de las dos formas de origen del criterio 1.

---

### Requisito 12: Imágenes embebidas

**User Story:** Como usuario del CLI_Generador, quiero que las imágenes viajen dentro del archivo IDML por defecto, para poder mover el documento sin arrastrar una carpeta de enlaces.

El formato de la imagen embebida está verificado empíricamente contra el Archivo_Evidencia_Imagenes: dos imágenes embebidas (PNG y JPEG) y una enlazada en un mismo spread. Este requisito especifica ese formato; ya no es una investigación.

#### Acceptance Criteria

1. THE Resolvedor_Imagenes SHALL emitir los bytes de la imagen embebida como contenido de texto del hijo `Properties/Contents` del elemento `Image`, colocando `Contents` entre `Profile` y `GraphicBounds` dentro de `Properties`.
2. THE Resolvedor_Imagenes SHALL codificar los bytes de la imagen con base64 estándar (RFC 4648) usando el paquete `encoding/base64` de la biblioteca estándar de Go, y SHALL ajustar la salida a 76 caracteres por línea, con la última línea de longitud menor o igual a 76, sin salto de línea inicial ni final.
3. THE Resolvedor_Imagenes SHALL emitir el atributo `StoredState` con el valor `Embedded` en el hijo `Link` del elemento `Image`, y SHALL conservar ese elemento `Link` con su atributo `LinkResourceURI`.
4. THE Resolvedor_Imagenes SHALL emitir el atributo `LinkResourceSize` con la forma `0~<hex>`, donde `<hex>` es el número de bytes decodificados de la imagen expresado en hexadecimal (por ejemplo `0~2093` para 8.339 bytes y `0~ec64` para 60.516 bytes).
5. THE Resolvedor_Imagenes SHALL emitir el elemento `Image` con el mismo conjunto de atributos en modo embebido que en modo enlazado, sin agregar los atributos `ContentsEncoding`, `ContentsType` ni `ContentsVersion`, que Adobe InDesign no emite en `Image`.
6. THE Resolvedor_Imagenes SHALL usar el modo embebido como modo por defecto, y SHALL permitir seleccionar explícitamente el modo enlazado del Requisito 11 por cada imagen de la entrada.
7. WHEN el Paquete_IDML parsea un archivo IDML existente que contiene una imagen embebida, THE Serializador_XML SHALL preservar el texto de `Properties/Contents` de forma literal, sin decodificarlo ni volver a codificarlo, de modo que el ajuste de línea de la entrada sobreviva el ciclo de parseo y serialización.
8. WHERE el Constructor_Documento genera un documento desde cero, THE Arnes_Fidelidad SHALL NOT tratar el ancho de ajuste de línea del base64 como una diferencia, porque la integridad de la imagen es independiente de él: decodificar la misma carga sin ajuste, a 64, a 76 y a 120 caracteres produce bytes idénticos.
9. WHEN el documento generado con una imagen embebida se abre en Adobe InDesign, THE Paquete_IDML SHALL mostrar la imagen sin solicitar la relocalización del enlace (verificación manual: requiere abrir el archivo en Adobe InDesign y registrar el resultado).
10. IF los bytes de la imagen no pueden codificarse o la imagen resuelta está vacía, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique el origen y el motivo, sin generar archivo.

---

### Requisito 13: Interfaz de línea de comandos de generación

**User Story:** Como operador o como proceso llamador, quiero pasar la descripción de un documento en JSON por la entrada estándar de un binario y recibir el archivo IDML por la salida estándar, para integrar la generación en cualquier flujo sin escribir Go y sin levantar un servicio.

#### Acceptance Criteria

1. THE CLI_Generador SHALL leer de la entrada estándar la descripción del documento en JSON.
2. THE CLI_Generador SHALL escribir el archivo IDML generado en la salida estándar cuando no se indica una ruta de salida.
3. WHEN se invoca el CLI_Generador con la opción `-out <ruta>`, THE CLI_Generador SHALL escribir el archivo IDML generado en esa ruta en lugar de la salida estándar.
4. THE CLI_Generador SHALL representar en el JSON de entrada las mismas capacidades que expone el Constructor_Documento según el Requisito 10.
5. THE Validador_Entrada SHALL validar el JSON recibido antes de invocar al Constructor_Documento, verificando el tamaño de página, el número de páginas, el número de elementos por página, la longitud del texto por elemento y la forma del origen de cada imagen según el Requisito 11, criterio 1.
6. WHEN la generación termina sin error, THE CLI_Generador SHALL terminar con código de salida `0`.
7. IF la entrada incumple una regla de validación, THEN THE CLI_Generador SHALL terminar con código de salida `1` y escribir en la salida de error estándar un mensaje que identifique el campo y la regla incumplida, sin invocar al Constructor_Documento.
8. IF la entrada estándar no contiene un documento JSON válido, THEN THE CLI_Generador SHALL terminar con código de salida `1` y escribir en la salida de error estándar la posición del error de sintaxis.
9. IF el JSON declara una construcción fuera del alcance enumerado en el Requisito 14, criterio 4, THEN THE CLI_Generador SHALL terminar con código de salida `1` identificando la construcción y el campo que la declara, sin invocar al Constructor_Documento.
10. IF la generación del documento falla, THEN THE CLI_Generador SHALL terminar con código de salida `2` y escribir el detalle del fallo en la salida de error estándar.
11. THE CLI_Generador SHALL construir sus errores con los tipos de `pkg/common/errors.go`.
12. THE CLI_Generador SHALL escribir en la salida estándar exclusivamente los bytes del archivo IDML, y SHALL escribir todo diagnóstico, advertencia y error en la salida de error estándar.
13. THE CLI_Generador SHALL NOT realizar ninguna petición de red.
14. THE CLI_Generador SHALL escribir en español los mensajes de diagnóstico y de error.
15. THE CLI_Generador SHALL usar el modo de imagen embebida del Requisito 12 como modo por defecto de toda imagen de la entrada, porque `LinkResourceURI` solo admite URI de esquema `file:` que apunta a la máquina que generó el documento, y SHALL admitir el modo enlazado únicamente cuando la entrada lo selecciona explícitamente.

---

### Requisito 14: Cierre de fidelidad y límites declarados

**User Story:** Como mantenedor de la librería, quiero que la prueba de fidelidad de roundtrip esté activa y que el alcance no cubierto esté escrito, para no asumir una cobertura que no existe.

#### Acceptance Criteria

1. THE Arnes_Fidelidad SHALL ejecutar `TestGoldenRoundtrip_ExampleIDML` de `pkg/idml/golden_test.go` sin la llamada `t.Skip` actual, sustituyendo la comparación byte a byte por Comparación_Estructural archivo por archivo, sobre el corpus del Requisito 1, criterio 9, que incluye el Documento_Referencia y el Archivo_Evidencia_Imagenes.
2. THE Arnes_Fidelidad SHALL comparar byte a byte únicamente los archivos que el Paquete_IDML copia sin parsear.
3. THE Arnes_Fidelidad SHALL verificar que el archivo generado contiene `mimetype` como primera entrada del ZIP y almacenada sin compresión.
4. THE Paquete_IDML SHALL enumerar en el comentario de paquete de `pkg/idml` las construcciones declaradas fuera de alcance de esta funcionalidad: tablas (`Table`, `Cell`), notas al pie (`Footnote`), hipervínculos (`Hyperlink`), referencias cruzadas (`CrossReferenceSource`, `CrossReferenceFormat`) y tipografía CJK (`KinsokuTable`, `MojikumiTable`).
5. WHEN el JSON recibido por el CLI_Generador declara una construcción enumerada en el criterio 4, THE Validador_Entrada SHALL rechazar la entrada con un error que identifique la construcción y el campo del JSON que la declara, sin invocar al Constructor_Documento.
6. WHEN el Paquete_IDML abre un archivo IDML existente que contiene una construcción enumerada en el criterio 4, THE Paquete_IDML SHALL preservarla mediante los mecanismos de elementos y atributos no modelados de los Requisitos 2 y 3 y SHALL emitirla al reescribir el paquete con las mismas categorías del Requisito 1, criterio 2, en cero diferencias.
7. THE Validador_Entrada SHALL aplicar el rechazo del criterio 5 únicamente al JSON recibido por el CLI_Generador, y THE Paquete_IDML SHALL aplicar la preservación del criterio 6 únicamente a los archivos leídos de un paquete IDML existente, sin que el rechazo del criterio 5 impida abrir, modificar y reescribir esos archivos.

---

### Requisito 15: Restricciones técnicas del proyecto

**User Story:** Como mantenedor de la librería, quiero que la funcionalidad se construya con lo que ya hay en el repositorio, para no aumentar la carga de mantenimiento.

#### Acceptance Criteria

1. THE Paquete_IDML SHALL implementar esta funcionalidad usando la biblioteca estándar de Go y las dependencias ya declaradas en `go.mod` (`github.com/beevik/etree`, `goldie/v2`, `gopter`, `go-cmp`), sin agregar dependencias nuevas.
2. THE Paquete_IDML SHALL declarar la versión de Go `1.23` en `go.mod` y en la clave `go` de `.golangci.yml`, reemplazando el `1.21.13` que hoy declara `go.mod`, porque `1.23` es la versión con la que ya se ejecutan el linter y la integración continua.
3. THE Paquete_IDML SHALL seguir las convenciones de `ARCHITECTURE.md` para etiquetas XML, elementos comodín `OtherElements`, nombres de tipos y campos, y manejo de errores con `pkg/common/errors.go`.
4. THE CLI_Generador SHALL residir en `cmd/idmlgen`, como tercer binario junto a `cmd/cli` y `cmd/debug-export`, porque `ARCHITECTURE.md` coloca los transportes fuera de `pkg/`, en `cmd/`, y SHALL dejar en `pkg/` el Constructor_Documento y el Resolvedor_Imagenes que ese binario invoca.
5. THE Paquete_IDML SHALL exponer la configuración de los componentes nuevos mediante structs de opciones, sin introducir interfaces no solicitadas.
6. WHEN un incremento de esta funcionalidad agrega o modifica lógica en un paquete, THE Paquete_IDML SHALL incluir en ese mismo paquete al menos una prueba automatizada que falle si esa lógica se revierte.
7. THE Paquete_IDML SHALL escribir en español los comentarios de paquete, los mensajes de error y los documentos de este spec.
8. WHEN se completa un incremento de esta funcionalidad, THE Paquete_IDML SHALL terminar sin fallos la ejecución de `go test ./...` y la ejecución de `golangci-lint run`.

---

## Riesgos y decisiones pendientes

Estos puntos no están resueltos y deben confirmarse antes de implementar los requisitos que dependen de ellos. Cada uno declara su estado: **medido** cuando la incógnita se cerró con datos del repositorio, **mitigado** cuando el plan ya no depende de que la suposición sea correcta, y **pendiente de verificación manual** cuando la única resolución posible es abrir un archivo generado en Adobe InDesign. Los riesgos 2, 4 y 6 comparten esa misma verificación, y la Tarea 3 la concentra en un único punto de la cadena.

1. **Tamaño del cambio bloqueante del contenedor ordenado y longitud de la cadena (Requisito 3) — medido y mitigado en sus dos mitades; queda en esta sección hasta que la Tarea 3 se ejecute.** Los Requisitos 2 y 3 son prerequisitos de los Requisitos 4 a 14: implementar el Constructor_Documento antes de resolverlos produciría documentos que pierden atributos y orden de apilamiento.

   **Estado.** La primera mitad —el tamaño y la forma del cambio— está mitigada: el plan ya no deja la compilación en rojo en ninguna fase, el modo de fallo silencioso de las copias tiene barrera del compilador (los receptores puntero del criterio 13 del Requisito 3) más la verificación del criterio 14 que lo delata, y la reescritura ambigua se resolvió con el renombrado consciente de tipos que abre la fase 2. Lo que **no** desaparece es el trabajo: siguen siendo 236 sitios que hay que tocar. El riesgo era que fuera peligroso, no que fuera grande. La segunda mitad —la longitud de la cadena— está mitigada por la sonda de la Tarea 3, que adelanta de profundidad 11 a profundidad 2 la única comprobación capaz de invalidar los cimientos.

   **Residuo, y por eso el riesgo sigue aquí.** Nada de lo anterior sustituye a ejecutar el trabajo: la migración en tres fases hay que hacerla, y la sonda hay que abrirla en Adobe InDesign.

   **Mitad resuelta (Tarea 3).** La sonda ya se abrió: InDesign acepta la salida de `NewFromTemplate()` completada, así que **la parte del riesgo que consistía en construir diez tareas de modelo sobre suposiciones sin confrontar está cerrada**. Los cimientos no quedaron invalidados. **Lo que queda del riesgo es solo el trabajo**, los 236 sitios de la migración en tres fases de las Tareas 7 a 9, que ni se ha reducido ni se ha vuelto más peligroso. Este punto se mueve a «Riesgos resueltos» cuando esas tres fases estén hechas.

   **Por qué el cambio es inevitable y no una preferencia de diseño: ningún atributo lleva el orden de apilamiento (prueba por eliminación).** Buscando en el spread y en los master spreads del Documento_Referencia todo atributo cuyo nombre contiene `Order`, `Index`, `ZOrder`, `Position`, `Stack` o `Depth`, lo que existe es:

   | Atributo | Apariciones | Qué es |
   |---|---|---|
   | `BeforeGroupingLayerPosition` | 76, **todas con el valor `-1`** (un único valor distinto) | centinela, no un índice |
   | `Index="-1"` | 7 | pertenece a `ClippingPathSettings`, no al elemento de página |
   | `ContourPathIndex="-1"`, `ContourAlphaIndex="0"` | 14 cada uno | recorte de imagen |
   | `PageIndex` | 156 | asignación de página, no profundidad |

   Ninguno permite reconstruir el orden de apilamiento. De ahí la conclusión decisiva: **si el orden no está en ningún atributo, el único lugar donde puede vivir es la posición del elemento en el XML.** Una vez perdido en el parseo no se recupera, así que los slices por tipo no son un detalle de representación: destruyen información.

   **Qué se pierde, medido.** Los elementos de página están fuertemente intercalados por tipo. `Spreads/Spread_uce7.xml` tiene **32** hijos directos de `<Spread>`: `Rectangle`, `Rectangle`, `Group`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `Rectangle`, `Rectangle`, `Rectangle`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `Rectangle`, `Rectangle`, `Rectangle`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `Rectangle`, `TextFrame`, `GraphicLine`, `Group` (20 `TextFrame`, 9 `Rectangle`, 2 `Group`, 1 `GraphicLine`). `MasterSpreads/MasterSpread_u2dd.xml` tiene **17** hijos directos de `<MasterSpread>`: `GraphicLine`, `GraphicLine`, `GraphicLine`, `TextFrame`, `TextFrame`, `TextFrame`, `TextFrame`, `GraphicLine`, `GraphicLine`, `Rectangle`, `Rectangle`, `TextFrame`, `Rectangle`, `Rectangle`, `GraphicLine`, `GraphicLine`, `GraphicLine` (8 `GraphicLine`, 5 `TextFrame`, 4 `Rectangle`). Los otros dos master spreads tienen **8** (`MasterSpread_u409.xml`: 4 `GraphicLine`, 4 `TextFrame`) y **0** elementos de página (`MasterSpread_ub0.xml`, que tiene páginas pero ningún contenido).

   Reordenar la secuencia por tipo, que es lo que producen los slices por tipo actuales, desplaza:

   | Archivo | Elementos | Conservan posición | Cambian de posición |
   |---|---|---|---|
   | `Spreads/Spread_uce7.xml` | 32 | 19 | **13** |
   | `MasterSpreads/MasterSpread_u2dd.xml` | 17 | 5 | **12** |

   El master spread sale proporcionalmente peor: 12 de 17.

   **El anidamiento importa: `<Image>` no es hermano.** Los **7** elementos `<Image>` de `Spread_uce7.xml` tienen todos `<Rectangle>` como padre inmediato; cero aparecen como hijos directos de `<Spread>`. El contenedor ordenado de nivel de spread guarda los marcos, y la imagen vive dentro de su marco. El struct `Rectangle` de `pkg/spread/page_items.go` ya lo modela así, declarando `Image *Image` en la línea 43. Mediciones anteriores que trataban `<Image>` como hermano estaban equivocadas. Este documento **no** afirma que IDML prohíba `<Image>` como hijo directo de un spread: eso no se midió, solo se registra lo observado.

   **Documentación, consistente con la medición.** Contenido reformulado por cumplimiento de licencia:

   - [Respuesta de StackOverflow sobre extraer texto de IDML en orden](https://stackoverflow.com/questions/15154675/idml-extract-text-content-in-proper-order): el orden de los elementos `TextFrame` en un spread IDML indica su profundidad de apilamiento, no el orden de lectura de la página. *(Contenido reformulado por cumplimiento de licencia.)*
   - [Documentación del DOM de InDesign para Spread](https://developer.adobe.com/indesign/dom/api/s/Spread/): describe la colección `pageItems` como la que procesa todos los elementos de un contenedor sin distinguir el tipo, consistente con un orden posicional. *(Contenido reformulado por cumplimiento de licencia.)*
   - [Hilo del foro de Adobe sobre el orden de apilamiento en CS5](https://community.adobe.com/questions-671/stacking-order-of-pageitems-in-cs5-849095): describe el mismo síntoma que produce agrupar por tipo, que una secuencia intercalada salga con los tipos juntos. *(Contenido reformulado por cumplimiento de licencia.)*

   **Límite honesto:** no se encontró una frase de Adobe que afirme literalmente que el orden documental del archivo IDML es el orden de apilamiento. La prueba fuerte es la ausencia de atributo (la tabla de arriba); la documentación solo es consistente con ella. Nadie debe citar la especificación como si Adobe lo hubiera confirmado por escrito.

   **Tamaño del cambio, medido sobre el código Go, y corregido.** Recorriendo todos los archivos `.go` de `pkg/`, `cmd/` e `internal/` en busca de accesos a los siete identificadores que el contenedor ordenado reemplaza (`TextFrames`, `Rectangles`, `Images`, `Ovals`, `Polygons`, `GraphicLines`, `Groups`) hay **319 accesos, todos ellos accesos a campo y ninguna llamada a método**: los accesores por tipo que el diseño plantea son API entera nueva que hoy nadie usa. Pero **83 de esos 319 no migran.** El struct `Selection` de `pkg/idml/selection.go` líneas 38 a 53 declara **los mismos seis nombres de campo**, y esos 83 accesos son suyos: 36 dentro de los métodos de `Selection`, 18 en `pkg/idms/exporter_build.go`, 14 en `pkg/idml/selection_test.go`, 6 en `pkg/analysis/tracker.go`, 6 en `pkg/idms/exporter.go` y 3 en `pkg/analysis/tracker_test.go`; 6 de ellos son escrituras. La cifra de 323 de una medición anterior de este documento conflacionaba los dos structs.

   **La superficie real de migración es de 236 accesos:** 80 en producción y 156 en tests, repartidos en 20 archivos de 6 paquetes, incluidos los dos binarios existentes `cmd/cli` y `cmd/debug-export`. De ellos, **18 son escrituras**, concentradas en dos archivos de producción y ninguna en tests: `pkg/idms/exporter_build.go` 12 (seis `make(...)` y seis asignaciones indexadas) y `pkg/idml/package_modifications.go` 6 (incluido el modismo `append(s[:i], s[i+1:]...)` de las líneas 151 y 158 para eliminar). Los otros 218 son lecturas.

   | Identificador | Accesos del modelo |
   |---|---|
   | `TextFrames` | 99 |
   | `Rectangles` | 92 |
   | `Ovals` | 12 |
   | `Polygons` | 11 |
   | `GraphicLines` | 11 |
   | `Groups` | 9 |
   | `Images` | 2 |

   **La consecuencia peligrosa: 39 accesos toman la dirección del elemento, 13 de ellos en producción.** Hoy los campos guardan valores (`TextFrames []SpreadTextFrame`, `pkg/spread/spread.go` línea 142), así que `&sp.InnerSpread.TextFrames[i]` es un puntero al almacenamiento del documento, y hay código que depende de eso: `pkg/idml/index.go` toma 6 de esos punteros y los **guarda en un índice de larga vida** (`p.indexState.index.textFrames[tf.Self] = tf`), `pkg/idml/resourcemgr_cleanup.go` toma 4 y muta a través de ellos, `pkg/idml/selection.go` toma 2 y `cmd/cli/tui/export_idms.go` 1; los 26 restantes están en tests. Si un accesor devolviera un slice de **valores** construido al filtrar el contenedor ordenado, `&x.TextFrames()[i]` **compilaría** —indexar un slice es una operación direccionable en Go aunque el slice venga de una llamada— pero apuntaría a una copia temporal, y las mutaciones dejarían de llegar al documento **sin error de compilación** y con la suite de pruebas en verde. Es la peor forma de romper el modelo, y de ahí salen los criterios 4, 13 y 14 del Requisito 3: los accesores devuelven punteros, el contenedor guarda punteros, la interfaz se declara con receptores puntero para que el compilador rechace guardar una copia, y una verificación muta a través del accesor y comprueba el XML re-serializado.

   **Restricción técnica que fuerza la forma del cambio:** Go no permite que un campo de struct y un método del mismo tipo compartan nombre, así que `TextFrames` no puede ser campo y accesor a la vez durante una transición.

   **Por qué la reescritura no puede ser `gofmt -r` a secas, y cómo se arregla.** `gofmt -r` reescribe por forma de expresión y es ciego a los tipos: `sel.TextFrames` y `sp.InnerSpread.TextFrames` son la misma forma, así que arrastraría los 83 accesos de `Selection`, y las 6 escrituras de entre ellos quedarían como `s.TextFrames() = append(...)`, que ni parsea. La mitigación es hacer primero un **renombrado consciente de tipos**: renombrar los seis campos de `SpreadElement` a un nombre único temporal toca exactamente los 236 accesos del modelo y ninguno de `Selection`; con el campo renombrado, el accesor con el nombre final ya no colisiona; y la conversión final de ese nombre único sí es textual y sin ambigüedad, porque existe en un solo tipo del repositorio. Con esa secuencia la compilación queda en verde después de cada una de las tres tareas y **desaparece la fase en rojo** del plan anterior.

   **Proporción del trabajo:** de los 236 accesos, los 18 de escritura y los 39 que toman la dirección del elemento son los que exigen criterio, porque cambian de forma y no solo de sintaxis (`&X[i]` pasa a `X()[i]`, sin el `&`). El resto es conversión mecánica de acceso a campo a llamada de accesor.

   **La mitad de ordenación del riesgo ya está mitigada por el orden de tareas:** las deudas estructurales van antes que todo lo demás y el Arnes_Fidelidad es el punto de control temprano que entrega la línea base numérica antes de tocar el modelo. Lo que queda es el tamaño del único cambio bloqueante, y se aborda partiéndolo en tres tareas por fases con compilación en verde después de cada una de las tres, gracias al renombrado consciente de tipos que abre la fase 2.

   **La segunda mitad del riesgo: la cadena es larga y serial, y se mitiga con una sonda temprana.** Medido sobre las dependencias declaradas, la ruta crítica hasta el primer IDML generado tiene profundidad 11 y hasta el primer `Resources/Preferences.xml` emitido, profundidad 8. El problema no es la longitud sino qué cuelga de ella: los riesgos 2, 4 y 6 comparten la misma resolución pendiente, abrir un archivo generado en Adobe InDesign, así que se construirían diez tareas de modelo sobre tres suposiciones sin confrontar.

   **La mitigación no agrega un componente: completa `NewFromTemplate()`, que ya existe y hoy produce un paquete sin ni una página** (9 archivos, y su `designmap.xml` sin ninguna referencia `idPkg:Spread`, una forma que ninguna de las tres exportaciones de InDesign del repositorio tiene). Su camino de escritura ya está probado de extremo a extremo por `TestNewFromTemplate_Roundtrip`, que escribe con `idml.Write` y reabre el paquete; lo único nunca verificado es si InDesign abre su salida. Completarla exige 4 archivos y 3 referencias de `designmap.xml`, y tiene oráculo gratis: `testdata/plain.idml` es una exportación de InDesign de exactamente esa forma —1 página, 1 marco de texto, 13 entradas— ya versionada en el repositorio. Los criterios 12, 13 y 14 del Requisito 10 recogen el resultado y la sonda queda a profundidad 2, así que la verificación que cierra los riesgos 2, 4 y 6 deja de ser lo último y pasa a ser lo tercero.

   **Contrapartida aceptada:** el código de la sonda se escribe antes del contenedor ordenado, así que usa los slices por tipo y suma unos 2 sitios de escritura a los 18 que migra la fase 2. El Arnes_Fidelidad sigue siendo lo que mantiene medible el avance del modelo mientras dura la cadena.

2. **Conjunto mínimo de archivos que exige InDesign (Requisito 10, criterio 10) — medido por intersección, pendiente de verificación manual.** La lista ya no se apoya en un único documento: es la **intersección de tres exportaciones independientes de InDesign** de complejidad creciente, `testdata/plain.idml` (13 entradas, documento casi vacío), `testdata/example.idml` (27 entradas) y el Documento_Referencia (43 entradas, página de periódico). Las tres contienen exactamente las mismas **10 rutas fijas** —`mimetype`, `designmap.xml`, `META-INF/container.xml`, `META-INF/metadata.xml`, `Resources/Fonts.xml`, `Resources/Graphic.xml`, `Resources/Preferences.xml`, `Resources/Styles.xml`, `XML/Tags.xml` y `XML/BackingStory.xml`— más al menos un archivo bajo `MasterSpreads/`, al menos uno bajo `Spreads/` y uno bajo `Stories/` por cada story referenciada. En las tres, el `designmap.xml` declara referencias `idPkg:BackingStory`, `idPkg:Spread`, `idPkg:Story` e `idPkg:MasterSpread` (3 en el Documento_Referencia, 1 en las otras dos).

   **Tres de esas rutas son estructuralmente obligatorias, no una convención.** `META-INF/container.xml` usa el espacio de nombres de contenedor de OASIS OpenDocument (`urn:oasis:names:tc:opendocument:xmlns:container`) y su único `rootfile` nombra `designmap.xml`: dentro de un ZIP no existe otro mecanismo para señalar cuál es el archivo raíz, así que `mimetype`, `META-INF/container.xml` y `designmap.xml` no pueden faltar.

   **La brecha de la plantilla actual, medida.** `NewFromTemplate()` emite **9** archivos y le faltan **4** de esas categorías: `META-INF/metadata.xml`, `Spreads/`, `XML/BackingStory.xml` y `Stories/`. Su `designmap.xml` declara 6 referencias `idPkg:` y **ninguna `idPkg:Spread`**, es decir, su salida no tiene ni una página. Ninguna de las tres exportaciones de InDesign tiene esa forma. El Requisito 9, criterio 7 ya cubre la creación de `META-INF/metadata.xml` al construir desde cero.

   **`Stories/` depende del contenido, y el criterio ya lo refleja.** `plain.idml` trae una story porque su spread contiene un `TextFrame`; un documento sin texto no tendría ninguna. Por eso el criterio 10 condiciona `Stories/` a las stories referenciadas en el `designmap.xml` en lugar de exigir un número fijo.

   **Límite honesto:** no hay lista oficial vigente contra la que validar, porque Adobe retiró la especificación de IDML de su sitio; en el [hilo del foro de Adobe donde se pregunta por ella](https://community.adobe.com/t5/indesign-discussions/where-is-the-idml-specification/m-p/13172643) la única copia disponible la adjunta un usuario, no Adobe. *(Contenido reformulado por cumplimiento de licencia.)* La evidencia es la intersección medida, no documentación. **Lo que queda pendiente** es únicamente la verificación manual, la única que puede revelar una entrada obligatoria que no aparezca en ninguna de las tres exportaciones. Ya no espera al final de la cadena: el criterio 14 del Requisito 10 la coloca sobre la salida de `NewFromTemplate()` completada, a profundidad 2 en lugar de 11. Si aparece una entrada adicional, la lista se amplía y se agrega un test de regresión.

   **RESUELTO (Tarea 3).** La verificación manual se hizo y la pasó: la salida de `NewFromTemplate()`, con las 10 rutas fijas más una de `MasterSpreads/`, una de `Spreads/` y una de `Stories/`, 13 entradas en total, se abre en Adobe InDesign sin aviso de daño ni petición de recuperación. **No apareció ninguna entrada obligatoria adicional**, así que la lista medida por intersección era correcta y no se amplía. Abre también en Affinity Publisher.

3. **Tamaño máximo decodificado por imagen base64 (Requisito 11, criterio 6) — cota recalibrada de 20 a 64 MiB.** La cota anterior de 20 MiB era demasiado ajustada, y el propio Documento_Referencia lo demuestra: sus 7 imágenes declaran en `LinkResourceSize` 3,06, 3,85, 4,32, 4,98, 5,29, 6,58 y **14,74 MiB** (42,83 MiB en total). Con 20 MiB, la imagen más grande del documento que define el alcance ocupaba el **74 %** de la cota: 1,36× de margen, sin espacio para una foto de portada a mayor resolución.

   **Calibración por resolución de impresión.** La página del Documento_Referencia mide 774 × 954 pt = 10,75 × 13,25 in. A 300 ppp son 3.225 × 3.975 px = 12,8 Mpx, lo que sin comprimir pesa **36,7 MiB** en RGB de 8 bits y **48,9 MiB** en CMYK de 8 bits; el mismo contenido en JPEG a 1:10 pesa 3,7 MiB. El nuevo valor por defecto, **64 MiB**, es la potencia de dos inmediatamente superior a una página completa CMYK sin comprimir a 300 ppp, así que admite cualquier formato comprimido y también el caso sin comprimir. Sigue siendo configurable en el struct de opciones del Resolvedor_Imagenes.

   **Consecuencia registrada, no acotada.** El base64 infla los bytes en 4/3, así que 64 MiB decodificados son unos 85 MiB de texto en el JSON de entrada. Y la cota es **por imagen**, no agregada: embeber las 7 imágenes del Documento_Referencia son 42,83 MiB de bytes, unos 57 MiB de base64 en un solo spread. Si el perfil de uso real exige un techo por documento, es una cota distinta y hoy no existe; debe decidirse con datos, no inventarse aquí. Esta sigue siendo la única cota de tamaño que sobrevive tras retirar el transporte HTTP: los límites de tiempo de descarga, tamaño de descarga, número de redirecciones, cuerpo de petición y tiempo de procesamiento ya no existen porque no hay red ni servidor.

4. **Subconjunto de `Resources/Preferences.xml` que emite el generador (Requisito 8, criterios 10 a 15) — subconjunto curado, pendiente de verificación manual.** El punto pendiente no es «qué hacer con 276 atributos»: el subconjunto candidato **ya existe en el repositorio**. La plantilla `pkg/idml/templates/minimal/Preferences.xml` (4.017 bytes) es lo que hoy embebe `NewFromTemplate()` y trae 10 elementos de primer nivel, con un `TextDefault` de **9 atributos**, no de 276. La brecha real son **tres elementos ausentes** —`DocumentPreference`, `MarginPreference` y `ViewPreference`— y sin `DocumentPreference` no puede fijarse el tamaño de página. Eso es todo lo que falta. El `README.md` de la propia plantilla enuncia el principio que cierra el riesgo: InDesign aplica sus propios valores por defecto a cualquier preferencia ausente, así que el subconjunto necesita ser válido, no completo.

   **Lo que sí queda pendiente es la verificación manual en Adobe InDesign, ahora adelantada.** Ningún test ni nota del repositorio registra que la salida de `NewFromTemplate()` se haya abierto en InDesign; el README de la plantilla dice que hay que verificarlo, pero no hay constancia de que se hiciera. Esa verificación es exactamente lo que hace el criterio 14 del Requisito 10, sobre la salida de `NewFromTemplate()` una vez completada con los tres elementos ausentes del subconjunto curado, y ocurre a profundidad 2 de la cadena de tareas en lugar de a profundidad 8. Las pruebas existentes (`pkg/idml/templates_test.go`, `pkg/idml/example_test.go`) solo comprueban que el paquete compila y parsea.

   **RESUELTO (Tarea 3).** El subconjunto curado, ya con los tres elementos que le faltaban —`DocumentPreference`, `MarginPreference` y `ViewPreference`— y con `DocumentPreference` emitido en `Resources/Preferences.xml` y no en `designmap.xml`, **fue aceptado por InDesign**: el documento abre sin daño y con el tamaño de página correcto. El principio del README de la plantilla queda confirmado: el subconjunto necesita ser válido, no completo. **No hubo que recurrir a la estrategia B.**

   **Estrategia B, reencuadrada: oráculo de diagnóstico, no plantilla.** Si InDesign rechaza el subconjunto curado, el `Resources/Preferences.xml` del Documento_Referencia (64.201 bytes, 146 nombres de elemento distintos; los mayores conteos de atributos son `TextDefault` 276, `PrintPreference` 91, `PrintBookletPrintPreference` 86, `EPubExportPreference` 67, `HTMLExportPreference` 39, `PageItemDefault` 36, `TextPreference` 34, `DocumentPreference` 25, `ViewPreference` 17) se usa para **comparar e identificar el atributo concreto que falta**, y se agrega solo ese al subconjunto curado. Nunca se copia el archivo completo: hacerlo hornearía la retícula, la configuración CJK y los ajustes de exportación de un periódico concreto en todos los documentos generados, lo que contradice el objetivo de servir a formatos arbitrarios y rompería el principio de «constructor completo desde cero» de la Introducción.

5. **Límite de 32 niveles de anidamiento de grupos (Requisito 4) — cota propuesta, con margen medido.** **Corrección de una medición previa de este documento:** se afirmaba que el spread del Documento_Referencia contenía «un solo `Group` de nivel 1». Es falso. Contiene **3** elementos `Group`: dos de nivel 1, `Self="u1724"` (con `Rectangle`, `TextFrame`, `TextFrame`, `GraphicLine` ×4 y un grupo anidado) y `Self="u1bc7"` (con `TextFrame` ×2), y uno de **nivel 2**, `Self="u175f"` dentro de `u1724`, con `Polygon` ×2. El anidamiento real existe y el Documento_Referencia ya lo ejercita, incluido un tipo, `Polygon`, que solo aparece dentro del grupo anidado.

   **Margen medido.** Recorriendo los 10 archivos exportados por InDesign disponibles en el repositorio —`plain.idml`, `example.idml`, `tripple.idml`, el Archivo_Evidencia_Imagenes, `testdata/Spread_u210.xml`, los 5 snippets `.idms` y el Documento_Referencia— la profundidad máxima de anidamiento de grupos es **2**, y solo el Documento_Referencia contiene grupos: los otros nueve tienen cero. El límite de 32 deja **16×** de margen sobre el documento más profundo observado.

   **Qué es y qué no es el límite.** No es una restricción del formato: se buscó un límite documentado de anidamiento de grupos en InDesign y no existe. Es un guardarraíl del parseo en un límite de confianza —el JSON de entrada puede venir de un origen que el operador no controla, y el parseo de grupos es recursivo—, así que el número solo necesita ser lo bastante alto para no rechazar documentos reales y lo bastante bajo para fallar con un error legible antes de agotar la pila. Si aparece un documento real que lo exceda, se revisa el número con el usuario antes de tratarlo como error de entrada.

6. **Derivación de `ColumnsPositions` (Requisito 8, criterio 16) — fórmula indeterminada.** En `MasterSpreads/MasterSpread_u2dd.xml` el `MarginPreference` de nivel de página declara `ColumnGutter="12"` y sin embargo emite `ColumnsPositions="0 59.4 73.39999999999999 132.79999999999998 146.79999999999998 206.2 ..."`: 20 números, los límites de 10 columnas. Aritmética verificada sobre esos valores: el último es `719,99998`, que equivale a 774 − 27 − 27, el ancho útil entre márgenes; cada columna mide 59,4 y cada hueco entre columnas mide **14**, no los 12 que declara `ColumnGutter`. La identidad se cumple exactamente: 10 × 59,4 + 9 × 14 = 720. La fórmula ingenua, por tanto, **no reproduce** lo que emitió InDesign. El origen del 14 es desconocido; el mismo archivo lleva además `LayoutGridDataInformation` y `CjkGridPreference`, que podrían intervenir.

   **No debe inventarse una fórmula.** La resolución es calcular con la fórmula directa (ancho útil, número de columnas y medianil declarado) y verificar en Adobe InDesign si acepta el valor emitido o lo recalcula al abrir el documento. Esa verificación sale casi gratis y temprano: `TemplateOptions` ya expone `ColumnCount` y `ColumnGutter`, así que la sonda del criterio 14 del Requisito 10 emite una variante de 1 columna y otra de 5 y registra qué hace InDesign con el valor. El punto importa porque `ColumnsPositions` es un valor derivado: si quien llama pide 5 columnas en una página A4, hay que calcularlo, no copiarlo.

   **RESUELTO (Tarea 3): la fórmula directa es correcta.** La sonda A4 de 5 columnas con medianil 12 emitió `ColumnsPositions="0 95.0552 107.0552 202.1104 214.1104 309.1656 321.1656 416.2208 428.2208 523.276"`, es decir columnas de 95,0552 y huecos de exactamente 12. Medido en InDesign tras abrir el documento: **12 pt**. InDesign **respeta el valor emitido y no lo recalcula**, así que el cálculo directo a partir del ancho útil, el número de columnas y el medianil es la derivación correcta y no hace falta ninguna fórmula adicional.

   **Y el 14 queda explicado por descarte:** si InDesign no impone un hueco distinto del declarado, el 14 de `MasterSpread_u2dd.xml` no es una regla del formato sino una particularidad de ese documento, con su `LayoutGridDataInformation` y su `CjkGridPreference` como candidatos. No se replica.

### Riesgos resueltos

- **La verificación manual en Adobe InDesign que compartían los riesgos 2, 4 y 6 — superada (Tarea 3).** Es el hito que la colocación temprana de la sonda buscaba, y llegó a profundidad 2 en lugar de 11. Las dos sondas A4 generadas por `NewFromTemplate()`, una de 1 columna y otra de 5, **se abren en Adobe InDesign sin aviso de daño y sin petición de recuperación**, y muestran la página A4 con su marco de texto dentro de los márgenes. Los tres riesgos quedan cerrados con un solo resultado: el conjunto de 13 entradas es suficiente (riesgo 2), el subconjunto curado de `Resources/Preferences.xml` es válido sin necesidad de copiar el del Documento_Referencia (riesgo 4), y la derivación directa de `ColumnsPositions` es la correcta porque InDesign respeta el valor emitido (riesgo 6).

  **Dato adicional no exigido:** los dos archivos abren también en **Affinity Publisher**, así que la forma emitida no depende de una tolerancia particular de InDesign. Es evidencia de que el paquete es estructuralmente correcto y no solo tolerado.

  **Lo que esto no cierra:** nada de fidelidad de parseo. El documento generado desde cero es abrible; que un IDML **existente** sobreviva el ciclo de lectura y escritura sigue midiéndose con el Arnes_Fidelidad, y esa cifra sigue lejos de cero.

- **Representación de la imagen embebida (Requisito 12) — resuelto empíricamente.** El formato se verificó inspeccionando el Archivo_Evidencia_Imagenes (exportado por InDesign), que contiene dos imágenes embebidas y una enlazada de control en un mismo spread. Los bytes viven en `Properties/Contents` del elemento `Image`, entre `Profile` y `GraphicBounds`, codificados en base64 estándar (RFC 4648) ajustado a 76 caracteres por línea; el `Link` se conserva con su `LinkResourceURI` y cambia solo `StoredState="Embedded"`; `LinkResourceSize` tiene la forma `0~<hex>` con el número de bytes decodificados. El elemento `Image` no gana ningún atributo respecto del modo enlazado: no lleva `ContentsEncoding`, que es exclusivo de `PastedSmoothShade` en `Resources/Graphic.xml` y corresponde a otra codificación (`Ascii64Encoding`), no relacionada. **Decisión sobre el archivo:** el Archivo_Evidencia_Imagenes queda **versionado** en `idmllib/testdata/archivo_evidencia_imagenes.idml`, junto al Documento_Referencia, de modo que la evidencia del formato viaja con el repositorio. **Regenerable solo en la máquina del usuario:** el documento fuente `imagenes_idml.indd` se conserva en la raíz del workspace pero fuera de git, así que cualquier variante futura (por ejemplo un TIFF embebido para comprobar si `Contents` cambia según el formato) se produce exportando de nuevo desde InDesign en esa máquina, no desde un clon del repositorio.
