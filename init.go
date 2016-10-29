package main

import (
	"github.com/BurntSushi/toml"
	"github.com/catpie/musdk-go"
	. "github.com/catpie/ss-go-mu/log"
	"io/ioutil"
	"time"
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
	go func() {
		for {
			CheckUsers()
			SubmitTraffic()
			time.Sleep(config.Base.SyncTime * time.Second)
		}
	}()
	return nil
}
