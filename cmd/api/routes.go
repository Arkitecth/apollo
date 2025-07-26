package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/songs/:id", app.requireAuthorizedUser("songs:delete", app.deleteSongHandler))
	router.HandlerFunc(http.MethodPost, "/v1/songs", app.requireAuthorizedUser("songs:create", app.createSongHandler))
	router.HandlerFunc(http.MethodPost, "/v1/upload/songs", app.requireAuthorizedUser("songs:upload", app.uploadSongHandler))

	router.HandlerFunc(http.MethodGet, "/v1/songs/:id", app.showSongHandler)
	router.HandlerFunc(http.MethodGet, "/v1/songs", app.listSongsHandler)

	//Users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	//Playlist
	router.HandlerFunc(http.MethodGet, "/v1/playlists/show/playlist/:id", app.requireActivatedUser(app.showPlaylistHandler))
	router.HandlerFunc(http.MethodPost, "/v1/playlists/create/playlist", app.requireActivatedUser(app.createPlaylistHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/playlists/delete/playlist/:id", app.requireActivatedUser(app.deletePlaylistHandler))
	router.HandlerFunc(http.MethodGet, "/v1/playlists/list/playlist", app.requireActivatedUser(app.listPlaylistHandler))

	router.HandlerFunc(http.MethodPost, "/v1/playlists/add/songs", app.requireActivatedUser(app.addSongToPlaylistHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/playlists/remove/songs/:song_id/:playlist_id", app.requireActivatedUser(app.removeSongFromPlaylistHandler))
	router.HandlerFunc(http.MethodGet, "/v1/playlists/show/songs/:id", app.requireActivatedUser(app.showSongsFromPlaylistHandler))

	//Tokens
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
