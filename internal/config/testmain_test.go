package config

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Иначе первый GetConfig в пакете падает на validate (пустой BOT_TOKEN) и loadOnce навсегда возвращает ошибку.
	_ = os.Setenv("BOT_TOKEN", "000000000:dummy-token-for-go-test")
	os.Exit(m.Run())
}
