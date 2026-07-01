//go:build !drv8311 && !ddt1502

package motor

import (
	"tinygo.org/x/drivers/mcp2515"

	"github.com/nobonobo/q16"

	"github.com/nobonobo/diy-controller/settings"
)

const (
	MaxOutput   = 120000
	MaxTorque   = q16.Fixed(39322) // 0.6 [N·m] round(0.6 * 65536)
	CanRate     = mcp2515.CAN1000kBps
	CanExtended = true
)

// DefaultSettings デフォルト設定
func DefaultSettings() settings.Settings {
	return settings.Settings{
		Neutral:      q16.DegToRad(q16.FromFloat32(0.0)),          // [deg]
		HalfOfL2L:    q16.DegToRad(q16.FromFloat32(540 / 2)),      // [deg]
		KLock:        q16.Div(q16.FromFloat32(-5.0), MaxTorque),   // [N·m/rad]
		KSpring:      q16.Div(q16.FromFloat32(-0.5), MaxTorque),   // [N·m/rad]
		KSpringLimit: q16.FromFloat32(0.1),                        // [0.0, 1.0]
		KDamper:      q16.Div(q16.FromFloat32(4.0), MaxTorque),    // [N·m·s/rad] Damper (normaly minus)
		KInertia:     q16.Div(q16.FromFloat32(0.02), MaxTorque),   // [N·m·s²/rad] inertia (normaly minus)
		KFriction:    q16.Div(q16.FromFloat32(-0.003), MaxTorque), // [N·m·s/rad] friction (normaly minus)
		Backlash:     q16.DegToRad(q16.FromFloat32(1)),            // [deg]
		MinOut:       q16.FromFloat32(0.0),                        // [0.0, 1.0]
		MaxOut:       q16.FromFloat32(1.0),                        // [0.0, 1.0]
		MaxSpeed:     q16.FromFloat32(5.0),                        // [rad/s]
		KBrake:       q16.Div(q16.FromFloat32(-0.005), MaxTorque), // [N·m·s/rad] Damper (normaly minus)
	}
}

var _ = (Motor)((*Dummy)(nil))

type Dummy struct{}

func New(can *mcp2515.Device) *Dummy {
	return &Dummy{}
}

func (m *Dummy) Setup() error {
	return nil
}

func (m *Dummy) State() (state *State, err error) {
	return &State{}, nil
}

func (m *Dummy) Output(pow q16.Fixed) error {
	return nil
}

func (m *Dummy) Enable() error {
	return nil
}

func (m *Dummy) Disable() error {
	return nil
}
