package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mycok/sunrise-api/internal/data"
	"github.com/mycok/sunrise-api/internal/validator"
)

func (app *application) createMovieHandler(rw http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(rw, r, &input)
	if err != nil {
		app.badRequestResponse(rw, r, err)

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
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

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
	err = app.writeJSON(rw, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}
}

func (app *application) showMovieHandler(rw http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(rw, r)

		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(rw, r)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	err = app.writeJSON(rw, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}
}

func (app *application) listMoviesHandler(rw http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	movies, metadata, err := app.models.Movies.List(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	err = app.writeJSON(rw, http.StatusOK, envelope{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}
}

func (app *application) replaceMovieHandler(rw http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(rw, r)

		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(rw, r)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	// Declare an input struct to hold the expected data from the client.
	var updateData struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Read JSON request body data into the updateData struct
	err = app.readJSON(rw, r, &updateData)
	if err != nil {
		app.badRequestResponse(rw, r, err)

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
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	// Pass the updated movie record to our new Update() method.
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(rw, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)
	}
}

func (app *application) updateMovieHandler(rw http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(rw, r)

		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(rw, r)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	// Declare an input struct to hold the expected data from the client.
	var updateData struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Read JSON request body data into the updateData struct
	err = app.readJSON(rw, r, &updateData)
	if err != nil {
		app.badRequestResponse(rw, r, err)

		return
	}

	// Map the values from the request body / updateData struct to the appropriate movie fields
	if updateData.Title != nil {
		movie.Title = *updateData.Title
	}

	if updateData.Year != nil {
		movie.Year = *updateData.Year
	}

	if updateData.Runtime != nil {
		movie.Runtime = *updateData.Runtime
	}

	if updateData.Genres != nil {
		movie.Genres = updateData.Genres
	}

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	// Pass the updated movie record to our new Update() method.
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(rw, r)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(rw, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)
	}
}

func (app *application) deleteMovieHandler(rw http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(rw, r)

		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(rw, r)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	err = app.writeJSON(rw, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}
}
