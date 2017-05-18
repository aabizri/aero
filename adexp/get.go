package adexp

// GetUnderlying returns the value behind a key.
//
// It is preferred to use GetPrimary / GetStructured / GetList instead
func (msg ADEXP) GetUnderlying(key string) (val interface{}, ok bool) {
	v, ok := msg[key]
	if !ok {
		return nil, false
	}
	return v.value, true
}

// GetKind returns the kind of the value associated with a key
func (msg ADEXP) GetKind(key string) (kind Kind, ok bool) {
	v, ok := msg[key]
	if !ok {
		return 0, false
	}

	return v.kind, true
}

// GetPrimary returns the primary field associated with the key.
//
// If the key isn't associated with a primary field, either because there is no such key or the value is of a different kind, ok returns as false.
func (msg ADEXP) GetPrimary(key string) (val string, ok bool) {
	v, ok := msg[key]
	if !ok {
		return "", false
	}

	if v.kind == Primary {
		pf, ok := v.value.(string)
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
func (msg ADEXP) GetStructured(key string) (val *Multi, ok bool) {
	v, ok := msg[key]
	if !ok {
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

// GetList returns the list field associated with the key.
//
// If the key isn't associated with a list field, either because there is no such key or the value is of a different kind, ok returns as false.
func (msg ADEXP) GetList(key string) (val *Multi, ok bool) {
	v, ok := msg[key]
	if !ok {
		return nil, false
	}

	if v.kind == List {
		lf, ok := v.value.(Multi)
		if !ok {
			panic("wildly unexpected wrong type")
		}
		return &lf, true
	}
	return nil, false
}
