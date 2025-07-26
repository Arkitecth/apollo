package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Arkitecth/apollo/internal/data"
	"github.com/Arkitecth/apollo/validator"
)

func (app *application) showSongHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	song, err := app.models.SongModel.Get(id)
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

	err = app.writeJSON(w, http.StatusOK, envelope{"song": song}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) listSongsHandler(w http.ResponseWriter, r *http.Request) {
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

	songs, err := app.models.SongModel.GetAll(input.Artist, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"songs": songs}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) uploadSongHandler(w http.ResponseWriter, r *http.Request) {
	url, err := app.uploadS3(r, "file")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "song uploaded sucessfully", "url": url}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createSongHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Artist    string `json:"artist"`
		Name      string `json:"name"`
		Thumbnail string `json:"thumbnail"`
		SongURL   string `json:"song_url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	song := data.Song{}
	song.Artist = input.Artist
	song.Name = input.Name
	song.SongURL = input.SongURL
	song.Thumbnail = input.Thumbnail

	v := validator.New()

	if data.ValidateSong(v, &song); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	err = app.models.SongModel.Insert(&song)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/songs/%d", song.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"song": song}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) deleteSongHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.SongModel.Delete(id)
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

func (app *application) updateSongHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	song, err := app.models.SongModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	var input struct {
		Artist    *string `json:"artist"`
		Name      *string `json:"name"`
		SongURL   *string `json:"song_url"`
		Thumbnail *string `json:"thumbnail"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		song.Name = *input.Name
	}

	if input.Artist != nil {
		song.Artist = *input.Artist
	}

	if input.SongURL != nil {
		song.SongURL = *input.SongURL
	}

	if input.Thumbnail != nil {
		song.Thumbnail = *input.Thumbnail
	}

	v := validator.New()

	if data.ValidateSong(v, song); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	err = app.models.SongModel.Update(song)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"song": song}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
