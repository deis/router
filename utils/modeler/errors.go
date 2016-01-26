package modeler

import (
	"fmt"
	"reflect"
)

// NilLiteralModelError represents a failed attempt to populate a "model" because a nil literal was
// passed to the modeler instead of a struct pointer.
type NilLiteralModelError struct {
}

func newNilLiteralModelError() NilLiteralModelError {
	return NilLiteralModelError{}
}

func (e NilLiteralModelError) Error() string {
	return "Cannot populate a nil literal from the map."
}

// NonPointerModelError represents a failed attempt to populate a model because something other
// than a pointer was passed to the modeler.
type NonPointerModelError struct {
	tipe reflect.Type
}

func newNonPointerModelError(tipe reflect.Type) NonPointerModelError {
	return NonPointerModelError{tipe: tipe}
}

func (e NonPointerModelError) Error() string {
	return fmt.Sprintf("Cannot populate non-pointer type %s from the map.", e.tipe)
}

// NilModelError represents a failed attempt to populate a "model" because a nil pointer was
// passed to the modeler instead of a struct pointer.
type NilModelError struct {
	tipe reflect.Type
}

func newNilModelError(tipe reflect.Type) NilModelError {
	return NilModelError{tipe: tipe}
}

func (e NilModelError) Error() string {
	return fmt.Sprintf("Cannot populate nil %s from the map.", e.tipe)
}

// NonStructPointerModelError represents a failed attempt to populate a model because something
// other than a pointer was passed to the modeler.
type NonStructPointerModelError struct {
	tipe reflect.Type
}

func newNonStructPointerModelError(tipe reflect.Type) NonStructPointerModelError {
	return NonStructPointerModelError{tipe: tipe}
}

func (e NonStructPointerModelError) Error() string {
	return fmt.Sprintf("Cannot populate non-struct-pointer type %s from the map.", e.tipe)
}

// ModelValidationError represents an error resulting from a field having a value that doesn't
// satisfy a prescribed constraint.
type ModelValidationError struct {
	field      string
	constraint string
	value      string
}

func newModelValidationError(field string, constraint string, value string) ModelValidationError {
	return ModelValidationError{
		field:      field,
		constraint: constraint,
		value:      value,
	}
}

func (e ModelValidationError) Error() string {
	return fmt.Sprintf("Field \"%s\" value \"%s\" does not satisfy constraint /%s/", e.field, e.value, e.constraint)
}
