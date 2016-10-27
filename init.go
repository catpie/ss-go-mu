package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

func InitConfig() error {
	data, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		return err
	}
	if _, err := toml.Decode(string(data), &config); err != nil {
		return err
	}
	return nil
}
