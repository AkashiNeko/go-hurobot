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

func CQRecord(text string) string {
	return fmt.Sprintf("[CQ:record,file=https://akashi.top/tts-file-cache/%s]", text)
}

func CQImage(url string) string {
	return fmt.Sprintf("[CQ:image,sub_type=0,url=%s]", url)
}

func CQRps() string {
	return "[CQ:rps]"
}

func CQDice() string {
	return "[CQ:dice]"
}
