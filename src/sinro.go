package main

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"pracserver/src/config"
)

const rensahyouPath = "./resource/rensahyou.json"

var (
	//連鎖表
	rensahyou map[string]Sinro

	//進路池
	sinroPool map[string]*LivingSinro
)

type Sinro struct {
	Buttons  []string
	Sections []string
	Turnouts []string
	Signals  []string
	Enemies  []string
}

type LivingSinro struct {
	id string
	*Sinro
}

func getSinroByName(name string) (*LivingSinro, error) {
	if val, ok := rensahyou[name]; ok {
		return &LivingSinro{name, &val}, nil
	} else {
		return nil, errors.New("未找到進路： " + name)
	}
}

//TODO
//func getSinroBy2Btn(btn1, btn2 string) (*LivingSinro, error) {
//
//}
//
//func getSinroBy3Btn(btn1, btn2, btn3 string) (*LivingSinro, error) {
//
//}

func (ls LivingSinro) livingEnemies() (result []*LivingSinro) {
	for _, v := range ls.Enemies {
		if enemy, err := getSinroByName(v); err != nil {
			log.Panic(err)
		} else {
			if val, ok := sinroPool[enemy.id]; ok {
				result = append(result, val)
			}
		}
	}
	return
}

func (ls LivingSinro) hasLivingEnemies() bool {
	return len(ls.livingEnemies()) > 0
}

func (ls LivingSinro) NewSinro() error {
	if ls.hasLivingEnemies() {
		return errors.New("無法建立：" + ls.id + "因為存在敵對進路")
	}
	sinroPool["id"] = &ls
	return nil
}

func init() {
	log.Info(config.LoadingRshMsg())
	if jsonFile, err := os.Open(rensahyouPath); err != nil {
		log.Fatal(config.OpenRshFailMsg())
	} else if byteValue, err := ioutil.ReadAll(jsonFile); err != nil {
		log.Fatal(config.ReadRshFailMsg())
	} else if err = jsonFile.Close(); err != nil {
		log.Warn(config.CloseRshFailMsg())
	} else if err := json.Unmarshal(byteValue, &rensahyou); err != nil {
		log.Fatal(config.ParseRshFailMsg())
	}
}
