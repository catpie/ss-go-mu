package main

import (
	"fmt"
)

func genUserInfoKey(user UserInterface) string {
	return fmt.Sprintf("userinfo:%v", user.GetPort())
}

func genUserFlowKey(user UserInterface) string {
	return fmt.Sprintf("userflow:%v", user.GetPort())
}

func genUserOnlineKey(user UserInterface) string {
	return fmt.Sprintf("useronline:%v", user.GetPort())
}
