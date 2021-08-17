package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mycok/sunrise-api/internal/data"
	"github.com/mycok/sunrise-api/internal/validator"
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
	// Declare a client struct to store client's information
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// Declare a map and a mutex to store client's IP addresses and limiter instances
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening
			//  while the cleanup is taking place
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Unlock the mutex when cleanup is complete
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(rw, r, err)

				return
			}

			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()

			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new rate limiter and add the IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(
						rate.Limit(app.config.limiter.rps),
						app.config.limiter.burst,
					),
				}

			}

			clients[ip].lastSeen = time.Now()

			// Call limiter.Allow() method on the rate limiter for the current client's IP address
			// to see if the request is permitted, and if it's not, unlock the mutex
			// then call the rateLimitExceededResponse() helper to return a 429 Too Many
			// Requests response.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(rw, r)

				return
			}

			// Unlock the mutex before calling the next handler in the chain
			mu.Unlock()
		}

		next.ServeHTTP(rw, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		rw.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no such header found.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)

			next.ServeHTTP(rw, r)

			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into its constituent parts, and if the
		// header isn't in the expected format we return a 401 Unauthorized response
		// using the invalidAuthenticationTokenResponse() helper.
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthTokenResponse(rw, r)

			return
		}

		token := headerParts[1]

		v := validator.New()
		if data.ValidatePlainTextToken(v, token); !v.Valid() {
			app.invalidAuthTokenResponse(rw, r)

			return
		}

		// Retrieve the details of the user associated with the authentication token,
		// again calling the invalidAuthenticationTokenResponse() helper if no
		// matching record was found.
		user, err := app.models.Users.GetForToken(token, data.ScopeAuthentication)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthTokenResponse(rw, r)
			default:
				app.serverErrorResponse(rw, r, err)
			}

			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(rw, r)
	})
}

func (app *application) requiresAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(rw, r)

			return
		}

		next.ServeHTTP(rw, r)
	})
}

// requiresActivatedUser() checks that the user is both authenticated and activated
func (app *application) requiresActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(rw, r)

			return
		}

		next.ServeHTTP(rw, r)
	})

	return app.requiresAuthentication(fn)
}

func (app *application) requiresPermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		// Get the slice of permissions for the user.
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(rw, r, err)

			return
		}

		// Check if the slice includes the required permission. If it doesn't, then
		// return a 403 Forbidden response.
		if !permissions.Include(code) {
			app.notPermittedResponse(rw, r)

			return
		}

		next.ServeHTTP(rw, r)
	}

	return app.requiresActivatedUser(fn)
}

func(app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Vary", "Origin")

		origin := r.Header.Get("Origin")

		// Only run this if there's an Origin request header present AND at least one 
		// trusted origin is configured.
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			// Loop through the list of trusted origins, checking to see if the request 
			// origin exactly matches one of them.
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					rw.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}

		next.ServeHTTP(rw, r)
	})
}
