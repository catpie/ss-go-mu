package main

import (
	"sync"
)

var (
// users = map[int]User{}
)

type User struct {
	lock sync.Mutex
	u, d int64
}

func (u User) AddU(t int64) {
	u.lock.Lock()
	u.u += t
	u.lock.Unlock()
}

func (u User) AddD(t int64) {
	u.lock.Lock()
	u.d += t
	u.lock.Unlock()
}
