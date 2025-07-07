package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Arkitecth/apollo/validator"
)

type Song struct {
	ID         int64
	Created_At time.Time
	Name       string
	SongURL    string
	Artist     string
	Thumbnail  string
	Version    int
}

type SongModel struct {
	DB *sql.DB
}

func ValidateSong(v *validator.Validator, song *Song) {
	v.Check(song.Name == "", "name", "name cannot be blank")
	v.Check(len(song.Name) > 30, "name", "name cannot be greater than 30 bytes")

	v.Check(song.Artist == "", "artist", "artist cannot be blank")
	v.Check(len(song.Artist) > 50, "artist", "artist name cannot be greater than 50 bytes")

	v.Check(len(song.SongURL) > 100, "url", "song url cannot be greater than 100")
	v.Check(len(song.Thumbnail) > 100, "thumbnail", "thumnbail cannot be greater than 100")
}

func (m *SongModel) Insert(song *Song) error {
	query := `INSERT INTO songs (name, artist, song_url, thumbnail) 
		  VALUES ($1, $2, $3, $4)
		  RETURNING id, created_at, version`

	args := []any{song.Name, song.Artist, song.SongURL, song.Thumbnail}

	err := m.DB.QueryRow(query, args...).Scan(&song.ID, &song.Created_At, &song.Version)
	if err != nil {
		return nil
	}
	return nil
}

func (m *SongModel) Get(id int64) (*Song, error) {
	query := `SELECT id, created_at, artist, song_url, thumbnail, version FROM songs
		  WHERE id = $1 `

	song := &Song{}

	err := m.DB.QueryRow(query, id).Scan(
		&song.ID,
		&song.Created_At,
		&song.Artist,
		&song.Thumbnail,
		&song.SongURL,
		&song.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return song, nil
}

func (m *SongModel) Update(song *Song) error {
	query := `UPDATE songs SET artist = $1, thumbnail = $2, song_url = $3, version = version + 1
	WHERE id = $4
	RETURNING version`

	args := []any{
		song.Artist,
		song.Thumbnail,
		song.SongURL,
		song.ID,
	}

	return m.DB.QueryRow(query, args...).Scan(&song.Version)

}

func (m *SongModel) Delete(id int64) error {
	query := `DELETE FROM songs
		  WHERE id = $1 `

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
