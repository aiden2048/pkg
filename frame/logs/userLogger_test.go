package logs

import (
	"fmt"
	"testing"
)

func TestUserLogger(t *testing.T) {
	InitServer("logV2Test", 0, "/data/logs/", true, func(typ string, tag string, msg string) {
		fmt.Printf("type:%s tag:%s msg:%s\n", typ, tag, msg)
	}, nil)
	LogUser(0, "HELLO OK")
	PrintUser(0, "HEE")
	LogInfoUser(0, "LogInfo")
	LogUserError(0, "Log")
	PrintUserError(0, "user err")
	PrintSess(nil, "PrintSess")
	PrintSessError(nil, "PrintSessErr")
	OnExit()
}
