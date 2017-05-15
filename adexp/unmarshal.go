package adexp

import (
	"encoding"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

var fields map[string]bool

const (
	hyphen    = "-"
	separator = " "
)

func unmarshalToField(structField reflect.Value, value string) error {
	// Assign to it
	if !structField.CanInterface() {
		return errors.New("cannot use interface of field")
	}

	// If it supports TextUnmarshaller, use it!
	if t, ok := structField.Interface().(encoding.TextUnmarshaler); ok {
		err := t.UnmarshalText([]byte(value))
		if err != nil {
			return err
		}
	} else if structField.Kind() == reflect.String && structField.CanSet() { // If it is a string and we can set to it
		structField.SetString(value)
	} else {
		return errors.New("can't set field")
	}

	return nil
}

func (msg *ADEXP) unmarshalExpression(expression string) error {
	// Split based on the first space
	splitted := strings.SplitN(expression, separator, 2)
	if len(splitted) != 2 {
		return errors.New("unexpected error while splitting")
	}

	// Check if the first part isn't a keyword
	fieldName := splitted[0]
	fieldValue := splitted[1]
	if !fields[fieldName] {
		return errors.New("unknown field")
	}

	// Now use reflect to retrieve the corresponding field
	structField := reflect.ValueOf(msg).FieldByName(fieldName)
	if !structField.IsValid() {
		return errors.New("no such field found")
	}

	// Assign to that field
	err := unmarshalToField(structField, fieldValue)
	if err != nil {
		return errors.Wrapf(err, "error while unmarshalling to field %s", fieldName)
	}

	return nil
}

func (msg *ADEXP) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), hyphen)
	for _, expression := range parts {
		err := msg.unmarshalExpression(expression)
		if err != nil {
			return errors.Wrapf(err, "UnmarshalText: error while unmarshalling expression \"%s\"", expression)
		}
	}
	return nil
}
