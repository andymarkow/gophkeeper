package filerepo

import "time"

type Secret struct {
	ID        string
	Name      string
	UserID    string
	Metadata  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
	Salt      string
	IV        string
	FileName  string
	Location  string
	Checksum  string
	Size      int64
}
