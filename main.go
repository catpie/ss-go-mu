package main

import (
	. "github.com/catpie/ss-go-mu/log"
)

func main() {
	Log.Info("Start")
	InitConfig()
	InitWebApi()
	BootSs()

	waitSignal()
}
