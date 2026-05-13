package gari

import (
	"fmt"
	"math"
	"reflect"
)

type (
	MapperFunc func(*Record)
)

func ColumnNameOf(fieldName string) string {
	return fieldName
}
func FieldNameOf(columnName string) string {
	return columnName
}

func CamelToSnakeMapper(ptrs any) MapperFunc {
	return func(r *Record) {
		// Check if argument ptr is a pointer type.
		ptrsV := reflect.ValueOf(ptrs)
		ptrsVKind := ptrsV.Kind()
		if ptrsVKind != reflect.Pointer {
			/*
				err := errs.Wrap(ErrReflectTypeNotPtr,
					errs.WithContext("type", ptrVKind))
				return nilV, err
			*/
			return
		}
		// Check if argument ptrs is a slice type.
		sliceV := ptrsV.Elem()
		if sliceV.Kind() != reflect.Slice {
			return
		}
		// Get type of item in slice.
		structType := sliceV.Type().Elem()
		fmt.Printf("Camel itemType=%v\n", structType)
		// Build a struct.
		if structType.Kind() != reflect.Struct {
			return
		}
		structPtrV := reflect.New(structType)
		structV := structPtrV.Elem()
		// Set fields in a struct.
		//count := structPtrV.NumField()
		for field, fieldV := range structV.Fields() {
			colName := UpperCamelToSnakeCase(field.Name)
			v := r.m[colName]
			switch tv := v.raw.(type) {
			case int64:
				if math.MinInt <= tv && tv <= math.MaxInt {
					fieldV.Set(reflect.ValueOf(int(tv)))
				}
			default:
				fieldV.Set(reflect.ValueOf(v.raw))
			}
			fmt.Printf("Struct type=%v, field=%s, value=%v, colName=%s, v=%v\n",
				structPtrV, field.Name, fieldV, colName, v)
		}
		// Add struct to slice.
		sliceV.Set(reflect.Append(sliceV, structV))
	}
}

func DefaultMapper(ptr any) MapperFunc {
	return func(r *Record) {
		m, err := structPtrToValuePtrs(ptr)
		if err != nil {
			return
		}
		for fieldName, valuePtr := range m {
			fmt.Printf("DefaultMapper fname=%s v=%v(%T)\n", fieldName, valuePtr, valuePtr)
		}
		/*
			m, err := structPtrToMap(ptr)
			if err != nil {
				return
			}
			for name, value :=
		*/
	}
}
