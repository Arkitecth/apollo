package data

import (
	"database/sql"
	"time"

	"github.com/Arkitecth/apollo/validator"
)

type Playlist struct {
	ID         int64
	Created_At time.Time
	Name       string
	UserID     int64
}

type PlaylistModel struct {
	DB *sql.DB
}

func (m *PlaylistModel) ValidatePlaylist(playlist *Playlist) *validator.Validator {
	v := validator.New()
	v.Check(len(playlist.Name) > 30, "name", "Length cannot be greater than 30")
	return v
}

func (m *PlaylistModel) Insert(playlist *Playlist) error {
	query := `INSERT INTO songs (name, usedID) 
		  VALUES ($1, $2)
		  RETURNING id, created_at`
	args := []any{playlist.Name, playlist.UserID}

	err := m.DB.QueryRow(query, args...).Scan(&playlist.ID, &playlist.Created_At)
	if err != nil {
		return nil
	}
	return nil
}
