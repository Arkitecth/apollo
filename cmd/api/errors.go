package main

import (
	"fmt"
	"net/http"
)

func (app *application) logErrors(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, msg any) {
	env := envelope{"error": msg}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logErrors(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logErrors(r, err)
	app.errorResponse(w, r, http.StatusInternalServerError, "a server error occurred")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "the content you requested could not be found")
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%s method now allowed for this request", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (app *application) failedInvalidationResponse(w http.ResponseWriter, r *http.Request, errorMap map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errorMap)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	msg := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "your are not authorized to use this resource"

	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

func (app *application) notAuthorizedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your account does not have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
