package service

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"pracserver/src/pb"
	"sync"
	"time"
)

const (
	SectionType = iota
	SignalType
	TurnoutType
)

type DeviceType uint
type DeviceState int32

type StateChangedEvent struct {
	st  DeviceType
	id  string
	old DeviceState
	new DeviceState
}

func NewStatusChangedEvent(st DeviceType, id string, old DeviceState, new DeviceState) *StateChangedEvent {
	return &StateChangedEvent{st, id, old, new}
}

type StationManager struct {
	sections   map[string]pb.Section_SectionState //當前軌道區段狀態
	turnouts   map[string]pb.Turnout_TurnoutState //當前道岔狀態
	signals    map[string]pb.Signal_SignalState   //當前信號狀態
	interlock  map[string]*Route                  //連鎖表
	channel    chan *StateChangedEvent            //通知 channel
	controller *StationController                 //車站控制器
}

func NewStationManager(sp *StationController, interlockPath string) *StationManager {
	sections := make(map[string]pb.Section_SectionState)
	turnouts := make(map[string]pb.Turnout_TurnoutState)
	signals := make(map[string]pb.Signal_SignalState)
	interlock := make(map[string]*Route)
	ioInfo := (*sp).GetIOInfo()
	channel := make(chan *StateChangedEvent, len(ioInfo["turnouts"])+len(ioInfo["sections"])+len(ioInfo["signals"]))
	loadRoute(interlock, interlockPath)
	return &StationManager{sections, turnouts, signals, interlock, channel, sp}
}

// SetTurnouts 設置多個道岔
func (m *StationManager) SetTurnouts(ts []*pb.Turnout) error {
	c := *m.controller
	errMsg := make(map[string]interface{})
	turnouts := make(map[*pb.Turnout]pb.Turnout_TurnoutState)
	for _, t := range ts {
		turnouts[t] = c.GetTurnoutStatus(t.Id)
		go func(t *pb.Turnout) {
			c.UpdateTurnoutStatus(t)
		}(t)
	}
	for timeout := time.After(5 * time.Second); ; {
		select {
		case <-timeout:
			for t, oldState := range turnouts {
				m.turnouts[t.Id] = pb.Turnout_BROKEN
				m.channel <- NewStatusChangedEvent(TurnoutType, t.Id, DeviceState(oldState), DeviceState(pb.Turnout_BROKEN))
				errMsg["turnout action timeout"] = t
				t.State = oldState
				c.UpdateTurnoutStatus(t)
			}
			str, _ := json.Marshal(errMsg)
			return status.Error(codes.Internal, string(str[:]))
		default:
			for t := range turnouts {
				if c.GetTurnoutStatus(t.Id) == t.State {
					delete(turnouts, t)
				}
				if len(turnouts) == 0 {
					return nil
				}
			}
		}
	}
}

// CreateRoute create a new route
func (m *StationManager) CreateRoute(r *Route) error {
	//存在未取消的相同的进路
	if r.alive {
		return status.Error(codes.Internal, "exist living route: "+r.Id)
	}

	errMsg := make(map[string]interface{})
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

	// 初步檢查，有錯直接返回
	if str, _ := json.Marshal(errMsg); len(errMsg) > 0 {
		return status.Error(codes.Internal, string(str[:]))
	}

	var turnouts []*pb.Turnout
	for _, v := range r.Turnouts {
		ts, err := ParseTurnout(v)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range ts {
			turnouts = append(turnouts, t)
		}
	}

	//設置道岔
	err := m.SetTurnouts(turnouts)
	if err != nil {
		return err
	}

	c := *m.controller
	for _, v := range r.Sections {
		s := ParseSection(v, pb.Section_LOCKED)
		go func() {
			c.UpdateSectionStatus(s)
		}()
	}

	for _, v := range r.Signals {
		s := ParseSignal(v)
		signal := c.GetSignalStatus(s.Id)
		//如果信號機損壞或未知
		if signal == pb.Signal_BROKEN || signal == pb.Signal_UNKNOWN {
			errMsg["signal state abnormal"] = signal
		}
		go func() {
			c.UpdateSignalStatus(s)
		}()
	}

	//信號機是否有錯
	if str, _ := json.Marshal(errMsg); len(errMsg) > 0 {
		return status.Error(codes.Internal, string(str[:]))
	}

	m.interlock[r.Id].Create()
	log.Info("route has been created: ", r.Id)
	return nil
}

