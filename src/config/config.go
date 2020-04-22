package config

import (
	"github.com/pelletier/go-toml"
	sys "github.com/shirou/gopsutil/host"
	log "github.com/sirupsen/logrus"
	"github.com/weekface/mgorus"
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
)

func StationName() string {
	return config.Get(stationName).(string)
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
			return defaultLang, defaultLangCode
		} else {
			return lang, v.(string)
		}
	} else {
		log.WithField("Option", userLang).
			Warn(Msg.UnspecOptMsg)
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
			log.WithField(userLang, code).Info(Msg.SetOptMsg)
		}

		if lvl := logLevel(); lvl != defaultLevel {
			setLogLevel(lvl)
			log.WithField(serverLogLevel, lvl.String()).Info(Msg.SetOptMsg)
		}

		if hook, url, db, err := mongoHook(); err != nil {
			log.WithField("Url", url).
				WithField("Database", db).
				Error(err)
		} else {
			setMongoHook(hook)
			log.WithField("Url", url).
				WithField("Database", db).
				Info(Msg.ConnMsg)
		}

		log.WithField("Station", StationName()).Info(Msg.LoadMsg)
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
