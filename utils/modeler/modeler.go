package modeler

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Modeler is a utility for populating an arbitrary model's fields with values from a
// map[string]string.
type Modeler struct {
	prefix                string
	fieldTag              string
	constraintTag         string
	warnOnValidationError bool
}

// NewModeler returns a pointer to a new Modeler, confgiured with the provided map key prefix and
// struct tag key.
func NewModeler(prefix string, fieldTag string, constraintTag string, warnOnValidationError bool) *Modeler {
	return &Modeler{
		prefix:                prefix,
		fieldTag:              fieldTag,
		constraintTag:         constraintTag,
		warnOnValidationError: warnOnValidationError,
	}
}

// MapToModel populates the provided model with values from the provided map.
func (m *Modeler) MapToModel(data map[string]string, out interface{}) error {
	rv := reflect.ValueOf(out)
	return m.mapToModel(data, "", rv)
}

func (m *Modeler) mapToModel(data map[string]string, context string, rv reflect.Value) error {
	// If rv is invalid (represents a nil literal), we cannot proceed.
	if rv.Kind() == reflect.Invalid {
		return newNilLiteralModelError()
	}
	// We also cannot proceed if rv is not a pointer.
	if rv.Kind() != reflect.Ptr {
		return newNonPointerModelError(rv.Type())
	}
	// Or if it is a pointer to a nil.
	if rv.IsNil() {
		return newNilModelError(rv.Type())
	}
	// At this point, we know we're working with a pointer that points to something.
	// Get what it points to...
	elem := rv.Elem()
	// If the thing it points to isn't a struct, that's also a problem.
	if elem.Kind() != reflect.Struct {
		return newNonStructPointerModelError(rv.Type())
	}
	rt := elem.Type()
	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		fieldTagValue := rf.Tag.Get(m.fieldTag)
		// If no tag value is found...
		if fieldTagValue == "" {
			// Just move to the next field.
			continue
		}
		if rf.Type.Kind() == reflect.Ptr || rf.Type.Kind() == reflect.Struct {
			// We're nested... use some recursion...
			var nestedContext string
			if context == "" {
				nestedContext = fieldTagValue
			} else {
				nestedContext = fmt.Sprintf("%s.%s", context, fieldTagValue)
			}
			err := m.mapToModel(data, nestedContext, elem.Field(i))
			if err != nil {
				return err
			}
		} else {
			// We're not nested!
			var key string
			prefix := m.prefix
			if prefix != "" && !strings.HasSuffix(prefix, "/") {
				prefix = fmt.Sprintf("%s/", prefix)
			}
			if context == "" {
				key = fmt.Sprintf("%s%s", prefix, fieldTagValue)
			} else {
				key = fmt.Sprintf("%s%s.%s", prefix, context, fieldTagValue)
			}
			stringVal, ok := data[key]
			if ok {
				constraintTagValue := rf.Tag.Get(m.constraintTag)
				if constraintTagValue != "" {
					constraint := regexp.MustCompile(constraintTagValue)
					if !constraint.MatchString(stringVal) {
						err := newModelValidationError(key, constraintTagValue, stringVal)
						if m.warnOnValidationError {
							log.Printf("WARNING: %s -- skipping this field and using default value \"%v\".", err, elem.Field(i))
							continue
						} else {
							return err
						}
					}
				}
				if rf.Type.Kind() == reflect.String {
					elem.Field(i).Set(reflect.ValueOf(stringVal))
				} else if rf.Type.Kind() == reflect.Int {
					intVal, err := strconv.Atoi(stringVal)
					if err != nil {
						return err
					}
					elem.Field(i).Set(reflect.ValueOf(intVal))
				} else if rf.Type.Kind() == reflect.Bool {
					boolVal, err := strconv.ParseBool(stringVal)
					if err != nil {
						return err
					}
					elem.Field(i).Set(reflect.ValueOf(boolVal))
				} else if rf.Type.Kind() == reflect.Slice {
					sliceVal := strings.Split(stringVal, ",")
					for j, token := range sliceVal {
						sliceVal[j] = strings.TrimSpace(token)
					}
					elem.Field(i).Set(reflect.ValueOf(sliceVal))
				} else {
					return fmt.Errorf("Unsupported type %s.", rf.Type.Kind())
				}
			}
		}
	}
	return nil
}
