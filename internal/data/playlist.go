package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Arkitecth/apollo/validator"
)

type Playlist struct {
	ID         int64     `json:"id"`
	Created_At time.Time `json:"created_at"`
	Name       string    `json:"name"`
	UserID     int64     `json:"user_id"`
	Version    int       `json:"version"`
}

type PlaylistModel struct {
	DB *sql.DB
}

func ValidateName(v *validator.Validator, name string) {
	v.Check(len(name) > 50, "name", "cannot be greater than 50 bytes")
}

func (m *PlaylistModel) Insert(playlist *Playlist) error {
	query := `INSERT INTO playlists (name, user_id) 
		  VALUES ($1, $2)
		  RETURNING id, created_at, version`
	args := []any{playlist.Name, playlist.UserID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&playlist.ID, &playlist.Created_At, &playlist.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}

func (m *PlaylistModel) Get(playlistID int64) (*Playlist, error) {
	if playlistID < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id, created_at, name, user_id, version FROM playlists
		  WHERE id = $1 `

	playlist := &Playlist{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, playlistID).Scan(
		&playlist.ID,
		&playlist.Created_At,
		&playlist.Name,
		&playlist.UserID,
		&playlist.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return playlist, nil

}

func (m *PlaylistModel) InsertSong(songID int64, playlistID int64) error {
	query := `INSERT INTO playlist_songs (song_id, playlist_id)
		  VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, songID, playlistID)
	if err != nil {
		return err
	}
	return nil
}

func (m *PlaylistModel) GetAll(userID int64) ([]*Playlist, error) {
	query := `SELECT id, created_at, name FROM playlists 
		  WHERE user_id = $1`

	playlists := []*Playlist{}

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var playlist Playlist
		err := rows.Scan(&playlist.ID, &playlist.Created_At, &playlist.Name)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, &playlist)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (m *PlaylistModel) Update(playlist *Playlist) error {
	query := `UPDATE playlists SET name = $1, version = version + 1
		  WHERE id = $2 AND version = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		&playlist.Name,
		&playlist.ID,
		&playlist.Version,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&playlist.Version)
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

func (m *PlaylistModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM playlists
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

func (m *PlaylistModel) DeleteSongFromPlaylist(songID int64, playlistID int64) error {
	query := `DELETE FROM playlist_songs WHERE song_id = $1 AND playlist_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, songID, playlistID)
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

func (m *PlaylistModel) GetSongsFromPlaylist(playlistID int64, artist string, name string, filters Filters) ([]*Song, error) {
	query := fmt.Sprintf(`
	SELECT songs.id, songs.created_at, songs.artist, songs.name, songs.song_url, songs.thumbnail, songs.version
	FROM songs 
	INNER JOIN playlist_songs ON song_id = songs.id
	WHERE playlist_songs.playlist_id = $1
	AND (to_tsvector('simple', artist) @@ plainto_tsquery('simple', $2) OR $2 = '') 
	AND (to_tsvector('simple', name) @@ plainto_tsquery('simple', $3) OR $3 = '')
	ORDER BY %s %s, id ASC
	LIMIT $4 OFFSET $5
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{playlistID, artist, name, filters.limit(), filters.offset()}

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
			&song.Artist,
			&song.Name,
			&song.SongURL,
			&song.Thumbnail,
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
