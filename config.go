package main

import (
	"github.com/orvice/utils/env"
	"time"
)

var (
	cfg = new(Config)
)

type Config struct {
	WebApi     WebApiCfg
	Base       BaseCfg
	SyncTime   time.Duration
	UDPTimeout time.Duration
}

type BaseCfg struct {
}

type WebApiCfg struct {
	Url    string
	Token  string
	NodeId int
}

func initCfg() {
	cfg.WebApi = WebApiCfg{
		Url:    env.Get("MU_URI"),
		Token:  env.Get("MU_TOKEN"),
		NodeId: env.GetInt("MU_NODE_ID"),
	}
	st := env.GetInt("SYNC_TIME", 60)
	udpTimeout := env.GetInt("UDP_TIMEOUT", 6)
	cfg.SyncTime = time.Second * time.Duration(st)
	cfg.UDPTimeout = time.Second * time.Duration(udpTimeout)
}
