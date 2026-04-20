package gari

// Column type
type kind = string

const (
	kindNull   = kind("Null")
	kindBool   = kind("Bool")
	kindString = kind("String")
	kindTime   = kind("Time")
	kindInt64  = kind("Int64")
)
