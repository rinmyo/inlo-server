package service

import (
	"sync"
)

const (
	reason    = "Reason"
	event     = "Event"
	content   = "Content"
	buttons   = "Buttons"
	routeName = "Name"
)

type Route struct {
	mutex     *sync.Mutex
	alive     bool
	Id        string   `json:"id,omitempty"`
	Buttons   []string `json:"buttons,omitempty"`
	Sections  []string `json:"section,omitempty"`
	Turnouts  []string `json:"turnout,omitempty"`
	Signals   []string `json:"signals,omitempty"`
	Enemies   []string `json:"enemies,omitempty"`
	Conflicts []string `json:"conflicts,omitempty"`
}

func (r *Route) Create() {
	r.mutex.Lock()
	r.alive = true
}

func (r *Route) Kill() {
	r.alive = false
	r.mutex.Unlock()
}
