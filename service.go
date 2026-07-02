package main

import (
	"github.com/nobonobo/diy-controller/controller"
)

type Service struct {
	controller *controller.Controller
}

func (s *Service) Gains() map[string]int32 {
	return s.controller.Gains().ToMap()
}

func (s *Service) SetGains(p map[string]int32) {
	g := s.controller.Gains()
	s.controller.SetGains(g.Merge(p))
}

func (s *Service) Settings() map[string]int32 {
	return s.controller.Settings().ToMap()
}

func (s *Service) SetSettings(p map[string]int32) {
	set := s.controller.Settings()
	s.controller.SetSettings(set.Merge(p))
}

func (s *Service) Store() error {
	return nil
}

func (s *Service) Load() error {
	return nil
}
