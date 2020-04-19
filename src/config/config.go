package config

import (
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLevel = log.DebugLevel
	defaultLang  = "zh-TW"

	configPath = "./resource/config.toml"
	i18nPath   = "./resource/i18n/"

	loadFailedMsg = "loading configuration failed!"
	failLangMsg   = "unknown language: "
	userLangMsg   = "user language: "
)

var (
	config *toml.Tree
)

func StationName() string {
	return config.Get(stationName).(string)
}

func setLogLevel() {
	level := defaultLevel
	if lvl := config.Get(logLevel); lvl == nil {
		log.Warn(logLevelUnsetMsg())
	} else {
		if v, err := log.ParseLevel(lvl.(string)); err != nil {
			log.Warn(logLevelErrMsg(lvl.(string)))
		} else {
			level = v
		}
	}
	log.Info(logLevelMsg(), level)
	log.SetLevel(level)
}

func language() string {
	if v := config.Get(userLang); v != nil {
		return v.(string)
	} else {
		return defaultLang
	}
}

func setLanguage() {
	if lan, err := toml.LoadFile(i18nPath + language() + ".toml"); err != nil {
		log.Fatal(failLangMsg, language())
	} else {
		lang = lan
		log.Info(userLangMsg, language())
	}
}

func init() {
	if cfg, err := toml.LoadFile(configPath); err != nil {
		log.Fatal(loadFailedMsg)
	} else {
		config = cfg
		setLanguage()
		setLogLevel()
	}
}
