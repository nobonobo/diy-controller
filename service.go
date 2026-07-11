package main

import (
	"fmt"

	"github.com/nobonobo/diy-controller/board"
	"github.com/nobonobo/diy-controller/controller"
	"github.com/nobonobo/diy-controller/motor"
	"github.com/nobonobo/diy-controller/service"
	"github.com/nobonobo/diy-controller/settings"
)

type Service struct {
	controller *controller.Controller
}

func (s *Service) Gains() map[string]int32 {
	return s.controller.Gains().ToMap()
}

func (s *Service) SetGains(p map[string]int32) {
	g := s.controller.Gains()
	newGains := g.Merge(p)
	if err := newGains.ValidateAll(); err != nil {
		println("failed to validate gains:", err)
		return
	}
	s.controller.SetGains(newGains)
}

func (s *Service) Settings() map[string]int32 {
	return s.controller.Settings().ToMap()
}

func (s *Service) SetSettings(p map[string]int32) {
	set := s.controller.Settings()
	newSet := set.Merge(p)
	if err := newSet.ValidateAll(); err != nil {
		println("failed to validate settings:", err)
		return
	}
	s.controller.SetSettings(newSet)
}

func (s *Service) Store() error {
	gs := s.controller.Gains()
	if err := gs.ValidateAll(); err != nil {
		return fmt.Errorf("failed to validate gains: %w", err)
	}
	ss := s.controller.Settings()
	if err := ss.ValidateAll(); err != nil {
		return fmt.Errorf("failed to validate settings: %w", err)
	}
	b, err := settings.Store(gs, ss)
	if err != nil {
		return err
	}
	if err := board.WriteFlashBlock(b); err != nil {
		return fmt.Errorf("failed to write flash block: %w", err)
	}
	return nil
}

func (s *Service) Load() error {
	b, err := board.ReadFlashBlock()
	if err != nil {
		return fmt.Errorf("failed to read flash block: %w", err)
	}
	gs, ss, err := settings.Load(b)
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}
	s.controller.SetGains(*gs)
	s.controller.SetSettings(*ss)
	return nil
}

func (s *Service) Reset() error {
	s.controller.SetGains(settings.NewGains())
	s.controller.SetSettings(motor.DefaultSettings())
	return nil
}

func (s *Service) SetVibration(index int, params *service.Vibration) error {
	s.controller.SetVibration(index, params)
	return nil
}

func (s *Service) SetEnvelope(index int, params *service.Envelope) error {
	s.controller.SetEnvelope(index, params)
	return nil
}

func (s *Service) StartVibration(index int) error {
	s.controller.StartVibration(index)
	return nil
}

func (s *Service) StopVibration(index int) error {
	s.controller.StopVibration(index)
	return nil
}

func (s *Service) StopAll() error {
	s.controller.StopAll()
	return nil
}

func (s *Service) ShowVibration(index int) (string, error) {
	return s.controller.ShowVibration(index)
}
