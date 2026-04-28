package gari

import (
	"database/sql"
	"time"
)

type (
	// Base type of model.
	Model struct {
		ID        int
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt sql.NullTime
	}
)
