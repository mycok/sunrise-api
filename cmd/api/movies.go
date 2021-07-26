package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mycok/sunrise-api/internal/data"
)

func (app *application) createMovieHandler(wr http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(wr, "create a new movies")
}

func (app *application) showMovieHandler(wr http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(wr, r)

		return
	}

	movie := data.Movie{
		ID: id,
		CreatedAt: time.Now(),
		Title: "so damn funny",
		Runtime: 102,
		Genres: []string{"comedy", "drama", "sci-fi"},
		Version: 1,
	}

	err = app.writeJSON(wr, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}
}
