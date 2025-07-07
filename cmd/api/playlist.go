package main

import "net/http"

func (app *application) showPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}
	Playlist, err := app.PlaylistModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return

		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, r, http.StatusOK, envelope{"Playlist": Playlist})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) uploadPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	url, err := app.uploadS3(r, "file")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, r, http.StatusOK, envelope{"message": "Playlist uploaded sucessfully", "url": url})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Artist      string
		Name        string
		Thumbnail   string
		PlaylistURL string
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	Playlist := Playlist{}
	Playlist.Artist = input.Artist
	Playlist.Name = input.Name
	Playlist.PlaylistURL = input.PlaylistURL
	Playlist.Thumbnail = input.Thumbnail

	v := app.PlaylistModel.ValidatePlaylist(&Playlist)
	if !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	err = app.PlaylistModel.Insert(&Playlist)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, r, http.StatusOK, envelope{"Playlist": Playlist})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) deletePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}
	err = app.PlaylistModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			app.notFoundResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, r, http.StatusOK, envelope{"Playlist": "record succesfully deleted"})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
