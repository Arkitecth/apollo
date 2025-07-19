package main

import (
	"errors"
	"net/http"

	"github.com/Arkitecth/apollo/internal/data"
	"github.com/Arkitecth/apollo/validator"
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

	err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
		app.logger.Info("HERE")
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
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
