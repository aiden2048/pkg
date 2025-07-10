package main

import (
	"fmt"

	"github.com/aiden2048/pkg/utils/console"
)

// usage bin -f funcName -param paramValue -param2 paramValue2
func main() {
	console.Register("doPrint", "console -f doPrint -bar $bar", func() {
		fmt.Println(console.OptStr("bar"))
	})
	console.Register("doPrint2", "console -f doPrint2 -bar $bar", func() {
		fmt.Println(console.OptStr("bar"))
	})
	console.Run()
}
