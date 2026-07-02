module github.com/nobonobo/diy-controller

go 1.24.0

require (
	github.com/marben/irpc v0.0.0-00010101000000-000000000000
	github.com/nobonobo/q16 v0.0.0-20260629173311-acacaa779693
	tinygo.org/x/drivers v0.35.0
)

require github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect

replace github.com/marben/irpc => github.com/nobonobo/irpc v0.0.0-20260702024300-0aa8db983f4b
