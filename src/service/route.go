package service

import (
	"pracserver/src/config"
)

const (
	reason    = "Reason"
	event     = "Event"
	content   = "Content"
	buttons   = "Buttons"
	routeName = "Name"
)

var (
	msg = config.Msg
)

type Route struct {
	Id        string   `json:"id,omitempty"`
	Buttons   []string `json:"buttons,omitempty"`
	Sections  []string `json:"section,omitempty"`
	Turnouts  []string `json:"turnout,omitempty"`
	Signals   []string `json:"signals,omitempty"`
	Enemies   []string `json:"enemies,omitempty"`
	Conflicts []string `json:"conflicts,omitempty"`
}
