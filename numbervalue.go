package jsbuiltin

import (
	"encoding/json"
	"errors"
	"strconv"
)

// NumberValue is the number JSON.parse makes of a JSON number: a float64 as
// it is, or a json.Number (a decoder run with UseNumber, or jsorder.Decode)
// read as a float64. A literal past float64's range is ±Infinity (or ±0 for
// an underflow), as JSON.parse returns it, not a refusal; ok is false only
// for something that is not a number at all.
//
// Moved from gotruto's jsonx (by way of truto-jsonata-go's copy), 2026-10-31.
func NumberValue(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case json.Number:
		f, err := strconv.ParseFloat(string(t), 64)
		if err != nil && !errors.Is(err, strconv.ErrRange) {
			return 0, false
		}
		return f, true
	}
	return 0, false
}
