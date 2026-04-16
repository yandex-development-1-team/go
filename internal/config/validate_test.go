package config

import "testing"

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("api_only allows empty bot token", func(t *testing.T) {
		t.Parallel()
		err := validateConfig(&Config{
			APIOnly:     true,
			DB:          DatabaseConfig{PostgresURL: "postgres://u:p@h/db?sslmode=disable"},
			Storage:     StorageConfig{Endpoint: "http://localhost:9000", Bucket: "test-bucket"},
			YandexForms: YandexFormsConfig{WebhookToken: "test-token"},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not api_only requires bot token", func(t *testing.T) {
		t.Parallel()
		err := validateConfig(&Config{
			APIOnly: false,
			DB:      DatabaseConfig{PostgresURL: "postgres://u:p@h/db?sslmode=disable"},
		})
		if err == nil {
			t.Fatal("expected error when bot token is empty")
		}
	})

	t.Run("telegram nested bot_token satisfies validation", func(t *testing.T) {
		t.Parallel()
		err := validateConfig(&Config{
			APIOnly: false,
			Telegram: Telegram{
				BotToken: "123:abc",
			},
			DB:          DatabaseConfig{PostgresURL: "postgres://u:p@h/db?sslmode=disable"},
			Storage:     StorageConfig{Endpoint: "http://localhost:9000", Bucket: "test-bucket"},
			YandexForms: YandexFormsConfig{WebhookToken: "test-token"},
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
