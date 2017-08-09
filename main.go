package main

import (
	"github.com/orvice/utils/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// var err error
	log.Info("Start")
	initCfg()
	InitWebApi()

	waitSignal()
}

func waitSignal() {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	for sig := range sigChan {
		if sig == syscall.SIGHUP {
		} else {
			// is this going to happen?
			log.Infof("caught signal %v, exit", sig)
			os.Exit(0)
		}
	}
}
