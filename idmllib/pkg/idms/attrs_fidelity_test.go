package idms

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/beevik/etree"
)

// Este test guarda la fidelidad de atributos del ciclo de un snippet IDMS, que es lo
// que el comodín `,any,attr` vino a arreglar.
//
// Complementa a los fixtures dorados sin sustituirlos: un fixture dice «la salida es
// exactamente este texto», y hay que regenerarlo cada vez que la salida cambia por un
// buen motivo. Este test dice algo más fuerte y que no caduca: **la salida tiene
// exactamente los atributos de la entrada**, ni uno menos ni uno más. Los fixtures de
// este paquete llegaron a fijar dos veces una salida con atributos perdidos, y un test
// así lo habría delatado en el momento.

// attrsPorSelf indexa los atributos de cada elemento con `Self`, que es el
// identificador que permite emparejar entrada y salida sin depender del orden.
func attrsPorSelf(t *testing.T, data []byte) map[string]map[string]string {
	t.Helper()

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		t.Fatalf("no se pudo parsear el XML: %v", err)
	}

	out := map[string]map[string]string{}
	var walk func(*etree.Element)
	walk = func(e *etree.Element) {
		if self := e.SelectAttrValue("Self", ""); self != "" {
			attrs := make(map[string]string, len(e.Attr))
			for _, a := range e.Attr {
				attrs[a.Key] = a.Value
			}
			out[e.Tag+"/"+self] = attrs
		}
		for _, child := range e.ChildElements() {
			walk(child)
		}
	}
	walk(doc.Root())
	return out
}

func TestSnippetRoundtrip_ConservaTodosLosAtributos(t *testing.T) {
	snippets := []string{
		"../../testdata/Snippet_31F27A2D0.idms",
		"../../testdata/Snippet_31F27A387.idms",
		"../../testdata/Snippet_31F2CBC52.idms",
		"../../testdata/Snippet_31F2EABC4.idms",
		"../../testdata/Snippet_31F5B1AE5.idms",
	}

	for _, path := range snippets {
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("snippet ausente: %v (ruta esperada: %s)", err, path)
			}

			pkg, err := Read(path)
			if err != nil {
				t.Fatalf("Read falló: %v", err)
			}
			out, err := Marshal(pkg)
			if err != nil {
				t.Fatalf("Marshal falló: %v", err)
			}

			entrada := attrsPorSelf(t, src)
			salida := attrsPorSelf(t, out)

			if len(entrada) == 0 {
				t.Fatal("el snippet no tiene ningún elemento con Self: el test no comprobaría nada")
			}

			faltan, sobran, distintos := 0, 0, 0
			for clave, attrsEntrada := range entrada {
				attrsSalida, ok := salida[clave]
				if !ok {
					t.Errorf("el elemento %s no está en la salida", clave)
					continue
				}
				for name, want := range attrsEntrada {
					got, ok := attrsSalida[name]
					switch {
					case !ok:
						faltan++
						t.Errorf("%s: el atributo %q se perdió (valor %q)", clave, name, want)
					case got != want:
						distintos++
						t.Errorf("%s: el atributo %q vale %q y debería valer %q", clave, name, got, want)
					}
				}
				for name, got := range attrsSalida {
					if _, ok := attrsEntrada[name]; !ok {
						sobran++
						t.Errorf("%s: se emitió el atributo %q=%q, que no estaba en la entrada", clave, name, got)
					}
				}
			}

			t.Logf("%d elementos con Self comprobados: %d atributos perdidos, %d sobrantes, %d con otro valor",
				len(entrada), faltan, sobran, distintos)
		})
	}
}
