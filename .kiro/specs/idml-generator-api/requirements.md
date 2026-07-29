# Documento de Requisitos

## Introducción

`idmllib` es hoy un *round-tripper*: abre un paquete IDML existente, parsea un subconjunto de sus archivos, permite modificarlos y los reescribe copiando byte a byte todo lo que no entendió. Esta funcionalidad convierte la librería en un **generador**: producir un archivo IDML válido y abrible por Adobe InDesign a partir de una entrada JSON recibida por una API HTTP, sin partir de ningún paquete previo.

El alcance funcional se define por un archivo de referencia real exportado por InDesign, descomprimido en `idml_con_imagenes_texto_y_divisiones/` (página de periódico: 1 spread, 3 master spreads, 30 stories, texto multicolumna, imágenes enlazadas, grupos anidados, líneas gráficas, capas, secciones, guías, control de cambios). Todo lo que ese archivo modela debe poder generarse; todo lo que ese archivo contiene debe poder parsearse y re-serializarse sin pérdida estructural.

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
- **Resolvedor_Imagenes**: el componente que dada una ruta local o una URL obtiene los bytes de la imagen, sus dimensiones en píxeles y su resolución.
- **API_HTTP**: el servidor `net/http` que expone el Constructor_Documento sobre HTTP.
- **Validador_Entrada**: el componente de la API_HTTP que valida el JSON recibido antes de invocar al Constructor_Documento.
- **Documento_Referencia**: el IDML exportado por InDesign descomprimido en `idml_con_imagenes_texto_y_divisiones/`.
- **SSRF**: *Server-Side Request Forgery*; abuso de una descarga iniciada por el servidor para alcanzar destinos de red no previstos.

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
7. IF el recorrido del directorio del Documento_Referencia selecciona cero archivos, THEN THE Arnes_Fidelidad SHALL terminar con resultado de fallo indicando la ruta recorrida.
8. WHERE un archivo seleccionado corresponde a un tipo de archivo que el Paquete_IDML copia sin parsear, THE Arnes_Fidelidad SHALL contarlo en la categoría «copiados sin parsear», separada del número de archivos sin diferencias, y SHALL mantener el resultado del arnés sin fallo por causa de ese archivo.

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
2. WHEN se parsea el `SpreadElement` de `Spreads/Spread_uce7.xml`, cuya secuencia de elementos de página es `Rectangle`, `Rectangle`, `Group`, `TextFrame`, `GraphicLine`, THE Modelo_Spread SHALL almacenar esos elementos en esa misma secuencia y SHALL conservar el número total de elementos de página de la entrada.
3. WHEN se serializa un `SpreadElement`, THE Serializador_XML SHALL emitir los elementos de página en el mismo Orden_Documental en que fueron almacenados, sin reagruparlos por tipo de elemento.
4. THE Modelo_Spread SHALL exponer accesores por tipo (marcos de texto, rectángulos, imágenes, óvalos, polígonos, líneas gráficas, grupos) que devuelvan en Orden_Documental los elementos de ese tipo contenidos en el nivel inmediato del `SpreadElement`, sin descender al contenido de los grupos, y que devuelvan una colección vacía de longitud 0 cuando ese nivel no contiene ningún elemento de ese tipo.
5. THE Modelo_Story SHALL almacenar los hijos de `StoryElement`, incluidos `ParagraphStyleRange`, los elementos de control de cambios y los elementos que el Modelo_Story no modela, en un contenedor único que conserve el Orden_Documental y la posición de cada hijo respecto de los demás, siguiendo el mismo patrón ya implementado en `CharacterStyleRange` (`Children []CharacterChild` con `MarshalXML`/`UnmarshalXML` propios en `pkg/story/parse.go`).
6. WHEN el contenedor ordenado reemplaza los slices por tipo, THE Modelo_Spread SHALL mantener compilables los consumidores existentes `pkg/idml/selection.go`, `pkg/idml/package_modifications.go`, `pkg/analysis` y `pkg/idms`, y SHALL mantener en verde las pruebas existentes de esos cuatro consumidores sin modificar sus aserciones.
7. WHEN el Arnes_Fidelidad procesa el spread, los 3 master spreads y las 30 stories del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias de categoría «orden de elementos distinto».
8. IF un elemento hijo de un `SpreadElement` o de un `StoryElement` no corresponde a ningún tipo modelado, THEN THE Serializador_XML SHALL almacenarlo como XML crudo dentro del mismo contenedor ordenado del elemento contenedor, conservando su posición respecto de los elementos modelados.
9. WHEN se agrega un elemento de página a un `SpreadElement`, THE Modelo_Spread SHALL colocarlo en la última posición del contenedor ordenado.
10. WHEN se elimina un elemento de página de un `SpreadElement`, THE Modelo_Spread SHALL conservar el Orden_Documental relativo de los elementos restantes y SHALL reducir en 1 el número total de elementos de página.

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
3. THE Modelo_Spread SHALL almacenar el contenido de página de `MasterSpread` en el mismo contenedor ordenado que el Requisito 3 define para `SpreadElement`, admitiendo los mismos tipos de contenido y exponiendo los mismos accesores por tipo.
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

