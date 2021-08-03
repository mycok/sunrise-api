package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mycok/sunrise-api/internal/data"
	"github.com/mycok/sunrise-api/internal/validator"
)

func (app *application) createMovieHandler(wr http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(wr, r, &input)
	if err != nil {
		app.badRequestResponse(wr, r, err)

		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(wr, r, v.Errors)

		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}

	// When sending an HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(wr, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}
}

func (app *application) showMovieHandler(wr http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(wr, r)

		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(wr, r)
		default:
			app.serverErrorResponse(wr, r, err)
		}

		return
	}

	err = app.writeJSON(wr, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}
}

func (app *application) updateMovieHandler(wr http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(wr, r)

		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(wr, r)
		default:
			app.serverErrorResponse(wr, r, err)
		}

		return
	}

	// Declare an input struct to hold the expected data from the client.
	var updateData struct {
		Title string `json:"title"`
		Year int32 `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}

	// Read JSON request body data into the updateData struct
	err = app.readJSON(wr, r, &updateData)
	if err != nil {
		app.badRequestResponse(wr, r, err)

		return
	}

	// Map the values from the request body / updateData struct to the appropriate movie fields
	movie.Title = updateData.Title
	movie.Year = updateData.Year
	movie.Runtime = updateData.Runtime
	movie.Genres = updateData.Genres

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(wr, r, v.Errors)

		return
	}

	// Pass the updated movie record to our new Update() method.
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(wr, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(wr, r, err)
	}
}

func (app *application) deleteMovieHandler(wr http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(wr, r)

		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch  {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(wr, r)
		default:
			app.serverErrorResponse(wr, r, err)
			
		}
	}

	err = app.writeJSON(wr, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(wr, r, err)

		return
	}
}