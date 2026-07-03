//go:build ddt1502

package motor

import (
	"encoding/binary"
	"runtime"
	"time"

	"github.com/nobonobo/q16"
	"tinygo.org/x/drivers/mcp2515"

	"github.com/nobonobo/diy-controller/settings"
)

/*
DDT M1502D: https://ssci.to/9219

得られる内部情報:
- 角度: 0..360degを0..32767で獲得（回転数集積無し）
- 速度: 100*rpm単位で獲得できる 210rpm Max
- 電流: 100*mA単位で獲得できる 55A Max
- 温度: 1.0℃単位で獲得できる

出力（電流モード）:
- 電流: -32767..32767 = -55 .. +55 [A]
*/
const (
	MaxOutput   = 32767
	MaxTorque   = q16.Fixed(97 * q16.Scale / 10) // 9.7 [N·m]
	CanRate     = mcp2515.CAN500kBps
	CanExtended = false
)

// DefaultSettings デフォルト設定
func DefaultSettings() settings.Settings {
	return settings.Settings{
		Neutral:           q16.DegToRad(q16.FromFloat32(-8)),      // [deg]
		HalfOfL2L:         q16.DegToRad(q16.FromFloat32(540 / 2)), // [deg]
		KLock:             q16.FromFloat32(0.5),                   // [N·m/rad]
		KSpring:           q16.FromFloat32(0.0),                   // [N·m/rad]
		KSpringDeadBand:   q16.DegToRad(q16.FromFloat32(1)),       // [deg]
		KSpringLimit:      q16.FromFloat32(0.01),                  // [0.0, 1.0]
		KDamper:           q16.FromFloat32(-2.0),                  // [N·m·s/rad] Damper (minus: cogging torque cancel)
		KDamperDeadBand:   q16.FromFloat32(0.0),                   // [rad/s]
		KInertia:          q16.FromFloat32(-0.03),                 // [N·m·s²/rad] inertia
		KInertiaDeadBand:  q16.FromFloat32(0.1),                   // [rad/s²]
		KFriction:         q16.FromFloat32(0.0),                   // [N·m·s/rad] friction
		KFrictionDeadBand: q16.DegToRad(q16.FromFloat32(0)),       // [deg]
		Backlash:          q16.DegToRad(q16.FromFloat32(1)),       // [deg]
		MinOut:            q16.FromFloat32(0.0),                   // [0.0, 1.0]
		MaxOut:            q16.FromFloat32(1.0),                   // [0.0, 1.0]
		MaxSpeed:          q16.FromFloat32(1.0),                   // [rad/s]
		KBrake:            q16.FromFloat32(0.5),                   // [N·m·s/rad] Damper (normaly minus)
	}
}

func powToMax(v q16.Fixed) int16 {
	output := (int64(v) * MaxOutput) >> q16.ShiftBits
	if output > MaxOutput {
		output = MaxOutput
	} else if output < -MaxOutput {
		output = -MaxOutput
	}
	return int16(output)
}

var _ = (Motor)((*DDT1502)(nil))

type DDT1502 struct {
	can       *mcp2515.Device
	lastAngle int16
	nRound    int
}

func New(can *mcp2515.Device) *DDT1502 {
	return &DDT1502{can: can}
}

func (m *DDT1502) read() (*mcp2515.CANMsg, error) {
	for !m.can.Received() {
		runtime.Gosched()
	}
	return m.can.Rx()
}

func (m *DDT1502) State() (state *State, err error) {
	if err := m.can.Tx(0x107, 8, []byte{0x01, 0x01, 0x02, 0x04, 0x55, 0, 0, 0}); err != nil {
		return nil, err
	}
	msg, err := m.read()
	if err != nil {
		return nil, err
	}
	verocity := -int16(binary.BigEndian.Uint16(msg.Data[0:2]))      // -220 .. 220 rpm の 100倍値
	current := -int16(binary.BigEndian.Uint16(msg.Data[2:4]))       // -32767 .. 32767 = -55 .. 55 A
	rawAngle := -int16(binary.BigEndian.Uint16(msg.Data[4:6]) << 1) // -32767 .. 32767 = -180 .. +180 deg
	diff := int32(rawAngle) - int32(m.lastAngle)
	if diff > 32767 {
		// -32767 → 32767 (逆回転)
		m.nRound--
	} else if diff < -32767 {
		// 32767 → -32767 (正回転)
		m.nRound++
	}
	m.lastAngle = rawAngle
	totalTurns := (int64(m.nRound) * 65536) + int64(rawAngle)

	const rpm100ToRadQ32 int64 = 4499045 // round((π / 3000) * 2^32)
	const currentScale = 110002          // round(55000 * 65536 / 32767)
	const roundToRadQ16 = 411775         // round(2π × 65536)
	return &State{
		Velocity: q16.Fixed((int64(verocity) * rpm100ToRadQ32) >> q16.ShiftBits),
		Current:  q16.Fixed((int64(current) * currentScale)),
		Angle:    q16.Fixed((int64(totalTurns) * roundToRadQ16) >> q16.ShiftBits),
	}, nil
}

// Setup ddt1502モーターコントローラーを初期化する
func (m *DDT1502) Setup() error {
	if err := m.can.Tx(0x109, 8, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	_, err := m.read()
	if err != nil {
		return err
	}
	if err := m.can.Tx(0x106, 8, []byte{0x80, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	_, err = m.read()
	if err != nil {
		return err
	}
	if err := m.can.Tx(0x105, 8, []byte{0x00, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	_, err = m.read()
	if err != nil {
		return err
	}
	return nil
}

// Enable ddt1502モーターコントローラーを有効にする
func (m *DDT1502) Enable() error {
	if err := m.can.Tx(0x105, 8, []byte{0x0A, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return m.Setup()
}

// Disable ddt1502モーターコントローラーを無効にする
func (m *DDT1502) Disable() error {
	if err := m.can.Tx(0x105, 8, []byte{0x09, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	return nil
}

// Output ddt1502にモーター出力を送信する
func (m *DDT1502) Output(pow q16.Fixed) error {
	raw := powToMax(pow)
	buf := [8]byte{}
	binary.BigEndian.PutUint16(buf[0:2], uint16(-raw))
	err := m.can.Tx(0x32, uint8(len(buf)), buf[:])
	//m.read()
	return err
}
