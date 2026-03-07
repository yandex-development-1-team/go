package config

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	paths := []string{"../../config"}

	cfg, err := GetConfig(paths)
	if err != nil {
		t.Skipf("config not found (run from repo root with config): %v", err)
		return
	}

	v := reflect.ValueOf(cfg)
	typ := reflect.TypeOf(cfg)

	t.Log("Loaded config:")
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		value := v.Field(i)

		fieldName := field.Name
		fieldValue := value.Interface()

		isEmpty := false
		switch value.Kind() {
		case reflect.String:
			isEmpty = value.String() == ""
		case reflect.Int:
			isEmpty = value.Int() == 0
		}

		status := "✓"
		if isEmpty {
			status = "✗"
		}

		t.Logf("  %s [%s]: %v (%s)",
			fieldName,
			field.Tag.Get("yaml"),
			fieldValue,
			status)
	}

}
