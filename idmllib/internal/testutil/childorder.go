package testutil

import "testing"

// AssertFieldOrderCovers comprueba que el orden de campos de un tipo con
// xmlutil.ChildOrder nombra exactamente las clases de hijo que ese tipo puede
// producir.
//
// Por qué hace falta: ChildOrder.Replay emite en dos pasadas, primero el orden
// registrado al parsear y después los hijos que el registro no menciona, y esa
// segunda pasada recorre el orden de campos. Una clase ausente de esa lista solo se
// emitiría cuando el registro la mencione, así que desaparecería al serializar un
// modelo construido desde cero. Es un fallo silencioso, y esta comprobación es lo
// que lo convierte en un test rojo.
func AssertFieldOrderCovers(t *testing.T, fieldOrder, kinds []string) {
	t.Helper()

	inOrder := make(map[string]bool, len(fieldOrder))
	for _, kind := range fieldOrder {
		if inOrder[kind] {
			t.Errorf("la clase %q está repetida en el orden de campos", kind)
		}
		inOrder[kind] = true
	}

	for _, kind := range kinds {
		if !inOrder[kind] {
			t.Errorf("la clase %q no está en el orden de campos: se perdería al serializar un modelo construido desde cero", kind)
		}
	}

	if len(fieldOrder) != len(kinds) {
		t.Errorf("el orden de campos tiene %d clases y el modelo declara %d: sobra o falta alguna", len(fieldOrder), len(kinds))
	}
}
