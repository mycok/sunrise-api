package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mycok/sunrise-api/internal/data"
)

func (app *application) createMovieHandler(wr http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := app.readJSON(wr, r, &input)
	if err != nil {
		app.badRequestResponse(wr, r, err)

		return
	}

	fmt.Fprintf(wr, "%+v\n", input)
}

func (app *application) showMovieHandler(wr http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(wr, r)

		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "so damn funny",
		Runtime:   102,
		Genres:    []string{"comedy", "drama", "sci-fi"},
		Version:   1,
	}

	err = app.writeJSON(wr, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}
}
