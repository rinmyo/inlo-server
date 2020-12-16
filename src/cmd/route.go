package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"pracserver/src/config"
)

const (
	interlockPath = "./resource/interlock.json"

	reason    = "Reason"
	event     = "Event"
	route     = "Route"
	content   = "Content"
	interlock = "interlock information"
	buttons   = "Buttons"
	routeName = "Name"
)

var (
	//連鎖表
	interlockTable map[string]*Route
	routePool      = make(map[string]*Route)

	msg = config.Msg
)

type Route struct {
	Id        string
	Buttons   []string
	Section   []string
	Turnout   []string
	Signals   []string
	Enemies   []string
	Conflicts []string
}

func (r *Route) Found() (ok bool) {
	ok = true
	logFields := log.WithField(route, r.Id)

	//存在未取消的相同的进路
	if r.IsLiving() {
		logFields = logFields.WithField(reason, msg.LivingRouteExistMsg)
		ok = false
	}

	//存在敌对进路
	if r.HasLivingEnemies() {
		logFields = logFields.WithField(reason, msg.EnemyRouteExistMsg)
		ok = false
	}

	//存在牴觸進路
	if r.HasLivingConflicts() {
		logFields = logFields.WithField(reason, "route conflict")
		ok = false
	}

	//軌道電路是否空閒
	for _, section := range r.Section {
		if ReadSectionState(section) != SectionFree {
			logFields = logFields.WithField(reason, msg.SectionNotFreeMsg)
			ok = false
		}
	}

	channels := make(map[string]chan bool)

	//調用硬件電路，開啓道岔至相應位置
	for _, turnout := range r.Turnout {
		if turnouts, err := ParseTurnout(turnout); err == nil {
			for _, turnout := range turnouts {
				channel := make(chan bool)
				channels[turnout.Tid] = channel
				go UpdateTurnoutState(turnout, channel)
				//延遲三秒以模擬道岔動作的延遲
				//time.Sleep(time.Second * 3)
			}
		}
	}

outer:
	for {
		for tid, channel := range channels {
			if <-channel {
				delete(channels, tid)
			} else {
				log.WithField("Turnout", tid).
					Error("turnout cannot be set")
			}
		}
		if len(channels) == 0 {
			log.Info("All turnout are set")
			break outer
		}
	}

	if ok {
		routePool[r.Id] = r
		logFields = logFields.WithField(event, msg.FoundEvent)
		logFields.Info(msg.RouteEventMsg)
	} else {
		logFields.Error(msg.FoundRouteFailMsg)
	}
	return
}

func (r *Route) Cancel() {
	delete(routePool, r.Id)
	log.WithField(route, r.Id).
		WithField(event, msg.CancelEvent).
		Info(msg.RouteEventMsg)
}

func (r *Route) IsLiving() bool {
	_, ok := routePool[r.Id]
	return ok
}

func (r *Route) LivingEnemies() (result []*Route) {
	for _, v := range r.Enemies {
		if route, ok := routePool[v]; ok {
			result = append(result, route)
		}
	}
	return
}

func (r *Route) HasLivingEnemies() bool {
	return len(r.LivingEnemies()) > 0
}

func (r *Route) LivingConflicts() (result []*Route) {
	for _, v := range r.Conflicts {
		if route, ok := routePool[v]; ok {
			result = append(result, route)
		}
	}
	return
}

func (r *Route) HasLivingConflicts() bool {
	return len(r.LivingConflicts()) > 0
}

func GetRouteByName(name string) (*Route, bool) {
	val, ok := interlockTable[name]
	if !ok {
		log.WithField(reason, msg.NoSuchRouteMsg).
			WithField(routeName, name).
			Error(msg.ObtainRouteFailMsg)
	}
	return val, ok
}

func GetRouteByBtn(btns ...string) (*Route, bool) {
outer:
	for _, v := range interlockTable {
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

func loadRoute() {
	log.WithField(content, interlock).Info(msg.LoadRouteMsg)
	if jsonFile, err := os.Open(interlockPath); err != nil {
		log.WithField(content, interlock).Fatal(msg.OpenFileFailMsg)
	} else if byteValue, err := ioutil.ReadAll(jsonFile); err != nil {
		log.WithField(content, interlock).Fatal(msg.ReadFileFailMsg)
	} else if err = jsonFile.Close(); err != nil {
		log.WithField(content, interlock).Fatal(msg.CloseFileFailMsg)
	} else if err := json.Unmarshal(byteValue, &interlockTable); err != nil {
		log.WithField(content, interlock).Fatal(msg.ParseFileFailMsg)
	} else {
		for k, v := range interlockTable {
			v.Id = k
			log.WithField(route, k).Info(msg.LoadRouteMsg)
		}
	}
}

func init() {
	loadRoute()
}
