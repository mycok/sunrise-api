package main

import (
	"fmt"
	"net/http"
)

func (app *application) createMovieHandler(wr http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(wr, "create a new movies")
}

func (app *application) showMovieHandler(wr http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(wr, r)

		return
	}

	fmt.Fprintf(wr, "show the details of movie %d\n", id)
}
