package main

import (
	"github.com/BurntSushi/toml"
	"github.com/catpie/musdk-go"
	. "github.com/catpie/ss-go-mu/log"
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
	Log.Info(config)
	return nil
}

func InitWebApi() error {
	cfg := config.WebApi
	WebApiClient = musdk.NewClient(cfg.Url, cfg.Token, cfg.NodeId)
	return nil
}

func BootSs() error {
	users, err := WebApiClient.GetUsers()
	Log.Info(users)
	if err != nil {
		// handle error
		Log.Error(err)
	}

	for _, user := range users {
		Log.Info(user.Id)
		runWithCustomMethod(user)
	}

	return nil
}
