package gari

type Mapper interface {
	ColumnNameOf(fieldName string) string
	FieldNameOf(columnName string) string
}
