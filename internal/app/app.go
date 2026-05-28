// Package app initializes the ventre-panel application.
package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/ventre-go/ventre-panel/internal/ui"
)

// New creates a new ventre-panel Fyne application.
func New() fyne.App {
	a := app.New()
	a.Settings().SetTheme(ui.NewTheme())
	return a
}

// Run starts the ventre-panel GUI.
func Run(a fyne.App) {
	w := a.NewWindow("ventre-panel — Batch SSH Panel")
	ui.BuildWindow(w)
	w.ShowAndRun()
}
