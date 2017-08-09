package main

import (
	"fmt"
	"github.com/orvice/utils/log"
)

func genUserInfoKey(user UserInterface) string {
	return fmt.Sprintf("userinfo:%v", user.GetId())
}

func genUserFlowKey(user UserInterface) string {
	return fmt.Sprintf("userflow:%v", user.GetId())
}

func genUserOnlineKey(user UserInterface) string {
	return fmt.Sprintf("useronline:%v", user.GetId())
}

func logf(format string, args ...interface{}) {
	log.Infof(format, args...)
}
