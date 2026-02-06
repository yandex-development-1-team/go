package database

import (
	"context"
	"github.com/yandex-development-1-team/go/internal/database/db_models"
)

type DB struct {
	//todo добавить logger
	//todo добавить клиента db
}

func NewDB() *DB {
	return &DB{}
}

func (db DB) GetBoxes(ctx context.Context) ([]db_models.Box, error) {
	return []db_models.Box{}, nil
}

func (db DB) GetBoxSolutions() []db_models.Box {
	var boxes []db_models.Box

	return boxes
}
