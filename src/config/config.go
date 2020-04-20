package config

import (
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"pracserver/src/tool"
)

const (
	defaultLevel    = log.DebugLevel
	defaultLangCode = "zh-TW"

	configPath = "./resource/config.toml"
	i18nPath   = "./resource/i18n/"

	loadFailedMsg = "loading configuration failed!"
)

var (
	config *toml.Tree

	replace = tool.Replace
)

func StationName() string {
	return config.Get(stationName).(string)
}

func logLevel() log.Level {
	if lvl := config.Get(serverLogLevel); lvl == nil {
		log.Warn(replace(Msg.UnspecOptMsg, Msg.LogLvl, defaultLevel.String()))
		return defaultLevel
	} else {
		if v, err := log.ParseLevel(lvl.(string)); err != nil {
			log.Error(replace(Msg.WrgOptMsg, lvl.(string), Msg.LogLvl, defaultLevel.String()))
			return defaultLevel
		} else {
			return v
		}
	}
}

func setLogLevel(level log.Level) {
	log.SetLevel(level)
}

func language() (*Lang, string) {
	if v := config.Get(userLang); v != nil {
		if lang, err := NewLang(v.(string)); err != nil {
			log.Error(replace(Msg.WrgOptMsg, v.(string), Msg.Lang, defaultLangCode))
			return defaultLang, defaultLangCode
		} else {
			return lang, v.(string)
		}
	} else {
		log.Warn(replace(Msg.UnspecOptMsg, Msg.Lang, defaultLangCode))
		return defaultLang, defaultLangCode
	}
}

func setLanguage(lang *Lang) {
	Msg = lang
}

func init() {
	setLanguage(defaultLang)
	setLogLevel(defaultLevel)

	if cfg, err := toml.LoadFile(configPath); err != nil {
		log.Fatal(loadFailedMsg)
	} else {
		config = cfg

		lang, code := language()
		log.Info(replace(Msg.SetOptMsg, Msg.Lang, code))
		if lang != defaultLang {
			setLanguage(lang)
		}

		lvl := logLevel()
		log.Info(replace(Msg.SetOptMsg, Msg.LogLvl, lvl.String()))
		if lvl != defaultLevel {
			setLogLevel(lvl)
		}
	}
}
