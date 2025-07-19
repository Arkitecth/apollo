package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/songs/:id", app.deleteSongHandler)
	router.HandlerFunc(http.MethodPost, "/v1/songs", app.createSongHandler)
	router.HandlerFunc(http.MethodPost, "/v1/upload/songs", app.uploadSongHandler)

	router.HandlerFunc(http.MethodGet, "/v1/songs/:id", app.showSongHandler)
	router.HandlerFunc(http.MethodGet, "/v1/songs", app.listSongsHandler)

	//Users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	//Playlist
	router.HandlerFunc(http.MethodGet, "/v1/playlists/show/playlist/:id", app.showPlaylistHandler)
	router.HandlerFunc(http.MethodPost, "/v1/playlists/create/playlist", app.createPlaylistHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/playlists/delete/playlist/:id", app.deletePlaylistHandler)
	router.HandlerFunc(http.MethodGet, "/v1/playlists/list/playlist", app.listPlaylistHandler) // Needs to be able to get Authenticated User ID Playlist ID

	router.HandlerFunc(http.MethodPost, "/v1/playlists/add/songs", app.addSongToPlaylistHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/playlists/remove/songs/:song_id/:playlist_id", app.removeSongFromPlaylistHandler)
	router.HandlerFunc(http.MethodGet, "/v1/playlists/show/songs/:id", app.showSongsFromPlaylistHandler)

	return app.recoverPanic(app.rateLimit(router))
}
