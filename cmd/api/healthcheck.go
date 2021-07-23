package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(wr http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(wr, "status: available")
	fmt.Fprintf(wr, "environment: %s\n", app.config.env)
	fmt.Fprintf(wr, "version: %s\n", version)
}

