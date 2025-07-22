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
	SongModel     SongModel
	PlaylistModel PlaylistModel
	UserModel     UserModel
	TokenModel    TokenModel
}

func NewModel(db *sql.DB) Model {
	return Model{
		SongModel: SongModel{
			DB: db,
		},
		PlaylistModel: PlaylistModel{
			DB: db,
		},
		UserModel: UserModel{
			DB: db,
		},

		TokenModel: TokenModel{
			DB: db,
		},
	}
}