1. THE Paquete_IDML SHALL modelar en `Resources/Preferences.xml` los cuatro elementos que el generador necesita emitir, `DocumentPreference`, `MarginPreference`, `ViewPreference` y `TextDefault`, declarando en `DocumentPreference` los atributos `PageHeight`, `PageWidth`, `PagesPerDocument` y `FacingPages`, en `MarginPreference` los atributos `Top`, `Bottom`, `Left`, `Right`, `ColumnCount`, `ColumnGutter`, `ColumnDirection` y `ColumnsPositions`, y en `ViewPreference` los atributos `HorizontalMeasurementUnits`, `VerticalMeasurementUnits` y `RulerOrigin`.
2. THE Paquete_IDML SHALL emitir el elemento `DocumentPreference` como hijo del elemento raíz de `Resources/Preferences.xml` y SHALL emitir `designmap.xml` sin ningún elemento `DocumentPreference`, escribiendo en `PageHeight`, `PageWidth`, `PagesPerDocument` y `FacingPages` los valores recibidos del Constructor_Documento, cuyos valores en el Documento_Referencia son `954`, `774`, `1` y `true`.
3. WHEN `NewFromTemplate()` de `pkg/idml/templates.go` construye un paquete, THE Paquete_IDML SHALL producir un `designmap.xml` que contiene cero elementos `DocumentPreference` y un `Resources/Preferences.xml` que contiene exactamente un elemento `DocumentPreference`.
4. THE Paquete_IDML SHALL modelar `XML/Tags.xml` con elementos `XMLTag` de atributos `Self` y `Name` y con su hijo `Properties/TagColor`, cuyo atributo `type` tiene el valor `enumeration` y cuyo contenido de texto es el nombre del color.
5. THE Paquete_IDML SHALL modelar `XML/BackingStory.xml` con el elemento `XmlStory` de atributos `Self`, `UserText`, `IsEndnoteStory`, `AppliedTOCStyle`, `TrackChanges`, `StoryTitle` y `AppliedNamedGrid`, y con el elemento recursivo `XMLElement` de atributos `Self` y `MarkupTag`.
6. WHEN el Arnes_Fidelidad procesa `Resources/Preferences.xml`, `XML/Tags.xml` y `XML/BackingStory.xml` del Documento_Referencia, THE Serializador_XML SHALL producir cero diferencias en las categorías definidas en el Requisito 1, criterio 2.
7. THE Paquete_IDML SHALL preservar mediante OtherAttrs según el Requisito 2 los 276 atributos que el Documento_Referencia declara en `TextDefault`, sin declarar ninguno de ellos como campo del struct.
8. IF un `Resources/Preferences.xml` de entrada contiene cero elementos `DocumentPreference`, THEN THE Paquete_IDML SHALL devolver un error que identifique la ruta del archivo y el elemento ausente, sin devolver un modelo parcial.
9. IF el Constructor_Documento recibe un ancho de página o un alto de página menor o igual que 0, THEN THE Constructor_Documento SHALL devolver un error que identifique la dimensión y el valor recibido, sin generar archivo.

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

