package gari

import (
	"database/sql"
	"time"
)

type (
	Model struct {
		ID        int
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt sql.NullTime
	}
)
