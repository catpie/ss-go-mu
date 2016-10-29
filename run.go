package main

import (
	"fmt"
	"github.com/catpie/musdk-go"
	. "github.com/catpie/ss-go-mu/log"
)

func RunSs(user UserInterface) error {
	isExist, err := storage.Exists(user)
	if err != nil {
		return err
	}
	if isExist {
		Log.Debugf("user %d is running... skip", user.GetId())
		return nil
	}
	err = storage.StoreUser(user)
	if err != nil {
		return err
	}
	runWithCustomMethod(user)
	return nil
}

func StopSs(user UserInterface) error {
	return nil
}

func CheckUser(user UserInterface) error {
	go RunSs(user)
	return nil
}

func CheckUsers() error {
	Log.Info("check users...")
	users, err := WebApiClient.GetUsers()
	Log.Debug(users)
	if err != nil {
		// handle error
		Log.Error(err)
	}

	for _, user := range users {
		go CheckUser(user)
	}

	return nil
}

func SubmitTraffic() error {
	Log.Info("submit traffic....")
	users, err := WebApiClient.GetUsers()
	if err != nil {
		return err
	}
	var logs []musdk.UserTrafficLog
	for _, user := range users {
		size, err := storage.GetSize(user)
		if err != nil {
			Log.Error(fmt.Sprintf("get size fail for port:%d", user.GetPort()), err)
			continue
		}
		if size < 1024 {
			continue
		}
		log := musdk.UserTrafficLog{
			U:      0,
			D:      size,
			UserId: user.GetId(),
		}
		logs = append(logs, log)
		err = storage.SetSize(user, 0)
		if err != nil {
			Log.Error(fmt.Sprintf("set storage size to 0 fail for port:%d", user.GetPort()), err)
			continue
		}
	}
	err = WebApiClient.UpdateTraffic(logs)
	if err != nil {
		// @todo
		return err
	}
	return nil
}
