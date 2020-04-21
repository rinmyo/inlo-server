package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Lang struct {
	Interlock   string `json:"interlock"`
	LogLvl      string `json:"log_level"`
	Lang        string `json:"language"`
	Station     string `json:"station"`
	Route       string `json:"route"`
	StartupTime string `json:"startup_time"`

	SetOptMsg           string `json:"setting_option_msg"`
	UnspecOptMsg        string `json:"unspecified_option_msg"`
	WrgOptMsg           string `json:"wrong_option_msg"`
	LoadingMsg          string `json:"loading_msg"`
	ConnMongoMsg        string `json:"connect_mongodb_msg"`
	OpenFileFailMsg     string `json:"open_file_failed_msg"`
	ReadFileFailMsg     string `json:"read_file_failed_msg"`
	CloseFileFailMsg    string `json:"close_file_failed_msg"`
	ParseFileFailMsg    string `json:"parse_file_failed_msg"`
	FoundRouteFailMsg   string `json:"found_route_failed_msg"`
	LivingRouteExistMsg string `json:"living_route_exist_msg"`
	EnemyRouteExistMsg  string `json:"enemy_route_exist_msg"`
}

var defaultLang, _ = NewLang(defaultLangCode)

func NewLang(langCode string) (*Lang, error) {
	var lang Lang
	completePath := i18nPath + langCode + ".json"
	if jsonFile, err := os.Open(completePath); err != nil {
		return nil, errors.New(replace(Msg.OpenFileFailMsg, completePath))
	} else if byteValue, err := ioutil.ReadAll(jsonFile); err != nil {
		return nil, errors.New(replace(Msg.ReadFileFailMsg, completePath))
	} else if err = jsonFile.Close(); err != nil {
		return nil, errors.New(replace(Msg.CloseFileFailMsg, completePath))
	} else if err := json.Unmarshal(byteValue, &lang); err != nil {
		return nil, errors.New(replace(Msg.ParseFileFailMsg, completePath))
	}
	return &lang, nil
}
