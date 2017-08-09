package main

import (
	"github.com/catpie/musdk-go"
	"github.com/orvice/utils/log"
	"time"
)

func InitWebApi() error {
	log.Info("init mu api")
	cfg := cfg.WebApi
	WebApiClient = musdk.NewClient(cfg.Url, cfg.Token, cfg.NodeId, musdk.TypeSs)
	return nil
}

func Boot() error {
	storage.ClearAll()
	go func() {
		for {
			CheckUsers()
			SubmitTraffic()
			time.Sleep(cfg.SyncTime * time.Second)
		}
	}()
	return nil
}
