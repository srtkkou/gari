package gari

type SnakeCamelMapper struct{}

func (m SnakeCamelMapper) ColumnNameOf(fieldName string) string {
	return upperCamelToSnakeCase(fieldName)
}

func (m SnakeCamelMapper) FieldNameOf(columnName string) string {
	return snakeToUpperCamelCase(columnName)
}