---

### Requisito 11: Resolución de imágenes desde ruta local y URL

**User Story:** Como cliente de la API, quiero indicar imágenes por ruta local o por URL, para no tener que preparar los archivos manualmente antes de generar el documento.

#### Acceptance Criteria

1. THE Resolvedor_Imagenes SHALL aceptar como origen de imagen una ruta de archivo local o una URL.
2. THE Resolvedor_Imagenes SHALL obtener el ancho y alto en píxeles y el formato de la imagen usando `image.DecodeConfig` de la biblioteca estándar.
3. THE Resolvedor_Imagenes SHALL calcular `ActualPpi`, `EffectivePpi` y el `ItemTransform` de escala del elemento `Image` a partir de las dimensiones obtenidas y del tamaño del marco de destino.
4. WHERE el origen es una URL, THE Resolvedor_Imagenes SHALL aceptar únicamente los esquemas `http` y `https`.
5. WHERE el origen es una URL, THE Resolvedor_Imagenes SHALL aplicar un tiempo límite de descarga con valor por defecto de 10 segundos y un tamaño máximo de descarga con valor por defecto de 20 MiB, ambos configurables mediante el struct de opciones del Resolvedor_Imagenes, y SHALL abortar la descarga al alcanzar cualquiera de los dos límites.
6. WHERE el origen es una URL, THE Resolvedor_Imagenes SHALL verificar la dirección IP a la que la conexión se establece efectivamente, en la función de control del marcador de red (`net.Dialer.Control` o equivalente) y no sobre el resultado previo de la resolución DNS, y SHALL rechazar la conexión cuando esa dirección pertenece a los rangos privados `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` y `fc00::/7`, a los rangos de bucle local `127.0.0.0/8` y `::1/128`, a los rangos de enlace local `169.254.0.0/16` y `fe80::/10`, a las direcciones no especificadas `0.0.0.0` y `::`, o a la dirección de metadatos de infraestructura `169.254.169.254`.
7. WHEN una URL responde con una redirección, THE Resolvedor_Imagenes SHALL aplicar a la URL de destino las mismas verificaciones de esquema del criterio 4 y de dirección de destino del criterio 6.
8. WHERE el origen es una ruta local, THE Resolvedor_Imagenes SHALL restringir la lectura a un directorio base configurado y THE Resolvedor_Imagenes SHALL rechazar las rutas que, tras resolver enlaces simbólicos y segmentos relativos, queden fuera de ese directorio base.
9. IF la imagen no puede obtenerse, no puede decodificarse o excede un límite, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique el origen y el motivo, sin incluir en el mensaje la respuesta cruda del servidor remoto.
10. THE Resolvedor_Imagenes SHALL admitir el modo de imagen enlazada, generando el elemento `Image` con sus hijos `Link`, `FrameFittingOption` y `ClippingPathSettings`, y copiando el archivo resuelto a una carpeta de recursos junto al documento generado.
11. WHEN el número de redirecciones seguidas para un origen alcanza el límite configurado, con valor por defecto de 3, THE Resolvedor_Imagenes SHALL abortar la descarga y devolver un error que identifique el origen y el número de redirecciones seguidas.
12. IF el origen recibido es una cadena vacía o declara un esquema distinto de `http` y de `https`, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique el origen y el esquema recibido, sin abrir ninguna conexión de red y sin leer ningún archivo.
13. IF el origen es una ruta local y el directorio base del criterio 8 no está configurado, THEN THE Resolvedor_Imagenes SHALL devolver un error que identifique la opción de configuración ausente, sin leer ningún archivo.

---

### Requisito 12: Imágenes embebidas

**User Story:** Como cliente de la API, quiero que las imágenes viajen dentro del archivo IDML por defecto, para poder mover el documento sin arrastrar una carpeta de enlaces.

#### Acceptance Criteria

