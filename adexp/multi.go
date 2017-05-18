package adexp

import "github.com/aabizri/aero/adexp/parser"

// Multi is the structure behind structured & list fields
type Multi struct {
	m    map[string]value
	kind Kind
}

// GetUnderlying returns the value behind a key.
//
// It is preferred to use GetPrimary / GetStructured instead
func (mul *Multi) GetUnderlying(key string) (val interface{}, ok bool) {
	v, ok := mul.m[key]
	if !ok {
		return nil, false
	}
	return v.value, true
}

// GetKind returns the kind of the value associated with a key
func (mul *Multi) GetKind(key string) (kind Kind, ok bool) {
	v, ok := mul.m[key]
	if !ok {
		return 0, false
	}

	return v.kind, true
}

// GetPrimary returns the primary field associated with the key.
//
// If the key isn't associated with a primary field, either because there is no such key or the value is of a different kind, ok returns as false.
func (mul *Multi) GetPrimary(key string) (val string, ok bool) {
	v, ok := mul.m[key]
	if !ok {
		return "", false
	}

	if v.kind == Primary {
		pf, ok := v.value.(parser.PrimaryField)
		if !ok {
			panic("wildly unexpected wrong type")
		}
		return string(pf), true
	}
	return "", false
}

// GetStructured returns the structured field associated with the key.
//
// If the key isn't associated with a structured field, either because there is no such key or the value is of a different kind, ok returns as false.
func (mul *Multi) GetStructured(key string) (val *Multi, ok bool) {
	v, ok := mul.m[key]
	if !ok {
		return nil, false
	}

	// If the kind of the Multi is a structured key, and as there are no embedded structured in structured, we return nil, false
	if mul.kind == Structured {
		return nil, false
	}

	if v.kind == Structured {
		sf, ok := v.value.(Multi)
		if !ok {
			panic("wildly unexpected wrong type")
		}
		return &sf, true
	}
	return nil, false
}
