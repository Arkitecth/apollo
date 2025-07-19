package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&song.ID, &song.Created_At, &song.Version)
}

func (m *SongModel) Get(id int64) (*Song, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id, created_at, artist, song_url, thumbnail, version FROM songs
		  WHERE id = $1 `

	song := &Song{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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

func (m *SongModel) GetAll(artist string, name string, filters Filters) ([]*Song, error) {

	query := fmt.Sprintf(`
	SELECT id, created_at, artist, name, song_url, thumbnail, version 
	FROM songs 
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{artist, name, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	songs := []*Song{}

	for rows.Next() {
		var song Song

		err := rows.Scan(
			&song.ID,
			&song.Created_At,
			&song.Name,
			&song.Artist,
			&song.Version,
		)

		if err != nil {
			return nil, err
		}

		songs = append(songs, &song)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return songs, nil
}
func (m *SongModel) GetAllSongs(playlistID int64) ([]*Song, error) {
	query := `SELECT songs.id, songs.created_at, songs.name, songs.song_url, songs.artist, songs.thumbnail, version 
		  FROM playlist_songs
		  WHERE playlist_id = $1
		  INNER JOIN playlistsongs ON songs.id = playlistsongs.song_id`

	rows, err := m.DB.Query(query, playlistID)
	if err != nil {
		return nil, err
	}
	songs := []*Song{}
	for rows.Next() {
		var song Song

		err := rows.Scan(
			&song.ID,
			&song.Created_At,
			&song.Name,
			&song.Artist,
			&song.Version,
		)

		if err != nil {
			return nil, err
		}

		songs = append(songs, &song)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *SongModel) Update(song *Song) error {
	query := `UPDATE songs SET artist = $1, thumbnail = $2, song_url = $3, version = version + 1
	WHERE id = $4 AND version = $5
	RETURNING version`

	args := []any{
		song.Artist,
		song.Thumbnail,
		song.SongURL,
		song.ID,
		song.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&song.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil

}

func (m *SongModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM songs
		  WHERE id = $1 `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
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
