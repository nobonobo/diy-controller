package motor

import "github.com/nobonobo/q16"

type Motor interface {
	Setup() error
	Enable() error
	Disable() error
	State() (*State, error)
	Output(pow q16.Fixed) error
}
