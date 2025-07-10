package console

import (
	"fmt"
	"testing"
)

// usage bin -f funcName -param paramValue -param2 paramValue2
func TestRun(t *testing.T) {
	// example: go run . -f doPrint -bar bar1
	Register("doPrint", "console -f doPrint -bar $bar", func() {
		fmt.Println(OptStr("bar"))
	})
	Register("doPrint2", "console -f doPrint2 -bar $bar", func() {
		fmt.Println(OptStr("bar"))
	})
	Run()
}