1. THE Resolvedor_Imagenes SHALL determinar la representación de una imagen embebida en IDML a partir de un archivo IDML exportado por Adobe InDesign que contenga al menos una imagen de mapa de bits embebida, sin derivarla de ninguna otra fuente.
2. WHERE Adobe InDesign representa imágenes de mapa de bits embebidas en IDML, THE Resolvedor_Imagenes SHALL emitir la imagen embebida como modo por defecto usando la representación registrada según el criterio 6.
3. WHERE Adobe InDesign no representa imágenes de mapa de bits embebidas en IDML, THE Resolvedor_Imagenes SHALL usar el modo de imagen enlazada del Requisito 11 como comportamiento por defecto y THE API_HTTP SHALL entregar el archivo IDML junto con la carpeta de recursos enlazados.
4. THE Resolvedor_Imagenes SHALL permitir seleccionar explícitamente el modo enlazado o el modo embebido por cada imagen de la entrada.
5. WHEN el modo embebido está disponible, THE Resolvedor_Imagenes SHALL generar un documento que Adobe InDesign abre mostrando la imagen sin solicitar la relocalización del enlace (verificación manual: requiere abrir el archivo en Adobe InDesign y registrar el resultado).
6. WHEN el análisis empírico del criterio 1 termina, THE Resolvedor_Imagenes SHALL quedar documentado en el documento de diseño de este spec con los nombres de los elementos y de los atributos que Adobe InDesign emite para la imagen embebida, o con la constancia escrita de que no emite ninguno.
7. IF la entrada solicita explícitamente el modo embebido y el análisis del criterio 1 registró que Adobe InDesign no representa imágenes de mapa de bits embebidas en IDML, THEN THE API_HTTP SHALL rechazar la petición con un error que identifique la imagen y el modo solicitado, sin generar archivo.
8. WHILE la decisión registrada según el criterio 6 no ha sido confirmada por el usuario, THE Resolvedor_Imagenes SHALL mantener el modo de imagen enlazada del Requisito 11 como único modo implementado.

---

### Requisito 13: API HTTP de generación

**User Story:** Como cliente externo, quiero enviar la descripción de un documento en JSON por HTTP y recibir el archivo IDML, para integrar la generación en mi flujo sin usar Go.

#### Acceptance Criteria

1. THE API_HTTP SHALL exponer un extremo `POST /documents` que reciba la descripción del documento en JSON y devuelva el archivo IDML generado.
2. THE API_HTTP SHALL implementarse con el paquete `net/http` de la biblioteca estándar, sin marco web adicional.
3. THE API_HTTP SHALL representar en el JSON de entrada las mismas capacidades que expone el Constructor_Documento según el Requisito 10.
4. THE Validador_Entrada SHALL validar el JSON recibido antes de invocar al Constructor_Documento, verificando el tamaño de página, el número de páginas, el número de elementos por página, la longitud del texto por elemento y el esquema de las URL de imagen.
5. THE Validador_Entrada SHALL limitar el tamaño del cuerpo de la petición a 1 MiB por defecto, mediante `http.MaxBytesReader`, con el límite configurable en el struct de opciones de la API_HTTP.
6. IF la entrada incumple una regla de validación, THEN THE API_HTTP SHALL responder con estado `400` y un cuerpo que identifique el campo y la regla incumplida.
7. IF la generación del documento falla, THEN THE API_HTTP SHALL responder con estado `500` y un identificador de error, y THE API_HTTP SHALL registrar el detalle del fallo del lado del servidor.
8. THE API_HTTP SHALL construir sus errores con los tipos de `pkg/common/errors.go`.
9. THE API_HTTP SHALL responder con estado `405` a los métodos distintos de `POST` sobre `/documents`.
10. WHEN el procesamiento de una petición alcanza el tiempo límite configurado, con valor por defecto de 30 segundos, THE API_HTTP SHALL cancelar el trabajo en curso mediante el contexto de la petición y responder con estado `503`.
11. IF el cuerpo de la petición excede el límite del criterio 5, THEN THE API_HTTP SHALL responder con estado `413` sin invocar al Constructor_Documento.
12. IF el cuerpo de la petición no es un documento JSON válido, THEN THE API_HTTP SHALL responder con estado `400` y un cuerpo que identifique la posición del error de sintaxis.
13. THE API_HTTP SHALL exponer `POST /documents` sin autenticación ni autorización en esta versión, y THE API_HTTP SHALL declarar esa ausencia en la documentación del extremo como decisión pendiente registrada en la sección «Riesgos y decisiones pendientes» de este documento.

