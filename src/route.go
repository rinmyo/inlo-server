package main

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"pracserver/src/config"
	"pracserver/src/tool"
	"time"
)

const (
	interlockPath = "./resource/interlock.json"
)

var (
	//連鎖表
	interlockTable map[string]*Route
	//進路池
	routePool []*LivingRoute

	msg     = config.Msg
	replace = tool.Replace
)

type Route struct {
	Id      string
	Buttons []string
	Section []string
	Turnout []string
	Signals []string
	Enemies []string
}

func (r *Route) setId(id string) {
	r.Id = id
}

func (r *Route) found() error {
	if r.hasLivingEnemies() {
		return errors.New("無法建立：" + r.Id + "因為存在敵對進路")
	}
	routePool = append(routePool, r.NewLivingRoute())
	return nil
}

func (r *Route) NewLivingRoute() *LivingRoute {
	return &LivingRoute{foundTime: time.Now(), Route: r}
}

func (r *Route) livingEnemies() (result []*LivingRoute) {
	for _, v := range r.Enemies {
		for _, s := range routePool {
			if s.GetRoute() == getRouteByName(v) {
				result = append(result, s)
			}
		}
	}
	return
}

func (r *Route) hasLivingEnemies() bool {
	return len(r.livingEnemies()) > 0
}

type LivingRoute struct {
	foundTime  time.Time
	cancelTime time.Time
	*Route
}

func (lr *LivingRoute) GetRoute() *Route {
	return lr.Route
}

func getRouteByName(name string) *Route {
	if val, ok := interlockTable[name]; ok {
		return val
	} else {
		return nil
	}
}

func getRouteByBtn(btns ...string) (*Route, error) {
outer:
	for _, v := range interlockTable {
		if len(v.Buttons) == len(btns) {
			for i := 0; i < len(btns); i++ {
				if v.Buttons[i] != btns[i] {
					continue outer
				}
			}
			return v, nil
		}
	}
	return nil, errors.New("no such route")
}

func LoadRoute() {
	log.Info(replace(msg.LoadingMsg, msg.Interlock))
	if jsonFile, err := os.Open(interlockPath); err != nil {
		log.Fatal(replace(msg.OpenFileFailMsg, msg.Interlock))
	} else if byteValue, err := ioutil.ReadAll(jsonFile); err != nil {
		log.Fatal(replace(msg.ReadFileFailMsg, msg.Interlock))
	} else if err = jsonFile.Close(); err != nil {
		log.Warn(replace(msg.CloseFileFailMsg, msg.Interlock))
	} else if err := json.Unmarshal(byteValue, &interlockTable); err != nil {
		log.Fatal(replace(msg.ParseFileFailMsg, msg.Interlock))
	} else {
		for k, v := range interlockTable {
			v.setId(k)
			log.Info(replace(msg.LoadingMsg, msg.Route+" "+k))
		}
	}
}

func init() {
	LoadRoute()
}
