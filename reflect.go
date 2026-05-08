package gari

import (
	"errors"
	"reflect"

	"github.com/goark/errs"
)

var (
	ErrReflectTypeNotPtr    = errors.New("ErrReflectTypeNotPtr")
	ErrReflectTypeNotStruct = errors.New("ErrReflectTypeNotStruct")
)

// Build map of values from struct pointer.
func structPtrToMap(ptr any) (map[string]any, error) {
	// Check if argument ptr is a pointer type.
	ptrV := reflect.ValueOf(ptr)
	ptrVKind := ptrV.Kind()
	if ptrVKind != reflect.Pointer {
		err := errs.Wrap(ErrReflectTypeNotPtr,
			errs.WithContext("type", ptrVKind))
		return nil, err
	}
	// Get struct value from pointer.
	structV := ptrV.Elem()
	structVKind := structV.Kind()
	if structVKind != reflect.Struct {
		err := errs.Wrap(ErrReflectTypeNotStruct,
			errs.WithContext("type", structVKind))
		return nil, err
	}
	// Store struct field values into map.
	fieldCount := structV.NumField()
	m := make(map[string]any, fieldCount)
	for structField, fieldV := range structV.Fields() {
		name := structField.Name
		value := fieldV.Interface()
		m[name] = value
	}
	return m, nil
}
