package gari

// Column type
type kind = string

const (
	kindInvalid = kind("Invalid")
	kindString  = kind("String")
	kindTime    = kind("Time")
	kindInt64   = kind("Int64")
	kindFloat64 = kind("Float64")
	kindBlob    = kind("Blob")
)
