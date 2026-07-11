package main

import (
	"fmt"

	"github.com/nobonobo/diy-controller/board"
	"github.com/nobonobo/diy-controller/controller"
	"github.com/nobonobo/diy-controller/motor"
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
	ss := s.controller.Settings()
	if err := ss.ValidateAll(); err != nil {
		return fmt.Errorf("failed to validate settings: %w", err)
	}
	b, err := ss.MarshalBinary()
	if err != nil {
		return err
	}
	if err := board.WriteFlashBlock(b); err != nil {
		return err
	}
	return nil
}

func (s *Service) Load() error {
	b, err := board.ReadFlashBlock()
	if err != nil {
		return fmt.Errorf("failed to read flash block: %w", err)
	}
	ss := &settings.Settings{}
	if err := ss.UnmarshalBinary(b); err != nil {
		s.controller.SetSettings(motor.DefaultSettings())
		return fmt.Errorf("failed to unmarshal binary: %w", err)
	}
	if err := ss.ValidateAll(); err != nil {
		return fmt.Errorf("failed to validate binary: %w", err)
	}
	s.controller.SetSettings(*ss)
	return nil
}
