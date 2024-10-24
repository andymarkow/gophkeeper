package credrepo

import "time"

type Secret struct {
	ID        string
	Name      string
	UserID    string
	Metadata  string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
