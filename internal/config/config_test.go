package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("YANDEX_FORMS_WEBHOOK_TOKEN", "test-webhook-token")

	configYAML := filepath.Join("..", "..", "config", "config.yaml")
	if _, err := os.Stat(configYAML); err != nil {
		t.Skip("нет локального config/config.yaml:", err)
	}

	paths := []string{"../../config"}

	cfg, err := GetConfig(paths)
	if err != nil {
		t.Fatalf("GetConfig error: %v", err)
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
