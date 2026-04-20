package gari

import (
	"time"
)

// テスト用日時2011/12/13 14:15:16.789Z
func time20111213() time.Time {
	return time.Date(
		2011, time.December, 13, 14, 15, 16, 789000000, time.UTC,
	)
}
