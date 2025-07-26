package main

import (
	"errors"
	"github.com/Arkitecth/apollo/internal/data"
	"github.com/Arkitecth/apollo/validator"
	"net/http"
	"time"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	err = app.models.UserModel.Insert(user)
	if err != nil {

		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.Add("email", "a user with this email address already exists")
			app.failedInvalidationResponse(w, r, v.ErrorMap)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	err = app.models.PermissionModel.AddForUsers(user.ID, "songs:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	token, err := app.models.TokenModel.New(user.ID, 3*24*time.Hour, "activation")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activatonToken": token.Plaintext,
			"userID":         user.ID,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidatePlaintext(v, input.TokenPlainText); !v.Valid() {
		app.failedInvalidationResponse(w, r, v.ErrorMap)
		return
	}

	user, err := app.models.UserModel.GetUserFromToken(data.ScopeActivation, input.TokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.Add("token", "invalid or expired activation token")
			app.failedInvalidationResponse(w, r, v.ErrorMap)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.UserModel.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.TokenModel.DeleteAllForUsers(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
