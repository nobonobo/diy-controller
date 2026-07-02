package service

//go:generate irpc $GOFILE

type Service interface {
	Gains() map[string]int32
	SetGains(s map[string]int32)
	Settings() map[string]int32
	SetSettings(s map[string]int32)
	Store() error
	Load() error
}
