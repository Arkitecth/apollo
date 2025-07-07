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
		Artist    string
		Name      string
		Thumbnail string
		SongURL   string
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

		default:
			app.serverErrorResponse(w, r, err)
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
		Artist    string `json:"artist"`
		SongURL   string `json:"song_url"`
		Thumbnail string `json:"thumbnail"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	song.Artist = input.Artist
	song.SongURL = input.SongURL
	song.Thumbnail = input.Thumbnail

	v := validator.New()

	if data.ValidateSong(v, song); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	err = app.models.SongModel.Update(song)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"song": song}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
