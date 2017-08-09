package main

import (
	"fmt"
	"github.com/catpie/musdk-go"
	"github.com/orvice/utils/log"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"sync"
)

var (
	syncLock = new(sync.Mutex)
)

func CheckUsers() error {
	log.Info("check users...")
	log.Info("user in memery: ", users)
	us, err := WebApiClient.GetUsers()
	log.Debug(us)
	if err != nil {
		// handle error
		log.Error(err)
	}

	for _, u := range us {
		go checkUser(u)
	}

	return nil
}

func checkUser(u musdk.User) {

}

func runUser(user User, u musdk.User) {
	addr := fmt.Sprintf(":%d", u.Port)
	cipher := u.Method
	password := u.Passwd

	key := []byte(u.Passwd)

	var err error

	ciph, err := core.PickCipher(cipher, key, password)
	if err != nil {
		log.Error(err)
	}

	go udpRemote(user, addr, ciph.PacketConn)
	go tcpRemote(user, addr, ciph.StreamConn)
}

func SubmitTraffic() {
	syncLock.Lock()
	defer syncLock.Unlock()
}
