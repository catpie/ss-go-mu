package main

import (
	"fmt"
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