//CancelRoute 取消一條進路
func (m *StationManager) CancelRoute(r *Route) error {
	//该进路不存在
	if !r.alive {
		return status.Error(codes.Internal, "not found living route:"+r.Id)
	}

	errMsg := make(map[string]interface{})
	//軌道電路是否空閒
	var occupiedSections []string
	for _, id := range r.Sections {
		if m.sections[id] != pb.Section_LOCKED {
			occupiedSections = append(occupiedSections, id)
		}
	}
	if len(occupiedSections) > 0 {
		errMsg["sections not locked"] = occupiedSections
	}
	// 初步檢查，有錯直接返回
	if str, _ := json.Marshal(errMsg); len(errMsg) > 0 {
		return status.Error(codes.Internal, string(str[:]))
	}
	c := *m.controller
	//無錯 先關閉信號
	for _, v := range r.Signals {
		s := ParseAbortSignal(v)
		signal := c.GetSignalStatus(s.Id)
		//如果信號機損壞或未知
		if signal == pb.Signal_BROKEN || signal == pb.Signal_UNKNOWN {
			errMsg["signal state abnormal"] = signal
		}
		go func() {
			c.UpdateSignalStatus(s)
		}()
	}

	//接着 動作道岔
	var turnouts []*pb.Turnout
	for _, v := range r.Turnouts {
		ts, err := ParseNormalTurnout(v)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range ts {
			turnouts = append(turnouts, t)
		}
	}
	err := m.SetTurnouts(turnouts)
	if err != nil {
		return err
	}

	//最後設定軌道電路
	for _, v := range r.Sections {
		s := ParseSection(v, pb.Section_FREE)
		go func() {
			c.UpdateSectionStatus(s)
		}()
	}

	r.Kill()
	return nil
}

func (m *StationManager) AliveRoutes() (result []*Route) {
	for _, v := range m.interlock {
		if v.alive {
			result = append(result, v)
		}
	}
	return
}

//LivingEnemies 檢測敵對進路
func (m *StationManager) LivingEnemies(r *Route) (result []*Route) {
	for _, v := range r.Enemies {
		if route, ok := m.interlock[v]; ok {
			if route.alive {
				result = append(result, route)
			}
		} else {
			log.Error("so such route: ", v)
		}
	}
	return
}

//LivingConflicts 檢測牴觸進路
func (m *StationManager) LivingConflicts(r *Route) (result []*Route) {
	for _, v := range r.Conflicts {
		if route, ok := m.interlock[v]; ok {
			if route.alive {
				result = append(result, route)
			}
		} else {
			log.Error("so such route: ", v)
		}
	}
	return
}

//repair 嘗試修復設備
func (m *StationManager) repair(dt DeviceType, id string) {
	c := *m.controller
	switch dt {
	case SignalType:
		if m.signals[id] == pb.Signal_BROKEN {
			m.signals[id] = c.GetSignalStatus(id)
		}
	case TurnoutType:
		if m.turnouts[id] == pb.Turnout_BROKEN {
			m.turnouts[id] = c.GetTurnoutStatus(id)
		}
	case SectionType:
		if m.sections[id] == pb.Section_BROKEN {
			m.sections[id] = c.GetSectionStatus(id)
		}
	}
}

// RefreshStationStatus 刷新車站狀態
func (m *StationManager) RefreshStationStatus() {
	c := *m.controller
	for _, id := range c.GetIOInfo()["turnouts"] {
		newState := c.GetTurnoutStatus(id)
		if oldState, ok := m.turnouts[id]; ok {
			if oldState != pb.Turnout_BROKEN && oldState != newState {
				m.turnouts[id] = newState
				select {
				case m.channel <- NewStatusChangedEvent(TurnoutType, id, DeviceState(oldState), DeviceState(newState)):
				default:
				}
			}
		} else {
			m.turnouts[id] = newState
		}
	}
	for _, id := range c.GetIOInfo()["sections"] {
		newState := c.GetSectionStatus(id)
		if oldState, ok := m.sections[id]; ok {
			if oldState != pb.Section_BROKEN && oldState != newState {
				m.sections[id] = newState
				select {
				case m.channel <- NewStatusChangedEvent(SectionType, id, DeviceState(oldState), DeviceState(newState)):
				default:
				}
			}
		} else {
			m.sections[id] = newState
		}
	}

	for _, id := range c.GetIOInfo()["signals"] {
		newState := c.GetSignalStatus(id)
		if oldState, ok := m.signals[id]; ok {
			if oldState != pb.Signal_BROKEN && oldState != newState {
				m.signals[id] = newState
				select {
				case m.channel <- NewStatusChangedEvent(SignalType, id, DeviceState(oldState), DeviceState(newState)):
				default:
				}
			}
		} else {
			m.signals[id] = newState
		}
	}
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
		v.alive = false
		v.mutex = &sync.Mutex{}

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
