package main

import (
	"fmt"
	"github.com/catpie/musdk-go"
	. "github.com/catpie/ss-go-mu/log"
	"strconv"
)

func RunSs(user UserInterface) error {
	runWithCustomMethod(user)
	users[user.GetId()] = user
	return nil
}

func StopSs(user UserInterface) error {
	passwdManager.del(strconv.Itoa(user.GetPort()))
	delete(users, user.GetId())
	return nil
}

func CheckUser(user UserInterface) error {
	Log.Info("check user: ", user)
	u, ok := users[user.GetId()]
	if !ok {
		return func() error {
			if user.IsEnable() {
				Log.Infof("run user %d", user.GetId())
				return RunSs(user)
			}
			return nil
		}()
	}
	if !u.IsEnable() {
		Log.Infof("disable user %d", u.GetId())
		return StopSs(user)
	}
	if user != u {
		Log.Infof("%d info is changed... restart ...", user.GetId())
		StopSs(user)
		return RunSs(user)
	}
	return nil
}

func CheckUsers() error {
	Log.Info("check users...")
	Log.Info("user in memery: ", users)
	us, err := WebApiClient.GetUsers()
	Log.Debug(us)
	if err != nil {
		// handle error
		Log.Error(err)
	}

	for _, u := range us {
		go CheckUser(u)
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
	if len(logs) == 0 {
		return nil
	}
	err = WebApiClient.UpdateTraffic(logs)
	if err != nil {
		// @todo
		return err
	}
	return nil
}
