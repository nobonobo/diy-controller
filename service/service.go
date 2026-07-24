package service

//go:generate irpc $GOFILE

type Vibration struct {
	Gain       int32
	EffectType uint8
	Duration   int32
	Frequency  int32
}

type Envelope struct {
	AttackLevel int32
	FadeLevel   int32
	AttackTime  int32
	FadeTime    int32
}

type GamePad struct {
	XAxis   int16
	YAxis   int16
	ZAxis   int16
	RxAxis  int16
	RyAxis  int16
	RzAxis  int16
	Buttons uint16
	Hat     uint8
}

type Service interface {
	Gains() map[string]int32
	SetGains(s map[string]int32)
	Settings() map[string]int32
	SetSettings(s map[string]int32)
	Store() error
	Load() error
	Reset() error
	// additional effects
	SetVibration(index int, params *Vibration) error
	SetEnvelope(index int, params *Envelope) error
	StartVibration(index int) error
	StopVibration(index int) error
	StopAll() error
	ShowVibration(index int) (string, error)
	// gamepad
	Send(params *GamePad) error
}
