package qbot

import "fmt"

func CQAt(userId uint64) string {
	return fmt.Sprintf("[CQ:at,qq=%d]", userId)
}

func CQReply(msgId uint64) string {
	return fmt.Sprintf("[CQ:reply,id=%d]", msgId)
}

func CQPoke(userId uint64) string {
	return fmt.Sprintf("[CQ:poke,qq=%d]", userId)
}
