package config

import (
	"github.com/pelletier/go-toml"
	sys "github.com/shirou/gopsutil/host"
	log "github.com/sirupsen/logrus"
	"github.com/weekface/mgorus"
	"pracserver/src/tool"
	"time"
)

const (
	defaultLevel     = log.InfoLevel
	defaultLangCode  = "en-US"
	defaultMongoHost = "localhost"
	defaultMongoPort = "27017"
	defaultLogDb     = "prac_log"

	configPath = "./resource/config.toml"
	i18nPath   = "./resource/i18n/"

	logFormat     = "log_20060102150405"
	loadFailedMsg = "loading configuration failed!"
)

var (
	config *toml.Tree
	Msg    *Lang

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

func mongoHook() (log.Hook, string, string, error) {
	var host, port, db interface{}
	if host = config.Get(mongoHost); host == nil {
		host = defaultMongoHost
	}

	if port = config.Get(mongoPort); port == nil {
		port = defaultMongoPort
	}

	if db = config.Get(mongoDatabase); db == nil {
		db = defaultLogDb
	}

	url := host.(string) + ":" + port.(string)
	col := time.Now().Format(logFormat)

	if hooker, err := mgorus.NewHooker(url, db.(string), col); err != nil {
		return nil, "", "", err
	} else {
		return hooker, url, db.(string), nil
	}
}

func setMongoHook(hook log.Hook) {
	log.AddHook(hook)
}

func setConfig(cfg *toml.Tree) {
	config = cfg
}

func loadConfig() {
	setLanguage(defaultLang)
	setLogLevel(defaultLevel)
	if cfg, err := toml.LoadFile(configPath); err != nil {
		log.Fatal(loadFailedMsg)
	} else {
		setConfig(cfg)

		if lang, code := language(); lang != defaultLang {
			setLanguage(lang)
			log.Info(replace(Msg.SetOptMsg, Msg.Lang, code))
		}

		if lvl := logLevel(); lvl != defaultLevel {
			setLogLevel(lvl)
			log.Info(replace(Msg.SetOptMsg, Msg.LogLvl, lvl.String()))
		}

		if hook, url, db, err := mongoHook(); err != nil {
			log.Error(err)
		} else {
			setMongoHook(hook)
			log.Info(replace(Msg.ConnMongoMsg, url, db))
		}

		log.Info(Msg.Station, StationName())
		logEnvInfo()
	}
}

func logEnvInfo() {
	sysInfo, _ := sys.Info()
	log.Debug(sysInfo.String())
}

func init() {
	loadConfig()
}
