package gari

import (
	"errors"
	"reflect"

	"github.com/goark/errs"
)

var (
	ErrReflectNotPointer = errors.New("ErrReflectNotPointer")
	ErrReflectNotStruct  = errors.New("ErrReflectNotStruct")
	ErrReflectNotSlice   = errors.New("ErrReflectNotSlice")
	ErrReflectNilSlice   = errors.New("ErrReflectNilSlice")
)

// Build map of values from struct pointer.
func structPtrToMap(ptr any) (map[string]any, error) {
	structV, err := structPtrToV(ptr)
	if err != nil {
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

/*
	func structPtrToValuePtrs(ptr any) (map[string]any, error) {
		structV, err := structPtrToV(ptr)
		if err != nil {
			return nil, err
		}
		// Store struct field value pointers into map.
		fieldCount := structV.NumField()
		m := make(map[string]any, fieldCount)
		for structField, fieldV := range structV.Fields() {
			name := structField.Name
			valuePtr := fieldV.Addr().Interface()
			m[name] = valuePtr
		}
		return m, nil
	}
*/
func structPtrToV(ptr any) (reflect.Value, error) {
	nilV := reflect.ValueOf(nil)
	// Check if argument ptr is a pointer type.
	ptrV := reflect.ValueOf(ptr)
	ptrVKind := ptrV.Kind()
	if ptrVKind != reflect.Pointer {
		err := errs.Wrap(ErrReflectNotPointer,
			errs.WithContext("type", ptrVKind))
		return nilV, err
	}
	// Get struct value from pointer.
	structV := ptrV.Elem()
	structVKind := structV.Kind()
	if structVKind != reflect.Struct {
		err := errs.Wrap(ErrReflectNotStruct,
			errs.WithContext("type", structVKind))
		return nilV, err
	}
	return structV, nil
}

func makeSlice(ptr any) error {
	// Check if argument ptr is a pointer type.
	ptrV := reflect.ValueOf(ptr)
	ptrVKind := ptrV.Kind()
	if ptrVKind != reflect.Pointer {
		return errs.Wrap(ErrReflectNotPointer,
			errs.WithContext("kind", ptrVKind))
	}
	// Check if argument is a pointer to slice.
	sliceV := ptrV.Elem()
	sliceVKind := sliceV.Kind()
	if sliceVKind != reflect.Slice {
		return errs.Wrap(ErrReflectNotSlice,
			errs.WithContext("kind", sliceVKind))
	}
	// Do nothing if slice is not nil.
	if !sliceV.IsNil() {
		return nil
	}
	// Make slice.
	newV := reflect.MakeSlice(sliceV.Type(), 0, 0)
	ptrV.Set(newV)
	return nil
}

func newStruct(ptr any) (any, error) {
	// Check if argument ptr is a pointer type.
	ptrV := reflect.ValueOf(ptr)
	ptrVKind := ptrV.Kind()
	if ptrVKind != reflect.Pointer {
		err := errs.Wrap(ErrReflectNotPointer,
			errs.WithContext("kind", ptrVKind))
		return nil, err
	}
	// Check if argument is a pointer to slice.
	sliceV := ptrV.Elem()
	sliceVKind := sliceV.Kind()
	if sliceVKind != reflect.Slice {
		err := errs.Wrap(ErrReflectNotSlice,
			errs.WithContext("kind", sliceVKind))
		return nil, err
	}
	// Get slice element type.
	var structType reflect.Type
	elementVType := sliceV.Type().Elem()
	elementVKind := elementVType.Kind()
	switch elementVKind {
	case reflect.Struct:
		structType = elementVType
	case reflect.Pointer:
		realElementVType := elementVType.Elem()
		realElementVKind := realElementVType.Kind()
		if realElementVKind == reflect.Struct {
			structType = realElementVType
		}
	}
	if structType == nil {
		err := errs.Wrap(ErrReflectNotStruct,
			errs.WithContext("type", elementVType))
		return nil, err
	}
	// Build new struct.
	newV := reflect.New(structType)
	return newV.Interface(), nil
}

func eachField(ptr any, fn func(name string, ptr any)) error {
	structV, err := structPtrToV(ptr)
	if err != nil {
		return err
	}
	// Iterate over struct fields.
	for structField, fieldV := range structV.Fields() {
		name := structField.Name
		valuePtr := fieldV.Addr()
		fn(name, valuePtr.Interface())
	}
	return nil
}

func appendToSlice(slicePtr any, itemPtr any) error {
	// Check if argument ptr is a pointer type.
	slicePtrV := reflect.ValueOf(slicePtr)
	ptrVKind := slicePtrV.Kind()
	if ptrVKind != reflect.Pointer {
		return errs.Wrap(ErrReflectNotPointer,
			errs.WithContext("kind", ptrVKind))
	}
	// Check if argument is a pointer to slice.
	sliceV := slicePtrV.Elem()
	sliceVKind := sliceV.Kind()
	if sliceVKind != reflect.Slice {
		return errs.Wrap(ErrReflectNotSlice,
			errs.WithContext("kind", sliceVKind))
	}
	// Do nothing if slice is not nil.
	if sliceV.IsNil() {
		return ErrReflectNilSlice
	}
	// Append item to slice.
	itemPtrV := reflect.ValueOf(itemPtr)
	itemV := itemPtrV.Elem()
	newSliceV := reflect.Append(sliceV, itemV)
	sliceV.Set(newSliceV)
	return nil
}
