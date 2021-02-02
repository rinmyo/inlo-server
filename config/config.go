package config

import (
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLevel    = log.InfoLevel
	defaultLangCode = "en-US"

	loadFailedMsg = "loading configuration failed!"
)

var (
	config *toml.Tree
	Msg    *Lang
)

func StationName() string {
	return config.Get(stationName).(string)
}

func UserId() string {
	return config.Get(userId).(string)
}

func UserPassword() string {
	return config.Get(userPassword).(string)
}

func logLevel() log.Level {
	if lvl := config.Get(serverLogLevel); lvl == nil {
		log.WithField("Option", userLang).
			Warn(Msg.UnspecOptMsg)
		return defaultLevel
	} else {
		if v, err := log.ParseLevel(lvl.(string)); err != nil {
			log.WithField("Option", serverLogLevel).
				Error(Msg.IllegalOptMsg)
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
			log.WithField("Option", userLang).
				Error(Msg.IllegalOptMsg, ":")
			log.Error(err)
			return DefaultLang(), defaultLangCode
		} else {
			return lang, v.(string)
		}
	} else {
		log.WithField("Option", userLang).
			Warn(Msg.UnspecOptMsg)
		return DefaultLang(), defaultLangCode
	}
}

func setLanguage(lang *Lang) {
	Msg = lang
}

func setConfig(cfg *toml.Tree) {
	config = cfg
}

func LoadConfig(configPath string) {
	setLanguage(DefaultLang())
	setLogLevel(defaultLevel)
	if cfg, err := toml.LoadFile(configPath); err != nil {
		log.Fatal(loadFailedMsg)
	} else {
		setConfig(cfg)

		if lang, code := language(); lang != DefaultLang() {
			setLanguage(lang)
			log.WithField(userLang, code).Info(Msg.SetOptMsg)
		}

		if lvl := logLevel(); lvl != defaultLevel {
			setLogLevel(lvl)
			log.WithField(serverLogLevel, lvl.String()).Info(Msg.SetOptMsg)
		}

		log.WithField("Station", StationName()).Info(Msg.LoadRouteMsg)
	}
}
