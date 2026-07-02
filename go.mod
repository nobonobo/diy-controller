module github.com/nobonobo/diy-controller

go 1.24.0

require (
	github.com/marben/irpc v0.0.1
	github.com/nobonobo/q16 v0.0.0-20260629173311-acacaa779693
	go.bug.st/serial v1.6.4
	tinygo.org/x/drivers v0.35.0
)

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

replace github.com/marben/irpc => github.com/nobonobo/irpc v0.0.0-20260702151806-f65614d435b1