---

### Requisito 14: Cierre de fidelidad y límites declarados

**User Story:** Como mantenedor de la librería, quiero que la prueba de fidelidad de roundtrip esté activa y que el alcance no cubierto esté escrito, para no asumir una cobertura que no existe.

#### Acceptance Criteria

1. THE Arnes_Fidelidad SHALL ejecutar `TestGoldenRoundtrip_ExampleIDML` de `pkg/idml/golden_test.go` sin la llamada `t.Skip` actual, sustituyendo la comparación byte a byte por Comparación_Estructural archivo por archivo.
2. THE Arnes_Fidelidad SHALL comparar byte a byte únicamente los archivos que el Paquete_IDML copia sin parsear.
3. THE Arnes_Fidelidad SHALL verificar que el archivo generado contiene `mimetype` como primera entrada del ZIP y almacenada sin compresión.
4. THE Paquete_IDML SHALL enumerar en el comentario de paquete de `pkg/idml` las construcciones declaradas fuera de alcance de esta funcionalidad: tablas (`Table`, `Cell`), notas al pie (`Footnote`), hipervínculos (`Hyperlink`), referencias cruzadas (`CrossReferenceSource`, `CrossReferenceFormat`) y tipografía CJK (`KinsokuTable`, `MojikumiTable`).
5. WHEN el JSON recibido por la API_HTTP declara una construcción enumerada en el criterio 4, THE Validador_Entrada SHALL rechazar la petición con un error que identifique la construcción y el campo del JSON que la declara, sin invocar al Constructor_Documento.
6. WHEN el Paquete_IDML abre un archivo IDML existente que contiene una construcción enumerada en el criterio 4, THE Paquete_IDML SHALL preservarla mediante los mecanismos de elementos y atributos no modelados de los Requisitos 2 y 3 y SHALL emitirla al reescribir el paquete con las mismas categorías del Requisito 1, criterio 2, en cero diferencias.
7. THE Validador_Entrada SHALL aplicar el rechazo del criterio 5 únicamente al JSON recibido por la API_HTTP, y THE Paquete_IDML SHALL aplicar la preservación del criterio 6 únicamente a los archivos leídos de un paquete IDML existente, sin que el rechazo del criterio 5 impida abrir, modificar y reescribir esos archivos.

---

### Requisito 15: Restricciones técnicas del proyecto

**User Story:** Como mantenedor de la librería, quiero que la funcionalidad se construya con lo que ya hay en el repositorio, para no aumentar la carga de mantenimiento.

#### Acceptance Criteria

1. THE Paquete_IDML SHALL implementar esta funcionalidad usando la biblioteca estándar de Go y las dependencias ya declaradas en `go.mod` (`github.com/beevik/etree`, `goldie/v2`, `gopter`, `go-cmp`), sin agregar dependencias nuevas.
2. THE Paquete_IDML SHALL declarar la versión de Go `1.23` en `go.mod` y en la clave `go` de `.golangci.yml`, reemplazando el `1.21.13` que hoy declara `go.mod`, porque `1.23` es la versión con la que ya se ejecutan el linter y la integración continua.
3. THE Paquete_IDML SHALL seguir las convenciones de `ARCHITECTURE.md` para etiquetas XML, elementos comodín `OtherElements`, nombres de tipos y campos, y manejo de errores con `pkg/common/errors.go`.
4. THE Paquete_IDML SHALL exponer la configuración de los componentes nuevos mediante structs de opciones, sin introducir interfaces no solicitadas.
5. WHEN un incremento de esta funcionalidad agrega o modifica lógica en un paquete, THE Paquete_IDML SHALL incluir en ese mismo paquete al menos una prueba automatizada que falle si esa lógica se revierte.
6. THE Paquete_IDML SHALL escribir en español los comentarios de paquete, los mensajes de error y los documentos de este spec.
7. WHEN se completa un incremento de esta funcionalidad, THE Paquete_IDML SHALL terminar sin fallos la ejecución de `go test ./...` y la ejecución de `golangci-lint run`.

