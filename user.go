package main

import (
	"github.com/catpie/musdk-go"
	"net"
	"sync"
)

var (
	usersLock = new(sync.Mutex)
	users     = map[int64]*User{}
)

type User struct {
	Id      int64
	lock    *sync.Mutex
	u, d    int64
	tcpConn net.Listener
	udpConn net.PacketConn
	apiUser musdk.User
}

func (u *User) AddU(t int64) {
	u.lock.Lock()
	u.u += t
	u.lock.Unlock()
}

func (u *User) AddD(t int64) {
	u.lock.Lock()
	u.d += t
	u.lock.Unlock()
}

func (u *User) Close() error {
	u.tcpConn.Close()
	u.udpConn.Close()
	return nil
}

func NewUser(u musdk.User) *User {
	return &User{
		Id:      u.Id,
		lock:    new(sync.Mutex),
		apiUser: u,
	}
}

func AddUser(id int64, u *User) {
	usersLock.Lock()
	defer usersLock.Unlock()

	users[id] = u
}
