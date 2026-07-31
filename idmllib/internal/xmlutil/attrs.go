package xmlutil

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// El problema que resuelven estas dos funciones: encoding/xml asigna a los campos
// del struct los atributos que el struct declara y **descarta el resto en silencio**.
// Un `Rectangle` del Documento_Referencia trae 33 atributos y el modelo declara 21,
// así que 12 se pierden en cada ciclo de lectura y escritura. Medido sobre el corpus
// completo son casi 14.000 atributos perdidos.
//
// La solución es un campo comodín, `OtherAttrs []xml.Attr`, donde se guarda lo que no
// corresponde a ningún campo declarado. El patrón ya existía escrito a mano en
// `CharacterStyleRange` de pkg/story: un `switch` con un caso por atributo conocido y
// un `default` que acumula el resto. Funciona, pero repetirlo en los doce tipos que
// lo necesitan son unas 150 ramas para leer y otras 150 para emitir, y cada una es
// una ocasión de que un atributo desaparezca por una errata.
//
// Estas funciones hacen ese reparto una sola vez, leyendo del propio struct qué
// campos declara mediante reflexión.

// OtherAttrsField es el nombre por convención del campo comodín en el struct destino.
// También se reconoce cualquier campo `[]xml.Attr` etiquetado `,any,attr`, sea cual sea
// su nombre, que es la forma que entiende encoding/xml.
//
// Un struct sin comodín se acepta: sus atributos no declarados se descartan, que es el
// comportamiento anterior, pero UnmarshalAttrs no falla por ello. La decisión es
// deliberada: hay structs que no necesitan el comodín, y obligar a declararlo
// convertiría esta función en un problema para ellos.
const OtherAttrsField = "OtherAttrs"

// attrSliceType es el tipo que debe tener el campo comodín.
var attrSliceType = reflect.TypeOf([]xml.Attr{})

// wildcardField localiza el campo comodín de un struct: por su etiqueta `,any,attr`, o
// por el nombre de la convención. Devuelve un valor inválido si no hay ninguno.
func wildcardField(v reflect.Value) reflect.Value {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type != attrSliceType {
			continue
		}
		tag := field.Tag.Get("xml")
		if strings.Contains(tag, ",any,attr") || field.Name == OtherAttrsField {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

// UnmarshalAttrs reparte los atributos de un elemento entre los campos declarados del
// struct destino y su campo comodín OtherAttrs.
//
// dest tiene que ser un puntero a struct. Conserva de cada atributo no declarado su
// prefijo de namespace, su nombre local, su valor literal y su orden de aparición.
//
// Los campos de structs embebidos cuentan como declarados: `Rectangle` embebe
// `PageItemBase`, que aporta `Self`, `Name`, `ItemLayer`, `Visible`,
// `GeometricBounds` e `ItemTransform`. Sin recorrer los embebidos, esos seis
// atributos irían al comodín y acabarían emitidos dos veces.
func UnmarshalAttrs(attrs []xml.Attr, dest any) error {
	v, err := structValue(dest)
	if err != nil {
		return err
	}

	fields := attrFields(v.Type())

	// El comodín se vacía antes de rellenarlo. Sin esto, reutilizar un struct o
	// decodificar dos veces sobre el mismo valor acumularía los atributos, que es el
	// fallo que detecta la prueba de idempotencia.
	other := wildcardField(v)
	hasOther := other.IsValid() && other.CanSet()
	if hasOther {
		other.Set(reflect.Zero(other.Type()))
	}

	for _, attr := range attrs {
		info, declared := fields[attrKey(attr.Name)]
		if !declared {
			// Una declaración de namespace (xmlns:idPkg) no es un atributo de datos,
			// pero tampoco se puede descartar: se guarda en el comodín como cualquier
			// otro atributo no declarado, y así el elemento se re-emite igual.
			if hasOther {
				other.Set(reflect.Append(other, reflect.ValueOf(attr)))
			}
			continue
		}

		if err := setAttrField(v.FieldByIndex(info.index), attr); err != nil {
			return err
		}
	}

	return nil
}

// MarshalAttrs construye la lista de atributos de un elemento: primero los campos
// declarados que tienen valor, en el orden en que están declarados, y después los
// atributos del comodín.
//
// Un atributo del comodín cuyo nombre coincida con un campo declarado **no se emite**,
// para no duplicar el nombre. Gana el campo declarado, porque es el que un llamador
// puede haber modificado.
//
// src puede ser un struct o un puntero a struct.
func MarshalAttrs(src any) ([]xml.Attr, error) {
	v, err := structValue(src)
	if err != nil {
		return nil, err
	}

	fields := attrFields(v.Type())

	// Se recorren los campos por orden de índice para que la salida sea estable y
	// siga el orden de declaración del struct.
	ordered := make([]attrField, 0, len(fields))
	for key, info := range fields {
		ordered = append(ordered, attrField{key: key, info: info})
	}
	sortAttrFields(ordered)

	attrs := make([]xml.Attr, 0, len(ordered))
	emitted := make(map[string]bool, len(ordered))

	for _, f := range ordered {
		field := v.FieldByIndex(f.info.index)
		if f.info.omitEmpty && field.IsZero() {
			continue
		}
		value, err := attrFieldValue(field, f.key)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, xml.Attr{Name: nameFromKey(f.key), Value: value})
		emitted[f.key] = true
	}

	other := wildcardField(v)
	if other.IsValid() {
		for i := 0; i < other.Len(); i++ {
			attr, ok := other.Index(i).Interface().(xml.Attr)
			if !ok {
				continue
			}
			if emitted[attrKey(attr.Name)] {
				continue
			}
			attrs = append(attrs, attr)
			emitted[attrKey(attr.Name)] = true
		}
	}

	return attrs, nil
}

// attrFieldInfo describe el campo que declara un atributo: dónde está y si el tag
// pide omitirlo cuando está vacío.
type attrFieldInfo struct {
	index     []int
	omitEmpty bool
}

// attrField es una entrada de la lista ordenada que usa MarshalAttrs.
type attrField struct {
	key  string
	info attrFieldInfo
}

// sortAttrFields ordena por la ruta de índices del campo, de modo que los atributos
// se emitan en el orden en que están declarados en el struct, y los de un struct
// embebido en la posición donde está el embebido.
func sortAttrFields(fields []attrField) {
	sort.Slice(fields, func(i, j int) bool {
		a, b := fields[i].info.index, fields[j].info.index
		for k := 0; k < len(a) && k < len(b); k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return len(a) < len(b)
	})
}

// structValue resuelve un struct o puntero a struct a su valor, con errores que
// nombran lo que se recibió.
func structValue(x any) (reflect.Value, error) {
	if x == nil {
		return reflect.Value{}, fmt.Errorf("xmlutil: el destino de los atributos es nil")
	}

	v := reflect.ValueOf(x)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, fmt.Errorf("xmlutil: el destino de los atributos es un puntero nil a %s", v.Type().Elem())
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("xmlutil: el destino de los atributos debe ser un struct o un puntero a struct, se recibió %s", v.Kind())
	}
	return v, nil
}