---

## Riesgos y decisiones pendientes

Estos puntos no están resueltos y deben confirmarse antes de implementar los requisitos que dependen de ellos.

1. **Representación de la imagen embebida (Requisito 12) — incógnita técnica.** No hay ningún ejemplo de imagen de mapa de bits embebida en el repositorio: el único `<Contents>` binario es un `PastedSmoothShade` con `ContentsEncoding="Ascii64Encoding"` en `Resources/Graphic.xml`, y el Documento_Referencia usa exclusivamente enlaces `file:/Users/...`. La búsqueda de documentación no arrojó una especificación utilizable. El formato **no debe inventarse**: el Requisito 12.1 exige obtener empíricamente un IDML exportado por InDesign con una imagen embebida. El resultado se registra por escrito según el Requisito 12, criterio 6. Si el resultado es que IDML no soporta embebido de raster, aplica la ruta alternativa del Requisito 12.3 (IDML + carpeta de enlaces), y hasta que el usuario confirme esa decisión el único modo implementado sigue siendo el enlazado (Requisito 12, criterio 8).

2. **Autenticación del extremo HTTP (Requisito 13) — decisión pendiente.** Tal como está especificado, `POST /documents` queda **sin autenticación ni autorización**, y descarga URLs arbitrarias por cuenta de quien llama. Las mitigaciones del Requisito 11 (esquemas permitidos, límites de tamaño y tiempo, bloqueo de destinos internos) reducen el riesgo de SSRF pero no sustituyen un control de acceso. El Requisito 13, criterio 13, deja esa ausencia escrita en lugar de inventar un esquema. Debe definirse un esquema de autenticación y una política de límite de peticiones, o bien declararse explícitamente que el servicio solo se despliega en una red de confianza, antes de exponer el extremo fuera de un entorno local.

3. **Dependencia de orden entre tareas.** Los Requisitos 2 y 3 son prerequisitos de los Requisitos 4 a 14. Implementar el Constructor_Documento antes de resolverlos produciría documentos que pierden atributos y orden de apilamiento, y obligaría a rehacer el trabajo.

4. **Conjunto mínimo de archivos que exige InDesign (Requisito 10, criterio 10) — inferencia.** La lista de entradas del criterio 10 se infiere del Documento_Referencia y de los 9 archivos que hoy emite `NewFromTemplate()`, no de documentación de Adobe. Es un criterio automatizable que sustituye parcialmente la verificación manual del criterio 8; si al abrir el archivo en InDesign aparece un archivo obligatorio adicional, la lista debe ampliarse.

5. **Valores por defecto de los límites (Requisitos 11 y 13) — cotas propuestas.** Los valores 10 s de descarga, 20 MiB de tamaño de imagen, 3 redirecciones, 1 MiB de cuerpo de petición y 30 s de procesamiento son cotas elegidas para que los criterios sean verificables, no mediciones. Todas quedan configurables; los valores definitivos deben confirmarse con el perfil de uso real.

6. **`TextDefault` con 276 atributos (Requisito 8, criterio 7) — se preserva, no se declara.** El Documento_Referencia declara 276 atributos en `TextDefault`; ninguno se modela como campo. Eso resuelve el ciclo de lectura y escritura, pero no define qué subconjunto debe emitir el generador al construir `Resources/Preferences.xml` desde cero. Ese subconjunto debe acordarse con el usuario antes de implementar el criterio 1.

7. **Límite de 32 niveles de anidamiento de grupos (Requisito 4) — cota propuesta.** El valor 32 es una cota testeable elegida para que el límite sea verificable y para acotar la recursión del parseo, no una medición del Documento_Referencia, cuyo único spread contiene un solo `Group` de nivel 1. Si aparece un documento real que exceda esa profundidad, el número debe revisarse con el usuario antes de tratarlo como un error de entrada.
