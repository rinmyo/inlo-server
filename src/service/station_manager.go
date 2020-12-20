package service

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"pracserver/src/pb"
)

const (
	SectionType = iota
	SignalType
	TurnoutType
)

type StatusType uint
type Status int32

type StatusChangedEvent struct {
	st  StatusType
	id  string
	old Status
	new Status
}

func NewStatusChangedEvent(st StatusType, id string, old Status, new Status) *StatusChangedEvent {
	return &StatusChangedEvent{st, id, old, new}
}

type StationManager struct {
	sections   map[string]pb.Section_SectionState
	turnouts   map[string]pb.Turnout_TurnoutState
	signals    map[string]pb.Signal_SignalState
	interlock  map[string]*Route
	routePool  map[string]*Route
	channel    chan StatusChangedEvent
	controller *StationController
}

func NewStationManager(sp *StationController, interlockPath string) *StationManager {
	sections := make(map[string]pb.Section_SectionState)
	turnouts := make(map[string]pb.Turnout_TurnoutState)
	signals := make(map[string]pb.Signal_SignalState)
	routePool := make(map[string]*Route)
	interlock := make(map[string]*Route)
	loadRoute(interlock, interlockPath)
	channel := make(chan StatusChangedEvent)
	return &StationManager{sections, turnouts, signals, interlock, routePool, channel, sp}
}

//OnStatusChange 當車站狀態變化時
func (m *StationManager) OnStatusChange(handle func(StatusChangedEvent) error) {
	for {
		e := <-m.channel
		err := handle(e)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (m *StationManager) UpdateTurnoutStatus(id string, new pb.Turnout_TurnoutState) {
	old := m.turnouts[id]
	m.turnouts[id] = new
	e := *NewStatusChangedEvent(TurnoutType, id, Status(old), Status(new))
	select {
	case m.channel <- e:
	default:
	}
}

func (m *StationManager) UpdateSectionStatus(id string, new pb.Section_SectionState) {
	old := m.sections[id]
	m.sections[id] = new
	e := *NewStatusChangedEvent(SectionType, id, Status(old), Status(new))
	select {
	case m.channel <- e:
	default:
	}
}

func (m *StationManager) UpdateSignalStatus(id string, new pb.Signal_SignalState) {
	old := m.signals[id]
	m.signals[id] = new
	e := *NewStatusChangedEvent(SignalType, id, Status(old), Status(new))
	select {
	case m.channel <- e:
	default:
	}
}

// CreateRoute create a new route
func (m *StationManager) CreateRoute(r *Route) error {
	errMsg := make(map[string]interface{})

	//存在未取消的相同的进路
	if m.IsLiving(r) {
		errMsg["exist living route"] = r.Id
	}

	//存在敌对进路
	result := m.LivingEnemies(r)
	var enemiesId []string
	for _, e := range result {
		enemiesId = append(enemiesId, e.Id)
	}
	if len(enemiesId) > 0 {
		errMsg["exist living enemies"] = enemiesId
	}

	//存在牴觸進路
	result = m.LivingConflicts(r)
	var conflictsId []string
	for _, e := range result {
		conflictsId = append(conflictsId, e.Id)
	}
	if len(conflictsId) > 0 {
		errMsg["exist living conflicts"] = conflictsId
	}

	//軌道電路是否空閒
	var occupiedSections []string
	for _, id := range r.Sections {
		if m.sections[id] != pb.Section_FREE {
			occupiedSections = append(occupiedSections, id)
		}
	}

	if len(occupiedSections) > 0 {
		errMsg["sections not free"] = occupiedSections
	}

	if len(errMsg) > 0 {
		str, err := json.Marshal(errMsg)
		if err != nil {
			log.Fatal(err)
		}
		return status.Error(codes.Internal, string(str[:]))
	}
	return nil
}

//CancelRoute 取消一條進路
func (m *StationManager) CancelRoute(r *Route) {

	delete(m.routePool, r.Id)
	log.WithField("route", r.Id).
		WithField(event, msg.CancelEvent).
		Info(msg.RouteEventMsg)
}

//IsLiving 檢測是否存在進路
func (m *StationManager) IsLiving(r *Route) bool {
	_, ok := m.routePool[r.Id]
	return ok
}

//LivingEnemies 檢測敵對進路
func (m *StationManager) LivingEnemies(r *Route) (result []*Route) {
	for _, v := range r.Enemies {
		if route, ok := m.routePool[v]; ok {
			result = append(result, route)
		}
	}
	return
}

//LivingConflicts 檢測牴觸進路
func (m *StationManager) LivingConflicts(r *Route) (result []*Route) {
	for _, v := range r.Conflicts {
		if route, ok := m.routePool[v]; ok {
			result = append(result, route)
		}
	}
	return
}

func (m *StationManager) RefreshStationStatus() {
	go func() {
		c := *m.controller
		for {
			for _, id := range c.GetIOInfo()["turnouts"] {
				newState := c.GetTurnoutStatus(id)
				if oldState, ok := m.turnouts[id]; ok {
					if oldState != newState {
						m.UpdateTurnoutStatus(id, newState)
					}
				} else {
					m.turnouts[id] = newState
				}
			}
			for _, id := range c.GetIOInfo()["sections"] {
				newState := c.GetSectionStatus(id)
				if oldState, ok := m.sections[id]; ok {
					if oldState != newState {
						m.UpdateSectionStatus(id, newState)
					}
				} else {
					m.sections[id] = newState
				}
			}
			for _, id := range c.GetIOInfo()["signals"] {
				newState := c.GetSignalStatus(id)
				if oldState, ok := m.signals[id]; ok {
					if oldState != newState {
						m.UpdateSignalStatus(id, newState)
					}
				} else {
					m.signals[id] = newState
				}
			}
		}
	}()
}

func (m *StationManager) GetRouteByName(name string) (*Route, bool) {
	val, ok := m.interlock[name]
	if !ok {
		log.WithField(reason, msg.NoSuchRouteMsg).
			WithField(routeName, name).
			Error(msg.ObtainRouteFailMsg)
	}
	return val, ok
}

func (m *StationManager) GetRouteByBtn(btns ...string) (*Route, bool) {
outer:
	for _, v := range m.interlock {
		if len(v.Buttons) == len(btns) {
			for i := 0; i < len(btns); i++ {
				if v.Buttons[i] != btns[i] {
					continue outer
				}
			}
			return v, true
		}
	}
	log.WithField(reason, msg.NoSuchRouteMsg).
		WithField(buttons, btns).
		Error(msg.ObtainRouteFailMsg)
	return nil, false
}

func loadRoute(interlock map[string]*Route, interlockRoute string) {
	log.WithField(content, "interlock").Info(msg.LoadRouteMsg)
	bytes, err := ioutil.ReadFile(interlockRoute)

	if err != nil {
		log.WithField(content, "interlock").Fatal(msg.ReadFileFailMsg)
		return
	}

	err = json.Unmarshal(bytes, &interlock)
	if err != nil {
		log.WithField(content, "interlock").Fatal(msg.ParseFileFailMsg)
		return
	}

	for k, v := range interlock {
		v.Id = k
		text, err := json.Marshal(v)
		if err != nil {
			log.Error(err)
		}
		var field log.Fields
		err = json.Unmarshal(text, &field)
		if err != nil {
			log.Error(err)
		}
		log.WithFields(field).Info(msg.LoadRouteMsg)
	}

}
