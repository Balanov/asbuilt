package profile

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Parse reads a profile, refusing unknown keys instead of dropping expectations.
func Parse(data []byte) (*Profile, error) {
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if err := knownFields(data, reflect.TypeOf(p), ""); err != nil {
		return nil, err
	}
	return &p, nil
}

func knownFields(data []byte, typ reflect.Type, prefix string) error {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	known := make(map[string]reflect.Type)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		known[name] = field.Type
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fieldType, ok := known[key]
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if !ok {
			return fmt.Errorf("unknown field %q", path)
		}
		if err := knownFields(fields[key], fieldType, path); err != nil {
			return err
		}
	}
	return nil
}
