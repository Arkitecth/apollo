package main

import (
	"fmt"
	"net/http"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	app.logger.Info("starting server on", "port", app.config.port, "env", app.config.env)

	err := srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
