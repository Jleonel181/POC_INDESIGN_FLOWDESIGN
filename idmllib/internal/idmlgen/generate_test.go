package idmlgen

import (
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/document"
	idmlpkg "github.com/dimelords/idmllib/v2/pkg/idml"
)

// masterSpreadXML genera el XML mínimo de un MasterSpread con el Self y AppliedMaster
// dados, parseable por spread.ParseMasterSpread.
func masterSpreadXML(self, appliedMaster string) []byte {
	// AppliedMaster vive en OtherAttrs del inner MasterSpread, no como campo tipado.
	am := ""
	if appliedMaster != "" {
		am = ` AppliedMaster="` + appliedMaster + `"`
	}
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<idPkg:MasterSpread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">` + "\n" +
		`<MasterSpread Self="` + self + `" Name="Test" NamePrefix="T" BaseName="Master"` + am +
		` PageCount="1" ShowMasterItems="true" ItemTransform="1 0 0 1 0 0" OverriddenPageItemProps="" />` + "\n" +
		`</idPkg:MasterSpread>`)
}

func TestCollectMasterChain_DetectsCycle(t *testing.T) {
	// Crear un paquete base para inyectar los master spreads sintéticos.
	pkg, err := idmlpkg.NewFromTemplate(nil)
	if err != nil {
		t.Fatalf("NewFromTemplate: %v", err)
	}

	// Dos masters que se referencian mutuamente: mA → mB → mA (ciclo).
	pathA := "MasterSpreads/MasterSpread_mA.xml"
	pathB := "MasterSpreads/MasterSpread_mB.xml"

	pkg.SetFileData(pathA, masterSpreadXML("mA", "mB"))
	pkg.SetFileData(pathB, masterSpreadXML("mB", "mA"))

	// Construir un Document con ambos masters registrados en el designmap.
	doc := &document.Document{
		MasterSpreads: []document.ResourceRef{
			{Src: pathA},
			{Src: pathB},
		},
	}

	// collectMasterChain debe detectar el ciclo y retornar error con "ciclo".
	_, err = collectMasterChain(pkg, doc, pathA)
	if err == nil {
		t.Fatal("se esperaba error por ciclo, pero collectMasterChain retornó nil")
	}
	if !strings.Contains(err.Error(), "ciclo") {
		t.Errorf("el error debería contener 'ciclo', pero fue: %s", err.Error())
	}
}
