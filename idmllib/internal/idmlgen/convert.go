package idmlgen

import (
	"strconv"
	"strings"
)

const ptPerMm = 72.0 / 25.4

func mmToPt(mm float64) float64 { return mm * ptPerMm }

// num formatea un float con hasta 6 decimales, eliminando ceros innecesarios.
func num(v float64) string {
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}
