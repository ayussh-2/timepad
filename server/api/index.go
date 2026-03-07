package handler

import (
	"net/http"

	"github.com/ayussh-2/timepad/app"
)

// Handler is the single HTTP entry point exposed to Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	app.Handler().ServeHTTP(w, r)
}
