package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("no record found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Model struct {
	SongModel     *SongModel
	PlaylistModel *PlaylistModel
}

func NewModel(db *sql.DB) Model {
	return Model{
		SongModel: &SongModel{
			DB: db,
		},
		PlaylistModel: &PlaylistModel{
			DB: db,
		},
	}
}
