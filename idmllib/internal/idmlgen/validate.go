package idmlgen

import "fmt"

// Validate comprueba que el input es coherente antes de intentar generar.
// Retorna un error descriptivo si algo no es válido.
func Validate(input *DocumentInput) error {
	doc := input.Document
	if doc.WidthMm <= 0 {
		return fmt.Errorf("document.widthMm: debe ser > 0, se recibió %v", doc.WidthMm)
	}
	if doc.HeightMm <= 0 {
		return fmt.Errorf("document.heightMm: debe ser > 0, se recibió %v", doc.HeightMm)
	}
	if len(input.Pages) == 0 {
		return fmt.Errorf("pages: debe tener al menos una página")
	}
	for i, page := range input.Pages {
		for j, frame := range page.Frames {
			if frame.Type == "" {
				return fmt.Errorf("pages[%d].frames[%d].type: es obligatorio", i, j)
			}
			if frame.Type != "text" {
				return fmt.Errorf("pages[%d].frames[%d].type: solo se soporta \"text\", se recibió %q", i, j, frame.Type)
			}
			b := frame.Bounds
			if b.BottomMm <= b.TopMm {
				return fmt.Errorf("pages[%d].frames[%d].bounds: bottomMm (%v) debe ser > topMm (%v)", i, j, b.BottomMm, b.TopMm)
			}
			if b.RightMm <= b.LeftMm {
				return fmt.Errorf("pages[%d].frames[%d].bounds: rightMm (%v) debe ser > leftMm (%v)", i, j, b.RightMm, b.LeftMm)
			}
		}
	}
	return nil
}
