// ventre-panel is a stateless cross-platform desktop client for batch SSH
// command execution and file transfer, powered by ventre-transport.
package main

import (
	"github.com/ventre-go/ventre-panel/internal/app"
)

func main() {
	a := app.New()
	app.Run(a)
}