// attrFields devuelve, por clave de atributo, la ruta de índices del campo que lo
// declara. Recorre los structs embebidos, cuyos campos cuentan como propios.
func attrFields(t reflect.Type) map[string]attrFieldInfo {
	fields := make(map[string]attrFieldInfo)
	collectAttrFields(t, nil, fields)
	return fields
}

func collectAttrFields(t reflect.Type, prefix []int, fields map[string]attrFieldInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		index := append(append([]int(nil), prefix...), i)

		// Struct embebido sin nombre: sus campos son campos de este struct.
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			collectAttrFields(field.Type, index, fields)
			continue
		}

		if field.PkgPath != "" { // campo sin exportar
			continue
		}

		key, omitEmpty, ok := attrKeyFromTag(field)
		if !ok {
			continue
		}
		// El primero gana: un campo del struct externo tiene prioridad sobre el
		// homónimo de un embebido, igual que en Go.
		if _, exists := fields[key]; !exists {
			fields[key] = attrFieldInfo{index: index, omitEmpty: omitEmpty}
		}
	}
}

// attrKeyFromTag extrae la clave del atributo que declara un campo y si lleva
// omitempty, o false si el campo no es un atributo.
func attrKeyFromTag(field reflect.StructField) (key string, omitEmpty, ok bool) {
	tag := field.Tag.Get("xml")
	if tag == "-" {
		return "", false, false
	}

	parts := strings.Split(tag, ",")
	isAttr, isWildcard := false, false
	for _, opt := range parts[1:] {
		switch opt {
		case "attr":
			isAttr = true
		case "any":
			isWildcard = true
		case "omitempty":
			omitEmpty = true
		}
	}
	if !isAttr {
		return "", false, false
	}
	// `,any,attr` es el comodín de encoding/xml, no un atributo con nombre. Tratarlo
	// como declarado hacía que MarshalAttrs intentara emitir el slice entero como si
	// fuera un atributo llamado «OtherAttrs».
	if isWildcard {
		return "", false, false
	}

	name := parts[0]
	if name == "" {
		name = field.Name
	}
	// La forma "espacio nombre" del tag de encoding/xml, y la forma "prefijo:nombre"
	// que usan las declaraciones de namespace.
	if space, local, found := strings.Cut(name, " "); found {
		return space + ":" + local, omitEmpty, true
	}
	return name, omitEmpty, true
}

// attrKey normaliza un xml.Name a la clave con la que se comparan atributos y campos.
func attrKey(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	return name.Space + ":" + name.Local
}

func nameFromKey(key string) xml.Name {
	if space, local, found := strings.Cut(key, ":"); found {
		return xml.Name{Space: space, Local: local}
	}
	return xml.Name{Local: key}
}

// setAttrField asigna el valor de un atributo al campo que lo declara, convirtiendo
// según el tipo del campo.
//
// El error nombra el atributo, el valor y el tipo esperado, que es lo que pide el
// Req 2 criterio 9, y deja el campo sin asignar.
func setAttrField(field reflect.Value, attr xml.Attr) error {
	if !field.CanSet() {
		return nil
	}

	convErr := func() error {
		return fmt.Errorf("xmlutil: el atributo %q con valor %q no es convertible al tipo %s del campo que lo declara",
			attrKey(attr.Name), attr.Value, field.Type())
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(attr.Value)

	case reflect.Bool:
		b, err := strconv.ParseBool(attr.Value)
		if err != nil {
			return convErr()
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(attr.Value, 10, field.Type().Bits())
		if err != nil {
			return convErr()
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(attr.Value, 10, field.Type().Bits())
		if err != nil {
			return convErr()
		}
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(attr.Value, field.Type().Bits())
		if err != nil {
			return convErr()
		}
		field.SetFloat(f)

	default:
		return convErr()
	}

	return nil
}

// attrFieldValue devuelve el valor a emitir de un campo declarado.
//
// El descarte por `omitempty` lo decide MarshalAttrs antes de llamar aquí, con la
// misma regla que encoding/xml: se omite si el campo está en su valor cero.
func attrFieldValue(field reflect.Value, key string) (string, error) {
	switch field.Kind() {
	case reflect.String:
		return field.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(field.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, field.Type().Bits()), nil
	default:
		return "", fmt.Errorf("xmlutil: el campo del atributo %q tiene el tipo %s, que no se puede emitir como atributo", key, field.Type())
	}
}
