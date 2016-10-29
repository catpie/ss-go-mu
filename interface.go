package main

import (
	ss "github.com/orvice/shadowsocks-go/shadowsocks"
)

type LogTraffic func(id, u, d int64) error

type UserInterface interface {
	GetId() int64
	GetPort() int
	GetPasswd() string
	GetMethod() string
	IsEnable() bool
	GetCipher() (*ss.Cipher, error, bool)
	// UpdateTraffic(storageSize int) error
}
