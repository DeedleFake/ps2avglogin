package main

import (
	"reflect"
)

// walkStruct walks through a struct and all sub-structs recustively,
// calling f on all 'leaf' fields. The first argument to f is the name
// of the field in the form parent.field, and the second argument is
// the field itself. If any call to f returns non-nil error,
// walkStruct stops and returns that error.
func walkStruct(s interface{}, f func(string, reflect.Value) error) error {
	var inner func(v reflect.Value, p string) error
	inner = func(v reflect.Value, p string) error {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			name := t.Field(i).Name
			if (name[0] < 'A') || (name[0] > 'Z') {
				continue
			}

			var err error
			switch field := reflect.Indirect(v.Field(i)); field.Kind() {
			case reflect.Struct:
				err = inner(field, p+name+".")
			default:
				err = f(p+name, field)
			}
			if err != nil {
				return err
			}
		}

		return nil
	}

	return inner(reflect.Indirect(reflect.ValueOf(s)), "")
}
