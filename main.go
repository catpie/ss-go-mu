package main

import (
	. "github.com/catpie/ss-go-mu/log"
)

func main() {
	var err error
	Log.Info("Start")
	err = InitConfig()
	if err != nil {
		panic(err)
	}
	InitWebApi()
	BootSs()

	waitSignal()
}
