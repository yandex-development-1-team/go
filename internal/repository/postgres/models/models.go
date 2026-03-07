package models

import "database/sql"

type Setting struct {
	Category string         `db:"category"`
	Key      sql.NullString `db:"key"`
	Value    sql.NullString `db:"value"`
}
