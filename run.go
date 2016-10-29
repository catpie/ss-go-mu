package main

import (
	. "github.com/catpie/ss-go-mu/log"
)

func RunSs(user UserInterface) error {
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
	Log.Info(users)
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
	return nil
}
