package main

import (
	"github.com/catpie/ss-go-mu/log"
)

const (
	QueueName = "ss-go-mu-queue"
)

func Pop() {
	err := Redis.GetClient().LPop(QueueName).Err()
	log.Log.Error(err)
}
