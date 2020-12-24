package service

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"pracserver/src/pb"
	"sync"
	"time"
)

const ioPath = "./resource/Io.json"

type StationController interface {
	GetIOInfo() map[string][]string
	GetSectionStatus(id string) pb.Section_SectionState
	GetSignalStatus(id string) pb.Signal_SignalState
	GetTurnoutStatus(id string) pb.Turnout_TurnoutState
	UpdateSectionStatus(section *pb.Section)
	UpdateSignalStatus(signal *pb.Signal)
	UpdateTurnoutStatus(turnout *pb.Turnout)
}

type SimulatedController struct {
	mutex    map[string]*sync.Mutex
	io       map[string]map[string]string
	sections map[string]pb.Section_SectionState
	turnouts map[string]pb.Turnout_TurnoutState
	signals  map[string]pb.Signal_SignalState
}

func signalStateFromString(s string) pb.Signal_SignalState {
	switch s {
	case "H":
		return pb.Signal_RED
	case "A":
		return pb.Signal_BLUE
	}
	return pb.Signal_UNKNOWN
}

func sectionStateFromString(s string) pb.Section_SectionState {
	switch s {
	case "FREE":
		return pb.Section_FREE
	case "OCCUPIED":
		return pb.Section_OCCUPIED
	}
	return pb.Section_UNKNOWN
}

func turnoutStateFromString(s string) pb.Turnout_TurnoutState {
	switch s {
	case "NORMAL":
		return pb.Turnout_NORMAL
	case "REVERSED":
		return pb.Turnout_REVERSED
	}
	return pb.Turnout_UNKNOWN
}

func NewSimulatedController() *SimulatedController {
	var io map[string]map[string]string
	bytes, err := ioutil.ReadFile(ioPath)
	if err != nil {
		log.WithField(content, "interlock").Fatal(msg.ReadFileFailMsg)
		return nil
	}

	err = json.Unmarshal(bytes, &io)
	if err != nil {
		log.WithField(content, "interlock").Fatal(msg.ParseFileFailMsg)
		return nil
	}

	sections := make(map[string]pb.Section_SectionState)
	turnouts := make(map[string]pb.Turnout_TurnoutState)
	signals := make(map[string]pb.Signal_SignalState)

	for k, v := range io["sections"] {
		sections[k] = sectionStateFromString(v)
	}

	for k, v := range io["turnouts"] {
		turnouts[k] = turnoutStateFromString(v)
	}

	for k, v := range io["signals"] {
		signals[k] = signalStateFromString(v)
	}

	mutex := map[string]*sync.Mutex{
		"sections": {},
		"turnouts": {},
		"signals":  {},
	}
	return &SimulatedController{mutex, io, sections, turnouts, signals}
}

func (d SimulatedController) GetIOInfo() map[string][]string {
	result := make(map[string][]string)
	for k, v := range d.io {
		for id := range v {
			result[k] = append(result[k], id)
		}
	}
	return result
}

func (d SimulatedController) GetSectionStatus(id string) pb.Section_SectionState {
	d.mutex["sections"].Lock()
	defer d.mutex["sections"].Unlock()
	result := d.sections[id]
	return result
}

func (d SimulatedController) GetSignalStatus(id string) pb.Signal_SignalState {
	d.mutex["signals"].Lock()
	defer d.mutex["signals"].Unlock()
	result := d.signals[id]
	return result
}

func (d SimulatedController) GetTurnoutStatus(id string) pb.Turnout_TurnoutState {
	d.mutex["turnouts"].Lock()
	defer d.mutex["turnouts"].Unlock()
	result := d.turnouts[id]
	return result
}

func (d SimulatedController) UpdateSectionStatus(section *pb.Section) {
	d.mutex["sections"].Lock()
	defer d.mutex["sections"].Unlock()
	d.sections[section.Id] = section.State
}

func (d SimulatedController) UpdateSignalStatus(signal *pb.Signal) {
	log.Debug("開始設定號誌機")
	time.Sleep(2 * time.Second)
	d.mutex["signals"].Lock()
	defer d.mutex["signals"].Unlock()
	d.signals[signal.Id] = signal.State
	log.Debug("號誌機已設定: ", signal.State)
}

func (d SimulatedController) UpdateTurnoutStatus(turnout *pb.Turnout) {
	log.Debug("開始動作道岔")
	time.Sleep(3 * time.Second)
	d.mutex["turnouts"].Lock()
	defer d.mutex["turnouts"].Unlock()
	d.turnouts[turnout.Id] = turnout.State
	log.Debug("道岔動作結束: ", turnout.State)
}
