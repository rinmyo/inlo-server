package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var Msg *Lang

type Lang struct {
	Rsh    string `json:"rensahyou"`
	LogLvl string `json:"log_level"`
	Lang   string `json:"language"`

	SetOptMsg        string `json:"setting_option_msg"`
	UnspecOptMsg     string `json:"unspecified_option_msg"`
	WrgOptMsg        string `json:"wrong_option_msg"`
	LoadingFileMsg   string `json:"loading_file_msg"`
	OpenFileFailMsg  string `json:"open_file_failed_msg"`
	ReadFileFailMsg  string `json:"read_file_failed_msg"`
	CloseFileFailMsg string `json:"close_file_failed_msg"`
	ParseFileFailMsg string `json:"parse_file_failed_msg"`
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
