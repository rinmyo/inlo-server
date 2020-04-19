package config

import (
	"bytes"
	"github.com/pelletier/go-toml"
	"text/template"
)

var lang *toml.Tree

func logLevelErrMsg(level string) string {
	msg := lang.Get("log_level_error").(string)
	var doc bytes.Buffer
	temp, _ := template.New("log_level_error").Parse(msg)
	_ = temp.Execute(&doc, level)
	return doc.String()
}

func logLevelMsg() string {
	return lang.Get("log_level").(string)
}

func logLevelUnsetMsg() string {
	return lang.Get("log_level_unset").(string)
}

func LoadingRshMsg() string {
	return lang.Get("loading_rensahyou").(string)
}

func OpenRshFailMsg() string {
	return lang.Get("open_rensahyou_failed").(string)
}

func ReadRshFailMsg() string {
	return lang.Get("read_rensahyou_failed").(string)
}

func CloseRshFailMsg() string {
	return lang.Get("close_rensahyou_failed").(string)
}

func ParseRshFailMsg() string {
	return lang.Get("parse_rensahyou_failed").(string)
}
