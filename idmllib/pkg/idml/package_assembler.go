package idml

// Ensamblador de paquete: operaciones Add/Remove de alto nivel que mantienen coherente
// el designmap.xml (referencias y StoryList) y el índice de orden documental.

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// idPkgNS es el namespace que llevan las referencias del designmap.
const idPkgNS = "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"

// AddSpreadToPackage agrega un spread al paquete y lo registra en el designmap.
// El spread se serializa y se guarda con la ruta indicada. El designmap recibe
// la referencia <idPkg:Spread src="..."/> y el ChildOrder queda actualizado.
//
// Error si la ruta ya existe en el paquete.
func (p *Package) AddSpreadToPackage(path string, sp *spread.Spread) error {
	if p.hasFile(path) {
		return common.WrapErrorWithPath("idml", "add spread", path, common.ErrAlreadyExists)
	}

	data, err := spread.MarshalSpread(sp)
	if err != nil {
		return common.WrapErrorWithPath("idml", "add spread", path, err)
	}
	p.setFileData(path, data)

	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "add spread", path, err)
	}
	doc.Spreads = append(doc.Spreads, document.ResourceRef{
		XMLName: xml.Name{Space: idPkgNS, Local: "Spread"},
		Src:     path,
	})
	doc.RecordChild(document.ChildKindSpread)
	return nil
}

// AddMasterSpreadToPackage agrega un master spread al paquete y lo registra en el designmap.
// Error si la ruta ya existe.
func (p *Package) AddMasterSpreadToPackage(path string, ms *spread.MasterSpread) error {
	if p.hasFile(path) {
		return common.WrapErrorWithPath("idml", "add master spread", path, common.ErrAlreadyExists)
	}

	data, err := spread.MarshalMasterSpread(ms)
	if err != nil {
		return common.WrapErrorWithPath("idml", "add master spread", path, err)
	}
	p.setFileData(path, data)

	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "add master spread", path, err)
	}
	doc.MasterSpreads = append(doc.MasterSpreads, document.ResourceRef{
		XMLName: xml.Name{Space: idPkgNS, Local: "MasterSpread"},
		Src:     path,
	})
	doc.RecordChild(document.ChildKindMasterSpread)
	return nil
}

// AddStoryToPackage agrega una story al paquete, la registra en el designmap
// (referencia <idPkg:Story> y StoryList) y queda lista para usar.
//
// storyID es el Self de la story (ej. "u2f0") y se agrega al StoryList.
// Error si la ruta ya existe.
func (p *Package) AddStoryToPackage(path string, st *story.Story, storyID string) error {
	if p.hasFile(path) {
		return common.WrapErrorWithPath("idml", "add story to package", path, common.ErrAlreadyExists)
	}

	// Serializar y guardar el archivo.
	data, err := story.MarshalStory(st)
	if err != nil {
		return common.WrapErrorWithPath("idml", "add story to package", path, err)
	}
	p.setFileData(path, data)

	// Registrar en el designmap.
	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "add story to package", path, err)
	}
	doc.Stories = append(doc.Stories, document.ResourceRef{
		XMLName: xml.Name{Space: idPkgNS, Local: "Story"},
		Src:     path,
	})
	doc.RecordChild(document.ChildKindStory)

	// Agregar al StoryList.
	if doc.StoryList == "" {
		doc.StoryList = storyID
	} else {
		doc.StoryList = doc.StoryList + " " + storyID
	}
	return nil
}

// RemoveSpreadFromPackage elimina un spread del paquete y su referencia del designmap.
// Error si la ruta no existe.
func (p *Package) RemoveSpreadFromPackage(path string) error {
	if !p.hasFile(path) {
		return common.WrapErrorWithPath("idml", "remove spread", path, common.ErrNotFound)
	}

	p.invalidateCache(path)
	p.removeFile(path)

	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "remove spread", path, err)
	}

	for i, ref := range doc.Spreads {
		if ref.Src == path {
			doc.Spreads = append(doc.Spreads[:i], doc.Spreads[i+1:]...)
			doc.RemoveChildAt(document.ChildKindSpread, i)
			break
		}
	}
	return nil
}

// RemoveMasterSpreadFromPackage elimina un master spread del paquete y su referencia del designmap.
// Error si la ruta no existe.
func (p *Package) RemoveMasterSpreadFromPackage(path string) error {
	if !p.hasFile(path) {
		return common.WrapErrorWithPath("idml", "remove master spread", path, common.ErrNotFound)
	}

	p.invalidateCache(path)
	p.removeFile(path)

	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "remove master spread", path, err)
	}

	for i, ref := range doc.MasterSpreads {
		if ref.Src == path {
			doc.MasterSpreads = append(doc.MasterSpreads[:i], doc.MasterSpreads[i+1:]...)
			doc.RemoveChildAt(document.ChildKindMasterSpread, i)
			break
		}
	}
	return nil
}

// RemoveStoryFromPackage elimina una story del paquete, su referencia del designmap
// y su ID del StoryList.
// Error si la ruta no existe.
func (p *Package) RemoveStoryFromPackage(path string, storyID string) error {
	if !p.hasFile(path) {
		return common.WrapErrorWithPath("idml", "remove story from package", path, common.ErrNotFound)
	}

	p.invalidateCache(path)
	p.removeFile(path)

	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "remove story from package", path, err)
	}

	// Quitar de Stories.
	for i, ref := range doc.Stories {
		if ref.Src == path {
			doc.Stories = append(doc.Stories[:i], doc.Stories[i+1:]...)
			doc.RemoveChildAt(document.ChildKindStory, i)
			break
		}
	}

	// Quitar del StoryList.
	doc.StoryList = removeIDFromList(doc.StoryList, storyID)
	return nil
}

// SetBackingStory establece la backing story del paquete.
func (p *Package) SetBackingStory(path string, st *story.Story) error {
	data, err := story.MarshalStory(st)
	if err != nil {
		return common.WrapErrorWithPath("idml", "set backing story", path, err)
	}
	p.setFileData(path, data)

	doc, err := p.Document()
	if err != nil {
		return common.WrapErrorWithPath("idml", "set backing story", path, err)
	}

	if doc.BackingStory == nil {
		doc.BackingStory = &document.ResourceRef{
			XMLName: xml.Name{Space: idPkgNS, Local: "BackingStory"},
			Src:     path,
		}
		doc.RecordChild(document.ChildKindBackingStory)
	} else {
		doc.BackingStory.Src = path
	}
	return nil
}

// removeIDFromList quita un ID de una lista separada por espacios.
func removeIDFromList(list, id string) string {
	if list == id {
		return ""
	}
	// Intentar las tres posiciones: al inicio, en medio, al final.
	if len(list) > len(id)+1 && list[:len(id)+1] == id+" " {
		return list[len(id)+1:]
	}
	if idx := len(list) - len(id) - 1; idx >= 0 && list[idx:] == " "+id {
		return list[:idx]
	}
	// En medio
	old := " " + id + " "
	if i := indexOf(list, old); i >= 0 {
		return list[:i] + " " + list[i+len(old):]
	}
	return list
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
