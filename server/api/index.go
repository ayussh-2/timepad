// Package handler is the Vercel serverless entry point.
// Vercel's Go runtime looks for an exported http.HandlerFunc in /api/*.go files.
//
// This file intentionally imports NO internal/ packages — Vercel's runtime
// resolves the api/ directory as "handler/api" (outside the module tree),
// which would violate Go's internal-package restriction. All initialisation

package handler

import (
	"net/http"

	"github.com/ayussh-2/timepad/app"
)

// Handler is the single HTTP entry point exposed to Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	app.Handler().ServeHTTP(w, r)
}
