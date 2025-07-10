package logs

import (
	"fmt"
	"testing"
)

func TestLog(t *testing.T) {
	InitServer("logV2Test", 0, "/data/logs/", true, func(typ string, tag string, msg string) {
		fmt.Printf("type:%s tag:%s msg:%s\n", typ, tag, msg)
	}, nil)
	Print("HELLO WORLD", "IM BUMBLE NEE\r\n")
	LogDebug("HELLO WORLD, IM %s", "BUMBLE BEE")
	PrintDebug("HELLO WORLD", "IM BUMBLE NEE")
	LogInfo("HELLO WORLD, IM %s", "BUMBLE BEE")
	LogError("HELLO WORLD, IM %s", "BUMBLE BEE")
	PrintErr("HELLO WORLD", "IM BUMBLE NEE")
	LogWarn("HELLO WORLD, IM %s", "BUMBLE BEE")
	PrintBill("adapterBill", "HELLO WORLD", "IM BUMBLE NEE")
	WriteBill("adapterBill2", "HELLO WORLD, IM %s", "BUMBLE BEE\r\n")

	CloseAllLogAdapter()
}
