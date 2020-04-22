package interlock

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
	Id      string
	Buttons []string
	Section []string
	Turnout []string
	Signals []string
	Enemies []string
}

func (r *Route) Found() bool {
	//检测是否存在未取消的相同的进路
	if r.IsLiving() {
		log.WithField(route, r.Id).
			WithField(reason, msg.LivingRouteExistMsg).
			Error(msg.FoundRouteFailMsg)
		return false
	}

	//是否存在敌对进路
	if r.HasLivingEnemies() {
		log.WithField(route, r.Id).
			WithField(reason, msg.EnemyRouteExistMsg).
			Error(msg.FoundRouteFailMsg)
		return false
	}

	////////////////////////////////
	//todo:以下检测项目需要硬件接口，怎么办？
	////////////////////////////////
	//todo:轨道电路是否空闲
	//todo:道岔是否在位且缩闭
	//todo:信号机是否开放
	////////////////////////////////

	routePool[r.Id] = r
	log.WithField(route, r.Id).
		WithField(event, msg.FoundEvent).
		Info(msg.RouteEventMsg)
	return true
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
	log.WithField(content, interlock).Info(msg.LoadMsg)
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
			log.WithField(route, k).Info(msg.LoadMsg)
		}
	}
}

func init() {
	loadRoute()
}
