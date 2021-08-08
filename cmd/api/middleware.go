package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverFromPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been // sent.
				rw.Header().Set("Connection", "close")
				// The recover() function returns type interface{},
				// so we use fmt.Errorf() to normalize it into an error.
				app.serverErrorResponse(rw, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(rw, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Initialize a new limiter instance which allows an average of 2 requests per second, 
	// with a maximum of 4 requests in a single ‘burst’.
	limiter := rate.NewLimiter(2, 4)

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Call limiter.Allow() to see if the request is permitted, and if it's not,
		// then we call the rateLimitExceededResponse() helper to return a 429 Too Many 
		// Requests response.
		if !limiter.Allow() {
			app.rateLimitExceededResponse(rw, r)

			return
		}

		next.ServeHTTP(rw, r)
	})
}
