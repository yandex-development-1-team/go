package models

import "database/sql"

type SettingRow struct {
	Category string         `db:"category"`
	Key      sql.NullString `db:"key"`
	Value    sql.NullString `db:"value"`
}

type Setting struct {
	Category string `db:"category"`
	Key      string `db:"key"`
	Value    string `db:"value"`
}
