package postgres

import (
	"time"

	"github.com/lib/pq"
)

type userRow struct {
	ID           int64          `db:"id"`
	TelegramNick *string        `db:"username"`
	FirstName    *string        `db:"first_name"`
	LastName     *string        `db:"last_name"`
	Email        string         `db:"email"`
	UserPass     string         `db:"password_hash"` // проверить
	Role         string         `db:"role"`
	Status       string         `db:"status"`
	Permissions  pq.StringArray `db:"permissions"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}
