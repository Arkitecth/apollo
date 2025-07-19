package main

import (
	"errors"
	"fmt"
	"github.com/Arkitecth/apollo/internal/data"
	"github.com/Arkitecth/apollo/validator"
	"net/http"
)

func (app *application) showPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	playlist, err := app.models.PlaylistModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return

		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"playlist": playlist}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	//How do we create a playlist
	var input struct {
		Name   string `json:"name"`
		UserID int64  `json:"user_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	_, err = app.models.UserModel.GetById(input.UserID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.badRequestResponse(w, r, err)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	playlist := data.Playlist{}
	playlist.Name = input.Name
	playlist.UserID = input.UserID

	v := validator.New()

	if data.ValidateName(v, playlist.Name); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	err = app.models.PlaylistModel.Insert(&playlist)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/playlist/%d", playlist.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"playlist": playlist}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) deletePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.PlaylistModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return

		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"song": "record succesfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) listPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	playlists, err := app.models.PlaylistModel.GetAll(1)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"playlists": playlists}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	var input struct {
		Name string `json:"name"`
	}

	playlist, err := app.models.PlaylistModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	v := validator.New()
	if data.ValidateName(v, input.Name); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	playlist.Name = input.Name

	err = app.models.PlaylistModel.Update(playlist)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"playlist": playlist}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) addSongToPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SongID     int64 `json:"song_id"`
		PlaylistID int64 `json:"playlist_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	_, err = app.models.SongModel.Get(input.SongID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.models.PlaylistModel.InsertSong(input.SongID, input.PlaylistID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "song has been added to playlist"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) removeSongFromPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	playlist_id, err := app.readParamID(r, "playlist_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	song_id, err := app.readParamID(r, "song_id")
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.PlaylistModel.DeleteSongFromPlaylist(song_id, playlist_id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "song has been deleted from playlist successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showSongsFromPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	playlist_id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var input struct {
		Name   string
		Artist string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Artist = app.readString(qs, "artist", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "artist", "-id", "-name", "-artist"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	songs, err := app.models.PlaylistModel.GetSongsFromPlaylist(playlist_id, input.Artist, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"playlist_songs": songs}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
