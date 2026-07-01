//go:build drv8311

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
RollerCAN Lite: https://ssci.to/10027

得られる内部情報:
- 角度: 100*deg単位で獲得できる(回転数も集積済み)
- 速度: 100*rpm単位で獲得できる
- 電流: 100*mA単位で獲得できる
- 温度: 1.0℃単位で獲得できる

出力（電流モード）:
- 電流: 100*mA単位(-120000..1200000 Max1.2A)
*/

const (
	MaxOutput   = 120000
	MaxTorque   = q16.Fixed(39322) // 0.6 [N·m] round(0.6 * 65536)
	CanRate     = mcp2515.CAN1000kBps
	CanExtended = true
)

// DefaultSettings デフォルト設定
func DefaultSettings() settings.Settings {
	return settings.Settings{
		Neutral:      q16.DegToRad(q16.FromFloat32(0.0)),         // [deg]
		HalfOfL2L:    q16.DegToRad(q16.FromFloat32(540 / 2)),     // [deg]
		KLock:        q16.Div(q16.FromFloat32(5.0), MaxTorque),   // [N·m/rad]
		KSpring:      q16.Div(q16.FromFloat32(0.5), MaxTorque),   // [N·m/rad]
		KSpringLimit: q16.FromFloat32(0.1),                       // [0.0, 1.0]
		KDamper:      q16.Div(q16.FromFloat32(4.0), MaxTorque),   // [N·m·s/rad] Damper (normaly minus)
		KInertia:     q16.Div(q16.FromFloat32(0.02), MaxTorque),  // [N·m·s²/rad] inertia (normaly minus)
		KFriction:    q16.Div(q16.FromFloat32(0.003), MaxTorque), // [N·m·s/rad] friction (normaly minus)
		Backlash:     q16.DegToRad(q16.FromFloat32(1)),           // [deg]
		MinOut:       q16.FromFloat32(0.0),                       // [0.0, 1.0]
		MaxOut:       q16.FromFloat32(1.0),                       // [0.0, 1.0]
		MaxSpeed:     q16.FromFloat32(5.0),                       // [rad/s]
		KBrake:       q16.Div(q16.FromFloat32(0.005), MaxTorque), // [N·m·s/rad] Damper (normaly minus)
	}
}

func powToMax(v q16.Fixed) int64 {
	output := (int64(v) * MaxOutput) >> 16
	if output > MaxOutput {
		output = MaxOutput
	} else if output < -MaxOutput {
		output = -MaxOutput
	}
	return output
}

var _ = (Motor)((*DRV8311)(nil))

type DRV8311 struct {
	can *mcp2515.Device
}

func New(can *mcp2515.Device) *DRV8311 {
	return &DRV8311{can: can}
}

const (
	MotorID = 0xA8
	HostID  = 0x00
)

// WriteFrame CANフレームを送信する
func (m *DRV8311) write(typ byte, data []byte) error {
	id := uint32(typ)<<24 | HostID<<8 | MotorID
	return m.can.Tx(id, uint8(len(data)), data)
}

// ReadFrame CANフレームを受信するまで待機し、受信したフレームを返す
func (m *DRV8311) read() (*mcp2515.CANMsg, error) {
	for !m.can.Received() {
		runtime.Gosched()
	}
	return m.can.Rx()
}

// Setup drv8311モーターコントローラーを初期化する
func (m *DRV8311) Setup() error {
	// Set Current Mode
	if err := m.write(0x12, []byte{0x05, 0x70, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	// Set Current 0.0[mA]
	if err := m.write(0x12, []byte{0x06, 0x70, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	if err := m.write(0x03, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	return nil
}

// State drv8311からモーターステータスを取得する
func (m *DRV8311) State() (*State, error) {
	if err := m.write(0x11, []byte{0x30, 0x70, 0, 0, 0, 0, 0, 0}); err != nil {
		return nil, err
	}
	vel, err := m.read()
	if err != nil {
		return nil, err
	}
	velValue := int32(binary.LittleEndian.Uint32(vel.Data[4:8])) // [100*rpm]

	if err := m.write(0x11, []byte{0x31, 0x70, 0, 0, 0, 0, 0, 0}); err != nil {
		return nil, err
	}
	pos, err := m.read()
	if err != nil {
		return nil, err
	}
	posValue := int32(binary.LittleEndian.Uint32(pos.Data[4:8])) // [100*deg]

	if err := m.write(0x11, []byte{0x32, 0x70, 0, 0, 0, 0, 0, 0}); err != nil {
		return nil, err
	}
	cur, err := m.read()
	if err != nil {
		return nil, err
	}
	curValue := int32(binary.LittleEndian.Uint32(cur.Data[4:8])) // [100*mA]

	if err := m.write(0x11, []byte{0x35, 0x70, 0, 0, 0, 0, 0, 0}); err != nil {
		return nil, err
	}
	tmp, err := m.read()
	if err != nil {
		return nil, err
	}
	tmpValue := int32(binary.LittleEndian.Uint32(tmp.Data[4:8])) // [100*℃]

	const rpm100ToRadQ16 int64 = 4499045 // round((π / 3000) * 2^16)
	const mA100ToQ16Q16 int64 = 42949673 // round(2^16 / 100)
	const deg100ToRadQ16 int64 = 749841  // round((π / 18000) * 2^16)
	return &State{
		Velocity:    q16.Fixed((int64(velValue) * rpm100ToRadQ16) >> q16.ShiftBits),
		Current:     q16.Fixed((int64(curValue) * mA100ToQ16Q16) >> q16.ShiftBits),
		Angle:       q16.Fixed((int64(posValue) * deg100ToRadQ16) >> q16.ShiftBits),
		Temperature: q16.Fixed(tmpValue << q16.ShiftBits),
	}, nil
}

// Enable drv8311モーターコントローラーを有効にする
func (m *DRV8311) Enable() error {
	if err := m.write(0x03, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return m.Setup()
}

// Disable drv8311モーターコントローラーを無効にする
func (m *DRV8311) Disable() error {
	if err := m.write(0x04, []byte{0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	if _, err := m.read(); err != nil {
		return err
	}
	return nil
}

// Output drv8311にモーター出力を送信する
func (m *DRV8311) Output(pow q16.Fixed) error {
	buf := [8]byte{0x06, 0x70, 0, 0, 0, 0, 0, 0}
	raw := powToMax(pow)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(raw))
	err := m.write(0x12, buf[:])
	m.read()
	return err
}
