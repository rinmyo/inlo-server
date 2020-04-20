package config

import (
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"github.com/weekface/mgorus"
	"pracserver/src/tool"
)

const (
	defaultLevel        = log.DebugLevel
	defaultLangCode     = "zh-TW"
	defaultMongoHost    = "localhost"
	defaultMongoPort    = "27017"
	defaultMongoDB      = "pracserver"
	serverLogCollection = "server_log"

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

func mongoHooker() (log.Hook, error) {
	var host, port, db interface{}
	if host = config.Get(mongoHost); host == nil {
		host = defaultMongoHost
	}

	if port = config.Get(mongoPort); port == nil {
		port = defaultMongoPort
	}

	if db = config.Get(mongoDatabase); db == nil {
		db = defaultMongoDB
	}

	if hooker, err := mgorus.NewHooker(host.(string)+":"+port.(string), db.(string), serverLogCollection); err != nil {
		return nil, err
	} else {
		return hooker, nil
	}
}

func setMongoHooker(hooker log.Hook) {
	log.AddHook(hooker)
}

func setConfig(cfg *toml.Tree) {
	config = cfg
}

func init() {
	setLanguage(defaultLang)
	setLogLevel(defaultLevel)

	if cfg, err := toml.LoadFile(configPath); err != nil {
		log.Fatal(loadFailedMsg)
	} else {
		setConfig(cfg)

		if hooker, err := mongoHooker(); err != nil {
			log.Error(err)
		} else {
			setMongoHooker(hooker)
		}

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
